// /home/krylon/go/src/github.com/blicero/jazz/jes/jes.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-18 18:39:19 krylon>

// Package jes ("Job Entry System") accepts jobs feeds them into the queue.
package jes

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync/atomic"

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
	sock     *net.UnixConn
	JobQ     chan *model.Job
}

// Create creates (well, duh) and returns a new JES that watches the
// given directory.
func Create(path string) (*JES, error) {
	var (
		err  error
		addr net.UnixAddr
		j    = &JES{
			sockPath: path,
			JobQ:     make(chan *model.Job),
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
	// go j.mainloop()
	return nil
} // func (j *JES) Start() error

// func (j *JES) mainloop() {
// 	defer j.active.Store(false)
// 	defer j.log.Printf("[INFO] JES daemon @%s is quitting.\n",
// 		j.sockPath)

// 	var ticker = time.NewTicker(common.TickInterval)
// 	defer ticker.Stop()

// 	for j.active.Load() {
// 		select {
// 		case <-ticker.C:
// 			continue
// 			// case job := <-j.JobQ:
// 			// 	// Submit Job to Monitor!
// 		}
// 	}

// }
// func (j *JES) mainloop()

func (j *JES) socketLoop() {
	const (
		bufSize = 1 << 16
		errMax  = 10
	)
	var (
		trouble bool
		errCnt  int
	)

	defer func() {
		if x := recover(); x != nil {
			j.log.Printf("[ERROR] Panic: %s (trouble=%t)\n",
				x,
				trouble)
		}
	}()

	defer func() {
		var x error
		if x = j.sock.Close(); x != nil {
			j.log.Printf("[CRITICAL] Cannot close socket %s: %s\n",
				j.sockPath,
				x.Error())
		} else if x = os.Remove(j.sockPath); x != nil {
			j.log.Printf("[CRITICAL] Cannot remove socket %s: %s\n",
				j.sockPath,
				x.Error())
		}
	}()

	for j.active.Load() {
		// FIXME Maybe I should attempt to reuse rbuf instead of
		//       allocating a fresh one each time.
		var (
			err              error
			bytesRcvd        int
			bytesSent, flags int
			sender           *net.UnixAddr
			decBuf           *bytes.Buffer
			dec              *gob.Decoder
			job              *model.Job
			rbuf             = make([]byte, bufSize)
			response         string
		)
		trouble = false

		//if bytesRcvd, err = j.sock.Read(rbuf); err != nil {
		if bytesRcvd, _, flags, sender, err = j.sock.ReadMsgUnix(rbuf, nil); err != nil {

			j.log.Printf("[ERROR] Failed to read from Unix socket %s: %s\n",
				j.sockPath,
				err.Error())
			if errCnt++; errCnt >= maxErr {
				j.log.Printf("[ERROR] Maximum number of errors (%d) has occured, bailing out\n",
					errCnt)
				j.active.Store(false)
				trouble = true
				return
			}
			continue
		} else if bytesRcvd >= bufSize {
			// Buffer overflow-ish?
			j.log.Printf("[ERROR] Possible buffer overrun? Read %d bytes\n",
				bytesRcvd)
			if errCnt++; errCnt >= maxErr {
				j.log.Printf("[ERROR] Maximum number of errors (%d) has occured, bailing out\n",
					errCnt)
				j.active.Store(false)
				trouble = true
				return
			}
			trouble = true
		} else if flags != 0 {
			j.log.Printf("[DEBUG] ReadMsgUnix returned flags: %08x\n",
				flags)
		}

		j.log.Printf("[DEBUG] Received %d bytes from %s/%s\n",
			bytesRcvd,
			sender.Net,
			sender.Name)

		decBuf = bytes.NewBuffer(rbuf)
		dec = gob.NewDecoder(decBuf)
		job = new(model.Job)

		if err = dec.Decode(job); err != nil {
			response = fmt.Sprintf("Cannot decode Job: %s\n",
				err.Error())
			j.log.Printf("[ERROR] %s\n", response)
		} else {
			response = "OK"
		}

		j.JobQ <- job

		// FIXME I've never worked with Unix sockets before, will the
		//       response even reach the process/thread that wrote to the
		//       socket? Only one way to find out 🤷‍♂️
		if bytesSent, err = j.sock.Write([]byte(response)); err != nil {
			j.log.Printf("[ERROR] Failed to send response: %s\n",
				err.Error())
			if errCnt++; errCnt >= maxErr {
				j.log.Printf("[ERROR] Maximum number of errors (%d) has occured, bailing out\n",
					errCnt)
				j.active.Store(false)
				trouble = true
				return
			}
		} else if bytesSent != len(response) {
			j.log.Printf("[ERROR] Response is %d bytes long, but we sent %d\n",
				len(response),
				bytesSent)
		}
	}
} // func (j *JES) socketLoop()
