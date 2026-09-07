// /home/krylon/go/src/github.com/blicero/jazz/monitor/monitor.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-04 12:04:31 krylon>

// Package monitor implements the heart of the application, so to speak.
package monitor

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/database"
	"github.com/blicero/jazz/logdomain"
	"github.com/blicero/jazz/model"
	"github.com/blicero/jazz/monitor/command"
	"github.com/blicero/krylib"
)

// Monitor handles the Job queue, submitting job, executing them, cleaning
// up after them.
type Monitor struct {
	log    *log.Logger
	lock   sync.RWMutex
	db     *database.Database
	active atomic.Bool
	jobs   map[int64]*model.Job
	cmdQ   chan command.Command
}

// Create creates and returns a new Monitor.
func Create() (*Monitor, error) {
	var (
		err error
		mon = &Monitor{
			jobs: make(map[int64]*model.Job),
			cmdQ: make(chan command.Command),
		}
		jobs []*model.Job
	)

	if mon.log, err = common.GetLogger(logdomain.Monitor); err != nil {
		return nil, err
	} else if mon.db, err = database.Open(common.DbPath); err != nil {
		mon.log.Printf("[CRITICAL] Failed to open Database at %s: %s\n",
			common.DbPath, err.Error())
		return nil, err
	} else if jobs, err = mon.db.JobGetAll(); err != nil {
		mon.log.Printf("[CRITICAL] Failed to load all Jobs: %s\n",
			err.Error())
		return nil, err
	}

	for _, j := range jobs {
		mon.jobs[j.ID] = j
	}

	return mon, nil
} // func Create() (*Monitor, error)

// Active returns the Monitor's active flag.
func (mon *Monitor) Active() bool {
	return mon.active.Load()
} // func (mon *Monitor) Active() bool

// Stop clears the Monitor's active flag.
func (mon *Monitor) Stop() {
	mon.active.Store(false)
} // func (mon *Monitor) Stop()

// Run executes the Monitor's main loop.
func (mon *Monitor) Run() {
	if swapped := mon.active.CompareAndSwap(false, true); !swapped {
		mon.log.Println("[WARNING] Monitor appears to be running already.")
		return
	}

	defer mon.active.Store(false)
	defer mon.log.Println("[INFO] Monitor is shutting down.")

	var ticker = time.NewTicker(common.TickInterval)
	defer ticker.Stop()

	for mon.Active() {
		select {
		case <-ticker.C:
			// bla
		case cmd := <-mon.cmdQ:
			mon.handleCommand(cmd)
		}
	}
} // func (mon *Monitor) Run()

func (mon *Monitor) handleCommand(cmd command.Command) {
	var err error

	mon.log.Printf("[DEBUG] Monitor received command %s\n",
		cmd.Verb)
	defer mon.log.Printf("[DEBUG] Monitor is done handling %s\n",
		cmd.Verb)

	switch cmd.Verb {
	case command.Stop:
		mon.active.Store(false)
	case command.Submit:
		var (
			job *model.Job
			ok  bool
		)

		if job, ok = cmd.Object.(*model.Job); !ok {
			mon.log.Printf("[ERROR] Invalid Object for Verb %s: %T (%+v)\n",
				cmd.Verb,
				cmd.Object,
				cmd.Object)
			return
		} else if err = mon.db.JobAdd(job); err != nil {
			mon.log.Printf("[ERROR] Cannot add Job %s to database: %s\n",
				job.Name,
				err.Error())
		}

		mon.lock.Lock()
		mon.jobs[job.ID] = job
		mon.lock.Unlock()
	default:
		mon.log.Printf("[ERROR] Don't know how to handle command %s\n",
			cmd.Verb)
	}
} // func (mon *Monitor) handleCommand(cmd command.Command)

// nolint: unused
func (mon *Monitor) execute(job *model.Job) error {
	// var (
	// 	err error
	// )
	mon.log.Printf("[TRACE] Execute Job #%d (%s)\n",
		job.ID,
		job.Name)
	return krylib.ErrNotImplemented
} // func (mon *Monitor) execute(job *model.Job) error
