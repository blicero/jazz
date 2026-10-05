// /home/krylon/go/src/github.com/blicero/jazz/database/02_database_update_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 05. 10. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-10-05 11:06:52 krylon>

package database

import (
	"testing"
	"time"

	"github.com/blicero/jazz/model"
)

func TestJobSave(t *testing.T) {
	if tdb == nil || len(tJobs) != jCnt {
		t.SkipNow()
	}

	var (
		err  error
		jobs []*model.Job
	)

	if jobs, err = tdb.JobGetAll(); err != nil {
		t.Fatalf("Failed to get all Jobs from Database: %s",
			err.Error())
	}

	for _, j := range jobs {
		j.TimeStarted = time.Now().Add(time.Second * -240)
		j.TimeFinished = time.Now().Add(time.Second * -3)

		if err = tdb.JobSave(j); err != nil {
			t.Errorf("Failed to save Job: %s",
				err.Error())
		}
	}
} // func TestJobSave(t *testing.T)
