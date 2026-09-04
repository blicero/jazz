// /home/krylon/go/src/github.com/blicero/jazz/monitor/command/id.go
// -*- mode: go; coding: utf-8; -*-
// Created on 03. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-03 13:32:56 krylon>

//go:generate stringer -type=ID

// Package command provides symbolic constants to identifiy commands sent
// to the Monitor.
package command

// ID describes a command to the Monitor.
type ID uint8

const (
	Stop ID = iota
	Submit
	Cancel
	Freeze
	Modify
)

// Command is a command to the Monitor.
type Command struct {
	Verb   ID
	Object any
}
