// /home/krylon/go/src/github.com/blicero/jazz/model/model.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-12 14:00:54 krylon>

// Package model defines data types used throughout the application
package model

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/blicero/jazz/common"
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

// Ready returns true if the Job is ready to be exected right now.
func (j *Job) Ready() bool {
	var now = time.Now().Truncate(time.Second)
	return j.ScheduledStart.Before(now) ||
		j.ScheduledStart.Equal(now)
} // func (j *Job) Ready() bool

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

// PrettyPrint returns a string represantion of the Job intended for a human reader.
func (j *Job) PrettyPrint() string {
	var (
		sb strings.Builder
	)

	// ...
	fmt.Fprintf(&sb,
		`
Job %d (%s)
WorkDir:        %s
Niceness:       %d
IOPrio:         %d
ScheduledStart: %s
Deadline:       %s
`,
		j.ID,
		j.Name,
		j.WorkDir,
		j.Niceness,
		j.IOPrio,
		j.ScheduledStart.Format(common.TimestampFormat),
		j.Deadline.Format(common.TimestampFormat))

	if len(j.Env) > 0 {
		var (
			keys        = make([]string, len(j.Env))
			maxLen, idx int
			fmtString   string
		)

		for key := range j.Env {
			if len(key) > maxLen {
				maxLen = len(key)
			}
			keys[idx] = key
			idx++
		}

		sb.WriteString("Env:\n")
		slices.Sort(keys)

		fmtString = fmt.Sprintf("%%-%ds = %%s\n",
			maxLen)

		fmt.Printf("DBG Env fmt string = %s\n",
			fmtString)

		for _, key := range keys {
			fmt.Fprintf(
				&sb,
				fmtString,
				key,
				j.Env[key])
		}
	}

	sb.WriteString("Steps:\n")

	var (
		idxWidth  = int(math.Log10(float64(len(j.Steps)))) + 1
		fmtString = fmt.Sprintf("%%%dd: %%s\n", idxWidth)
	)

	for idx, step := range j.Steps {
		fmt.Fprintf(
			&sb,
			fmtString,
			idx,
			step)
	}

	sb.WriteString("\n")

	return sb.String()
} // func (j *Job) PrettyPrint() string
