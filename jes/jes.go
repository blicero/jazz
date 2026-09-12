// /home/krylon/go/src/github.com/blicero/jazz/jes/jes.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-12 16:06:46 krylon>

// Package jes ("Job Entry System") accepts jobs feeds them into the queue.
package jes

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/logdomain"
	"github.com/blicero/jazz/model"
)

// const readJCLDelay = time.Millisecond * 500
const maxErr = 10

// JES accepts Job definitions, processes them, and hands them to the Job queue.
type JES struct {
	log      *log.Logger
	sockPath string
	active   atomic.Bool
	lock     sync.RWMutex
	sock     *net.UnixConn
	jobQ     chan *model.Job
}

// Create creates (well, duh) and returns a new JES that watches the
// given directory.
func Create(path string) (*JES, error) {
	var (
		err  error
		addr net.UnixAddr
		j    = &JES{
			sockPath: path,
			jobQ:     make(chan *model.Job),
		}
	)

	addr = net.UnixAddr{
		Name: path,
		Net:  "unix",
	}

	if j.log, err = common.GetLogger(logdomain.JES); err != nil {
		return nil, err
	} else if j.sock, err = net.ListenUnixgram("unixgram", &addr); err != nil {
		j.log.Printf("[ERROR] Cannot listen at socket %s: %s\n",
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

	go j.socketLoop()
	go j.mainloop()
	return nil
} // func (j *JES) Start() error

func (j *JES) mainloop() {
	defer j.active.Store(false)
	defer j.log.Printf("[INFO] JES daemon @%s is quitting.\n",
		j.sockPath)

	var ticker = time.NewTicker(common.TickInterval)
	defer ticker.Stop()

	for j.active.Load() {
		select {
		case <-ticker.C:
			continue
		case job := <-j.jobQ:
			// Submit Job to Monitor!
		}
	}

} // func (j *JES) mainloop()

func (j *JES) socketLoop() {
	const bufSize = 1 << 16
	var trouble bool

	defer func() {
		if x := recover(); x != nil {
			j.log.Printf("[ERROR] Panic: %s (trouble=%t)\n",
				x,
				trouble)
		}
	}()

	for j.active.Load() {
		var (
			err               error
			bytesRcvd, errCnt int
			decBuf            *bytes.Buffer
			dec               *gob.Decoder
			job               *model.Job
			rbuf              = make([]byte, bufSize)
		)
		trouble = false

		if bytesRcvd, err = j.sock.Read(rbuf); err != nil {
			j.log.Printf("[ERROR] Failed to read from Unix socket %s: %s\n",
				j.sockPath,
				err.Error())
			if errCnt >= maxErr {
				j.log.Printf("[ERROR] Maximum number of errors (%d) has occured, bailing out\n",
					errCnt)
				j.active.Store(false)
				return
			}
			errCnt++
			continue
		} else if bytesRcvd >= bufSize {
			// Buffer overflow-ish?
			j.log.Printf("[ERROR] Possible buffer overrun? Read %d bytes\n",
				bytesRcvd)
			errCnt++
			trouble = true
		}

		decBuf = bytes.NewBuffer(rbuf)
		dec = gob.NewDecoder(decBuf)
		job = new(model.Job)

		if err = dec.Decode(job); err != nil {
			j.log.Printf("[ERROR] Cannot decode Job: %s\n",
				err.Error())
			errCnt++
			continue
		}

		j.jobQ <- job
	}
} // func (j *JES) socketLoop()
