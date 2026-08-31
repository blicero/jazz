// /home/krylon/go/src/github.com/blicero/jazz/logdomain/logdomain.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-31 10:07:57 krylon>

//go:generate stringer -type=ID

// Package logdomain defines symbolic constants to identify the parts of
// the application that want to write to the log.
package logdomain

// ID identifies a subsystem.
type ID uint8

const (
	Job ID = iota
	Queue
	Server
)

// All returns all defined ID values.
func All() []ID {
	return []ID{
		Job,
		Queue,
		Server,
	}
} // func All() []ID
