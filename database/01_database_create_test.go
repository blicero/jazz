// /home/krylon/go/src/github.com/blicero/jazz/database/01_database_create_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-02 12:04:55 krylon>

package database

import (
	"fmt"
	"math/rand/v2"
	"os"
	"testing"
	"time"

	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/model"
	"github.com/blicero/jazz/model/predicate"
)

const (
	jCnt   = 256
	cmdCnt = 32
)

var (
	tdb   *Database
	tJobs []*model.Job
	tCmd  []string
)

func init() {
	var (
		err   error
		dirh  *os.File
		names []string
		idx   []int
	)

	if dirh, err = os.Open("/usr/bin"); err != nil {
		panic(fmt.Errorf("cannot open /usr/bin: %w", err))
	}

	defer dirh.Close() // nolint: errcheck

	if names, err = dirh.Readdirnames(0); err != nil {
		panic(fmt.Errorf("cannot read file names from /usr/bin: %w", err))
	}

	tCmd = make([]string, cmdCnt)
	idx = rand.Perm(len(names))

	for i := range cmdCnt {
		tCmd[i] = names[idx[i]]
	}
}

// create a random Job with bogus values for testing purposes.
func rndJob() *model.Job {
	var j = &model.Job{
		ID: rand.Int64(),
		Name: fmt.Sprintf("Random Job No. %04d",
			rand.Int()),
		WorkDir:  "/tmp",
		Niceness: 5,
		ScheduledStart: time.Now().Add(
			time.Minute *
				time.Duration(rand.Int64N(1440))),
	}

	var stepCnt = rand.IntN(3) + 1

	j.Steps = make([]model.Step, stepCnt)

	for i := range stepCnt {
		j.Steps[i] = model.Step{
			Seq:       int64(i),
			Predicate: predicate.Always,
			Command:   tCmd[rand.IntN(cmdCnt)],
			Stdout: fmt.Sprintf("/jazz/spool/%08x.%04x.out",
				j.ID,
				i),
			Stderr: fmt.Sprintf("/jazz/spool/%08x.%04x.err",
				j.ID,
				i),
		}
	}

	return j
}

func TestOpen(t *testing.T) {
	var err error

	if tdb, err = Open(common.DbPath); err != nil {
		tdb = nil
		t.Fatalf("Failed to open Database at %s: %s",
			common.DbPath,
			err.Error())
	}
} // func TestOpen(t *testing.T)

func TestJobAdd(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	var err error

	tJobs = make([]*model.Job, 0, jCnt)

	for i := range jCnt {
		var j = rndJob()

		if err = tdb.JobAdd(j); err != nil {
			t.Fatalf("Failed to add Job %d: %s",
				i+1,
				err.Error())
		}

		tJobs = append(tJobs, j)
	}
} // func TestJobAdd(t *testing.T)
