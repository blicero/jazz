// /home/krylon/go/src/github.com/blicero/jazz/jes/jes.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-19 12:15:45 krylon>

// Package web handles job submissions and provides a web interface to the
// Monitor.
package web

import (
	"context"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/logdomain"
	"github.com/blicero/jazz/model"
	"github.com/blicero/jazz/monitor"
	"github.com/gorilla/mux"
)

// Web provides a web interface for the Monitor and a web service for the CLI
// client to submit and manage Jobs.
type Web struct {
	mon       *monitor.Monitor
	addr      string
	active    atomic.Bool
	log       *log.Logger
	router    *mux.Router
	srv       http.Server
	mimeTypes map[string]string
}

// Create creates and returns a new JES instance.
func Create(addr string, mon *monitor.Monitor) (*Web, error) {
	var (
		err error
		j   = &Web{
			mon:  mon,
			addr: addr,
			mimeTypes: map[string]string{
				".css":  "text/css",
				".map":  "application/json",
				".js":   "text/javascript",
				".png":  "image/png",
				".jpg":  "image/jpeg",
				".jpeg": "image/jpeg",
				".webp": "image/webp",
				".gif":  "image/gif",
				".json": "application/json",
				".html": "text/html",
			},
			router: mux.NewRouter(),
		}
	)

	if j.log, err = common.GetLogger(logdomain.JES); err != nil {
		return nil, err
	}

	j.srv.Addr = addr
	j.srv.ErrorLog = j.log
	j.srv.Handler = j.router

	// ...

	return j, nil
} // func Create(addr string, mon *monitor.Monitor) (*JES, error)

// IsActive returns the value of the JES' active flag.
func (j *Web) IsActive() bool {
	return j.active.Load()
} // func (j *JES) IsActive() bool

// Stop tells the JES to stop.
func (j *Web) Stop() {
	j.active.Store(false)
	j.srv.Shutdown(context.Background())
} // func (j *JES) Stop()

// Run executes the JES server's main loop.
func (j *Web) Run() {
	var (
		err     error
		swapped bool
	)

	if swapped = j.active.CompareAndSwap(false, true); !swapped {
		j.log.Printf("[INFO] JES appears to be running already. Toodles!\n")
		return
	}

	defer j.log.Printf("[INFO] JES is shutting down.\n")

	// I have initially copied this from some tutorial or documentation, but
	// I am not sure if it is really necessary. OTOH, it does not appear to
	// do any harm.
	http.Handle("/", j.router)

	if err = j.srv.ListenAndServe(); err != nil {
		j.log.Printf("[ERROR] The web server ran into an error: %s\n",
			err.Error())
	}
} // func (j *JES) Run()

//////////////////////////////////////////////////////////////////////////////
/// Handle requests //////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////////////
/// Web service //////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

func (j *Web) handleSubmit(w http.ResponseWriter, r *http.Request) {
	j.log.Printf("[TRACE] Handle %s from %s\n",
		r.URL,
		r.RemoteAddr)
	var (
		err error
		job *model.Job
	)
}
