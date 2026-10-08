// /home/krylon/go/src/github.com/blicero/jazz/monitor/monitor.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-10-05 11:10:29 krylon>

// Package monitor implements the heart of the application, so to speak.
package monitor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/database"
	"github.com/blicero/jazz/logdomain"
	"github.com/blicero/jazz/model"
	"github.com/blicero/jazz/monitor/command"
	"github.com/google/shlex"
)

// Monitor handles the Job queue, submitting job, executing them, cleaning
// up after them.
type Monitor struct {
	log    *log.Logger
	lock   sync.RWMutex
	db     *database.Database
	active atomic.Bool
	jobs   map[int64]*model.Job
	CmdQ   chan command.Command
}

// Create creates and returns a new Monitor.
func Create() (*Monitor, error) {
	var (
		err error
		mon = &Monitor{
			jobs: make(map[int64]*model.Job),
			CmdQ: make(chan command.Command),
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

// Start starts the Monitor's main loop in a separate goroutine.
func (mon *Monitor) Start() {
	go mon.mainLoop()
}

// mainLoop executes the Monitor's main loop.
func (mon *Monitor) mainLoop() {
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
		case cmd := <-mon.CmdQ:
			go mon.handleCommand(cmd)
		}
	}
} // func (mon *Monitor) mainLoop()

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
		} else if err = mon.SubmitJob(job); err != nil {
			mon.log.Printf("[ERROR] Failed to submit Job: %s\n",
				err.Error())
		}
	default:
		mon.log.Printf("[ERROR] Don't know how to handle command %s\n",
			cmd.Verb)
	}
} // func (mon *Monitor) handleCommand(cmd command.Command)

// SubmitJob places a new Job in the Job queue and saves it to the Database.
func (mon *Monitor) SubmitJob(job *model.Job) error {
	var err error

	if err = mon.db.JobAdd(job); err != nil {
		mon.log.Printf("[ERROR] Cannot add Job %s to database: %s\n",
			job.Name,
			err.Error())
		return err
	}

	mon.lock.Lock()
	mon.jobs[job.ID] = job
	mon.lock.Unlock()

	return nil
} // func (mon *Monitor) SubmitJob(job *model.Job) error

// GetQueuedJobs returns a slice of the currently enqueued Jobs.
func (mon *Monitor) GetQueuedJobs() ([]*model.Job, error) {
	var (
		err  error
		jobs []*model.Job
	)

	if jobs, err = mon.db.JobGetAll(); err != nil {
		mon.log.Printf("[ERROR] Failed to query Job queue: %s\n",
			err.Error())
		return nil, err
	}

	return jobs, nil
} // func (mon *Monitor) GetQueuedJobs() ([]*model.Job, error)

// nolint: unused
func (mon *Monitor) executeJob(job *model.Job) error {
	var (
		err       error
		spoolPath [][2]string
	)

	mon.log.Printf("[TRACE] Execute Job #%d (%s)\n",
		job.ID,
		job.Name)

	spoolPath = make([][2]string, len(job.Steps))

	job.TimeStarted = time.Now()

	if err = mon.db.JobSave(job); err != nil {
		mon.log.Printf("[ERROR] Failed to save Job status: %s\n",
			err.Error())
	}

	for i := range len(job.Steps) {
		spoolPath[i][0] = filepath.Join(
			common.SpoolDir,
			fmt.Sprintf("%08x.%02d.out",
				job.ID,
				i+1))

		if err = mon.executeStep(job, i, spoolPath[i]); err != nil {
			mon.log.Printf("[ERROR] Job %d Step %d failed: %s\n",
				job.ID, i,
				err.Error())
		} else if err = mon.db.JobSave(job); err != nil {
			mon.log.Printf("[ERROR] Failed to save Job status: %s\n",
				err.Error())
		}
	}

	job.TimeFinished = time.Now()

	return nil
} // func (mon *Monitor) executeJob(job *model.Job) error

func (mon *Monitor) executeStep(job *model.Job, idx int, out [2]string) error {
	var (
		err            error
		stdout, stderr *os.File
		ctx            context.Context
		cancel         context.CancelFunc
		proc           *exec.Cmd
		pieces         []string
		command        = make([]string, 0, 4)
	)

	if job.Niceness > 0 {
		command = append(
			command,
			"nice",
			"-n",
			strconv.FormatInt(job.Niceness, 10))
	}

	if runtime.GOOS == "linux" && (job.IOPrio == 1 || job.IOPrio == 3) {
		command = append(
			command,
			"ionice",
			"-c",
			strconv.FormatInt(job.IOPrio, 10))
	}

	if pieces, err = shlex.Split(job.Steps[idx].Command); err != nil {
		mon.log.Printf("[ERROR] Cannot tokenize command line %q: %s\n",
			job.Steps[idx].Command,
			err.Error())
		return err
	}

	command = append(command, pieces...)
	ctx = context.Background()

	if !job.Deadline.IsZero() {
		ctx, cancel = context.WithDeadline(ctx, job.Deadline)
		defer cancel()
	}

	proc = exec.CommandContext(
		ctx,
		command[0],
		command[1:]...)

	proc.Env = job.Environment()
	proc.Dir = job.WorkDir

	if stdout, err = os.Create(out[0]); err != nil {
		mon.log.Printf("[ERROR] Cannot open stdout %s: %s\n",
			out[0],
			err.Error())
		return err
	}

	defer stdout.Close() // nolint: errcheck

	if stderr, err = os.Create(out[1]); err != nil {
		mon.log.Printf("[ERROR] Cannot open stderr %s: %s\n",
			out[1],
			err.Error())
		return err
	}

	defer stderr.Close() // nolint: errcheck

	proc.Stdout = stdout
	proc.Stderr = stderr
	job.Steps[idx].Stdout = out[0]
	job.Steps[idx].Stderr = out[1]

	var ticker = time.NewTicker(common.Timeout)
	defer ticker.Stop()

	if err = proc.Start(); err != nil {
		mon.log.Printf("[ERROR] Failed to start Job %d step %d: %s\n",
			job.ID, idx, err.Error())
		return err
	}

	job.Steps[idx].TimeStarted = time.Now()

	if err = proc.Wait(); err != nil {
		mon.log.Printf("[ERROR] Error executing Job %d Step %d: %s\n",
			job.ID,
			idx,
			err.Error())
		return err
	}

	job.Steps[idx].TimeFinished = time.Now()
	job.Steps[idx].Status = proc.ProcessState.ExitCode()

	return nil
} // func (mon *Monitor) executeStep(job *model.Job, idx int, out [2]string) error
