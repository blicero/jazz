// /home/krylon/go/src/github.com/blicero/jazz/model/model.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-02 13:30:08 krylon>

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

// Equal returns true if the argument given is another Step value with the same
// values field by field.
func (s *Step) Equal(other any) bool {
	switch s2 := other.(type) {
	case Step:
		return s.Seq == s2.Seq &&
			s.Predicate == s2.Predicate &&
			s.Command == s2.Command &&
			s.TimeStarted.Equal(s2.TimeStarted) &&
			s.TimeFinished.Equal(s2.TimeFinished) &&
			s.Status == s2.Status &&
			s.Stdout == s2.Stdout &&
			s.Stderr == s2.Stderr
	case *Step:
		return s.Seq == s2.Seq &&
			s.Predicate == s2.Predicate &&
			s.Command == s2.Command &&
			s.TimeStarted.Equal(s2.TimeStarted) &&
			s.TimeFinished.Equal(s2.TimeFinished) &&
			s.Status == s2.Status &&
			s.Stdout == s2.Stdout &&
			s.Stderr == s2.Stderr
	default:
		return false
	}
} // func (s *Step) Equal(other any) bool

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

// Equal returns true if the argument is another Job instance and all fields
// have equal values.
func (j *Job) Equal(other any) bool {
	var ok bool
	switch j2 := other.(type) {
	case Job:
		ok = j.ID == j2.ID &&
			j.Name == j2.Name &&
			j.WorkDir == j2.WorkDir &&
			j.Niceness == j2.Niceness &&
			j.IOPrio == j2.IOPrio &&
			j.ScheduledStart.Equal(j2.ScheduledStart) &&
			j.Deadline.Equal(j2.Deadline) &&
			j.TimeStarted.Equal(j2.TimeStarted) &&
			j.TimeFinished.Equal(j2.TimeFinished) &&
			j.CurStep == j2.CurStep
		if !ok ||
			len(j.Env) != len(j2.Env) ||
			len(j.Steps) != len(j2.Steps) {
			return false
		}

		for k, v := range j.Env {
			var (
				v2 string
				x  bool
			)

			if v2, x = j2.Env[k]; !x || v != v2 {
				return false
			}
		}

		for i, s := range j.Steps {
			if !s.Equal(j2.Steps[i]) {
				return false
			}
		}

		return true
	case *Job:
		ok = j.ID == j2.ID &&
			j.Name == j2.Name &&
			j.WorkDir == j2.WorkDir &&
			j.Niceness == j2.Niceness &&
			j.IOPrio == j2.IOPrio &&
			j.ScheduledStart.Equal(j2.ScheduledStart) &&
			j.Deadline.Equal(j2.Deadline) &&
			j.TimeStarted.Equal(j2.TimeStarted) &&
			j.TimeFinished.Equal(j2.TimeFinished) &&
			j.CurStep == j2.CurStep
		if !ok ||
			len(j.Env) != len(j2.Env) ||
			len(j.Steps) != len(j2.Steps) {
			return false
		}

		for k, v := range j.Env {
			var (
				v2 string
				x  bool
			)

			if v2, x = j2.Env[k]; !x || v != v2 {
				return false
			}
		}

		for i, s := range j.Steps {
			if !s.Equal(j2.Steps[i]) {
				return false
			}
		}

		return true
	default:
		return false
	}
} // func (j *Job) Equal(other any) bool
