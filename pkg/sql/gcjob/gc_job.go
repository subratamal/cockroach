// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package gcjob

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var (
	// MaxSQLGCInterval is the longest the polling interval between checking if
	// elements should be GC'd.
	MaxSQLGCInterval = 5 * time.Minute
)

// SetSmallMaxGCIntervalForTest sets the MaxSQLGCInterval and then returns a closure
// that resets it.
// This is to be used in tests like:
//    defer SetSmallMaxGCIntervalForTest()
func SetSmallMaxGCIntervalForTest() func() {
	oldInterval := MaxSQLGCInterval
	MaxSQLGCInterval = 500 * time.Millisecond
	return func() {
		MaxSQLGCInterval = oldInterval
	}
}

type schemaChangeGCResumer struct {
	jobID int64
}

// performGC GCs any schema elements that are in the DELETING state and returns
// a bool indicating if it GC'd any elements.
func performGC(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	details *jobspb.SchemaChangeGCDetails,
	progress *jobspb.SchemaChangeGCProgress,
) error {
	if details.Indexes != nil {
		return errors.Wrap(gcIndexes(ctx, execCfg, details.ParentID, progress), "attempting to GC indexes")
	} else if details.Tables != nil {
		if err := gcTables(ctx, execCfg, progress); err != nil {
			return errors.Wrap(err, "attempting to GC tables")
		}

		// Drop database zone config when all the tables have been GCed.
		if details.ParentID != descpb.InvalidID && isDoneGC(progress) {
			if err := deleteDatabaseZoneConfig(ctx, execCfg.DB, execCfg.Codec, details.ParentID); err != nil {
				return errors.Wrap(err, "deleting database zone config")
			}
		}
	}
	return nil
}

// Resume is part of the jobs.Resumer interface.
func (r schemaChangeGCResumer) Resume(
	ctx context.Context, phs interface{}, _ chan<- tree.Datums,
) error {
	p := phs.(sql.PlanHookState)
	// TODO(pbardea): Wait for no versions.
	execCfg := p.ExecCfg()
	if fn := execCfg.GCJobTestingKnobs.RunBeforeResume; fn != nil {
		if err := fn(r.jobID); err != nil {
			return err
		}
	}
	details, progress, err := initDetailsAndProgress(ctx, execCfg, r.jobID)
	if err != nil {
		return err
	}

	// If there are any interleaved indexes to drop as part of a table TRUNCATE
	// operation, then drop the indexes before waiting on the GC timer.
	if len(details.InterleavedIndexes) > 0 {
		// Before deleting any indexes, ensure that old versions of the table
		// descriptor are no longer in use.
		if err := sql.WaitToUpdateLeases(ctx, execCfg.LeaseManager, details.InterleavedTable.ID); err != nil {
			return err
		}
		if err := sql.TruncateInterleavedIndexes(
			ctx,
			execCfg,
			tabledesc.NewImmutable(*details.InterleavedTable),
			details.InterleavedIndexes,
		); err != nil {
			return err
		}
	}

	tableDropTimes, indexDropTimes := getDropTimes(details)

	timer := timeutil.NewTimer()
	defer timer.Stop()
	timer.Reset(0)
	gossipUpdateC, cleanup := execCfg.GCJobNotifier.AddNotifyee(ctx)
	defer cleanup()
	for {
		select {
		case <-gossipUpdateC:
			if log.V(2) {
				log.Info(ctx, "received a new system config")
			}
		case <-timer.C:
			timer.Read = true
			if log.V(2) {
				log.Info(ctx, "SchemaChangeGC timer triggered")
			}
		case <-ctx.Done():
			return ctx.Err()
		}

		// Refresh the status of all tables in case any GC TTLs have changed.
		remainingTables := getAllTablesWaitingForGC(details, progress)
		expired, earliestDeadline := refreshTables(ctx, execCfg, remainingTables, tableDropTimes, indexDropTimes, r.jobID, progress)
		timerDuration := time.Until(earliestDeadline)

		if expired {
			// Some elements have been marked as DELETING so save the progress.
			persistProgress(ctx, execCfg, r.jobID, progress)
			if err := performGC(ctx, execCfg, details, progress); err != nil {
				return err
			}
			persistProgress(ctx, execCfg, r.jobID, progress)

			// Trigger immediate re-run in case of more expired elements.
			timerDuration = 0
		}

		if isDoneGC(progress) {
			return nil
		}

		// Schedule the next check for GC.
		if timerDuration > MaxSQLGCInterval {
			timerDuration = MaxSQLGCInterval
		}
		timer.Reset(timerDuration)
	}
}

// OnFailOrCancel is part of the jobs.Resumer interface.
func (r schemaChangeGCResumer) OnFailOrCancel(context.Context, interface{}) error {
	return nil
}

func init() {
	createResumerFn := func(job *jobs.Job, settings *cluster.Settings) jobs.Resumer {
		return &schemaChangeGCResumer{
			jobID: *job.ID(),
		}
	}
	jobs.RegisterConstructor(jobspb.TypeSchemaChangeGC, createResumerFn)
}
