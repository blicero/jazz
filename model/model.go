// /home/krylon/go/src/github.com/blicero/jazz/model/model.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-31 10:39:19 krylon>

// Package model defines data types used throughout the application
package model

import (
	"time"

	"github.com/blicero/jazz/model/predicate"
)

// Step defines a command to be executed as part of a Job.
type Step struct {
	Seq          int64 // ???
	Predicate    predicate.ID
	Command      string
	TimeStarted  time.Time
	TimeFinished time.Time
	Status       int
}

// Job is a sequence of commands to be exected.
type Job struct {
	ID             int64
	Name           string
	ScheduledStart time.Time
	TimeStarted    time.Time
	TimeFinished   time.Time
	Env            map[string]string
	Steps          []Step
}
