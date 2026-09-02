// /home/krylon/go/src/github.com/blicero/jazz/model/model.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-02 11:10:23 krylon>

// Package model defines data types used throughout the application
package model

import (
	"fmt"
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
	Stdout       string
	Stderr       string
}

// Job is a sequence of commands to be exected.
type Job struct {
	ID             int64
	Name           string
	WorkDir        string
	Niceness       int64
	IOPrio         int64
	ScheduledStart time.Time
	Deadline       time.Time
	TimeStarted    time.Time
	TimeFinished   time.Time
	Env            map[string]string
	Steps          []Step
	CurStep        int64
}

// KeyStr returns a that contains the Job's ID, for use as a database key.
func (j *Job) KeyStr() []byte {
	return fmt.Appendf(nil, "%06d", j.ID)
}
