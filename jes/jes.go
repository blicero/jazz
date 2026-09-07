// /home/krylon/go/src/github.com/blicero/jazz/jes/jes.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-07 11:22:46 krylon>

// Package jes ("Job Entry System") accepts jobs feeds them into the queue.
package jes

import (
	"errors"
	"log"
	"os"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/logdomain"
	"github.com/blicero/krylib"
	"github.com/fsnotify/fsnotify"
)

const readJCLDelay = time.Millisecond * 500

var (
	jclPat  = regexp.MustCompile("[.](?i:jcl|lua|job)$")
	sheBang = regexp.MustCompile(`(?m)\A#!.*$`)
)

// JES accepts Job definitions, processes them, and hands them to the Job queue.
type JES struct {
	log     *log.Logger
	workDir string
	active  atomic.Bool
	lock    sync.RWMutex
	watch   *fsnotify.Watcher
	flist   map[string]time.Time
}

// Create creates (well, duh) and returns a new JES that watches the
// given directory.
func Create(path string) (*JES, error) {
	var (
		err error
		j   = &JES{
			workDir: path,
			flist:   make(map[string]time.Time),
		}
	)

	if j.log, err = common.GetLogger(logdomain.JES); err != nil {
		return nil, err
	} else if j.watch, err = fsnotify.NewWatcher(); err != nil {
		j.log.Printf("[CRITICAL] Cannot create FSNotify Watcher: %s\n",
			err.Error())
		return nil, err
	} else if err = os.MkdirAll(path, 0755); err != nil {
		j.log.Printf("[ERROR] Cannot create directory %s: %s\n",
			path,
			err.Error())
		return nil, err
	} else if err = j.watch.Add(path); err != nil {
		j.log.Printf("[ERROR] Failed to watch for events on %s: %s\n",
			path,
			err.Error())
		return nil, err
	}

	return j, nil
} // func Create(path string) (*JES, error)

// Active returns the JES' active flag.
func (j *JES) Active() bool {
	return j.active.Load()
} // func (j *JES) Active() bool

// Start initiates the JES' main loop.
func (j *JES) Start() error {
	if swapped := j.active.CompareAndSwap(false, true); !swapped {
		j.log.Printf("[WARNING] JES appears to be running already.\n")
		return errors.New("jes appears to be running already")
	}

	go j.mainloop()
	return nil
} // func (j *JES) Start() error

func (j *JES) mainloop() {
	defer j.active.Store(false)
	defer j.log.Printf("[INFO] JES watcher on %s is quitting.\n",
		j.workDir)

	var ticker = time.NewTicker(common.TickInterval)
	defer ticker.Stop()

	var ckTicker = time.NewTicker(time.Millisecond * 2500)
	defer ckTicker.Stop()

	for j.active.Load() {
		select {
		case <-ticker.C:
			continue
		case <-ckTicker.C:
			j.checkJCL()
		case ev := <-j.watch.Events:
			// so, what are you going to do about it?
			j.log.Printf("[TRACE] Received Event %s on %s\n",
				ev.Op,
				ev.Name)
			j.handleEvent(ev)
		}
	}
} // func (j *JES) mainloop()

func (j *JES) handleEvent(ev fsnotify.Event) {
	j.lock.Lock()
	defer j.lock.Unlock()
	switch ev.Op {
	case fsnotify.Create:
		j.log.Printf("[TRACE] New file %s\n",
			ev.Name)

		if jclPat.MatchString(ev.Name) {
			j.flist[ev.Name] = time.Now()
		}
	case fsnotify.Write:
		j.log.Printf("[TRACE] File %s was written to\n",
			ev.Name)
		if _, ok := j.flist[ev.Name]; ok {
			j.flist[ev.Name] = time.Now()
		}
	case fsnotify.Remove:
		j.log.Printf("[TRACE] %s was deleted.\n",
			ev.Name)
		delete(j.flist, ev.Name)
	default:
		j.log.Printf("[TRACE] Ignore %s being %s-ed\n",
			ev.Name,
			ev.Op)
	}
} // func (j *JES) handleEvent(ev fsnotify.Event)

func (j *JES) checkJCL() {
	j.lock.Lock()
	defer j.lock.Unlock()

	if len(j.flist) == 0 {
		return
	}

	for fname, timestamp := range j.flist {
		if time.Since(timestamp) >= readJCLDelay {
			go j.processJob(fname) // nolint: errcheck
			delete(j.flist, fname)
		}
	}
} // func (j *JES) checkJCL()

func (j *JES) processJob(path string) error {
	return krylib.ErrNotImplemented
} // func (j *JES) processJob(path string)
