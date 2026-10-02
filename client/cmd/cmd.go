// /home/krylon/go/src/github.com/blicero/jazz/client/cmd/cmd.go
// -*- mode: go; coding: utf-8; -*-
// Created on 02. 10. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-10-02 10:22:30 krylon>

// Package cmd provides symbolic constants to represent operations on the
// Job Queue web service.
package cmd

//go:generate stringer -type=Cmd

// Cmd represents a command/query to be sent to the web service.
type Cmd uint8

const (
	Nothing Cmd = iota
	Submit
	QueryQueue
)
