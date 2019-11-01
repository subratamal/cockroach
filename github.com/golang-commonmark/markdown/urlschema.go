// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

var urlSchemas = []string{
	"aaa",
	"aaas",
	"about",
	"acap",
	"adiumxtra",
	"afp",
	"afs",
	"aim",
	"apt",
	"attachment",
	"aw",
	"beshare",
	"bitcoin",
	"bolo",
	"callto",
	"cap",
	"chrome",
	"chrome-extension",
	"cid",
	"coap",
	"com-eventbrite-attendee",
	"content",
	"crid",
	"cvs",
	"data",
	"dav",
	"dict",
	"dlna-playcontainer",
	"dlna-playsingle",
	"dns",
	"doi",
	"dtn",
	"dvb",
	"ed2k",
	"facetime",
	"feed",
	"file",
	"finger",
	"fish",
	"ftp",
	"geo",
	"gg",
	"git",
	"gizmoproject",
	"go",
	"gopher",
	"gtalk",
	"h323",
	"hcp",
	"http",
	"https",
	"iax",
	"icap",
	"icon",
	"im",
	"imap",
	"info",
	"ipn",
	"ipp",
	"irc",
	"irc6",
	"ircs",
	"iris",
	"iris.beep",
	"iris.lwz",
	"iris.xpc",
	"iris.xpcs",
	"itms",
	"jar",
	"javascript",
	"jms",
	"keyparc",
	"lastfm",
	"ldap",
	"ldaps",
	"magnet",
	"mailto",
	"maps",
	"market",
	"message",
	"mid",
	"mms",
	"ms-help",
	"msnim",
	"msrp",
	"msrps",
	"mtqp",
	"mumble",
	"mupdate",
	"mvn",
	"news",
	"nfs",
	"ni",
	"nih",
	"nntp",
	"notes",
	"oid",
	"opaquelocktoken",
	"palm",
	"paparazzi",
	"platform",
	"pop",
	"pres",
	"proxy",
	"psyc",
	"query",
	"res",
	"resource",
	"rmi",
	"rsync",
	"rtmp",
	"rtsp",
	"secondlife",
	"service",
	"session",
	"sftp",
	"sgn",
	"shttp",
	"sieve",
	"sip",
	"sips",
	"skype",
	"smb",
	"sms",
	"snmp",
	"soap.beep",
	"soap.beeps",
	"soldat",
	"spotify",
	"ssh",
	"steam",
	"svn",
	"tag",
	"teamspeak",
	"tel",
	"telnet",
	"tftp",
	"things",
	"thismessage",
	"tip",
	"tn3270",
	"tv",
	"udp",
	"unreal",
	"urn",
	"ut2004",
	"vemmi",
	"ventrilo",
	"view-source",
	"webcal",
	"ws",
	"wss",
	"wtai",
	"wyciwyg",
	"xcon",
	"xcon-userid",
	"xfire",
	"xmlrpc.beep",
	"xmlrpc.beeps",
	"xmpp",
	"xri",
	"ymsgr",
	"z39.50r",
	"z39.50s",
}

var urlSchemasSet = make(map[string]bool)

func init() {
	for _, s := range urlSchemas {
		urlSchemasSet[s] = true
	}
}

func matchSchema(s string) bool {
	_, ok := urlSchemasSet[s]
	return ok
}