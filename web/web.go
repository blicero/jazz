// /home/krylon/go/src/github.com/blicero/jazz/jes/jes.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-19 13:07:58 krylon>

// Package web handles job submissions and provides a web interface to the
// Monitor.
package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/logdomain"
	"github.com/blicero/jazz/model"
	"github.com/blicero/jazz/monitor"
	"github.com/blicero/jazz/monitor/command"
	"github.com/gorilla/mux"
)

const (
	cacheControl = "max-age=120, public"
	noCache      = "no-store, max-age=0"
	tmplFolder   = "assets/templates"
)

func cacheSeconds(seconds int) string {
	if seconds == 0 {
		return noCache
	}

	return fmt.Sprintf("max-age=%d, public",
		seconds)
} // func cacheSeconds(second int) string

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
		srv = &Web{
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

	if srv.log, err = common.GetLogger(logdomain.JES); err != nil {
		return nil, err
	}

	srv.srv.Addr = addr
	srv.srv.ErrorLog = srv.log
	srv.srv.Handler = srv.router

	// ...

	return srv, nil
} // func Create(addr string, mon *monitor.Monitor) (*Web, error)

// IsActive returns the value of the JES' active flag.
func (srv *Web) IsActive() bool {
	return srv.active.Load()
} // func (j *JES) IsActive() bool

// Stop tells the JES to stop.
func (srv *Web) Stop() {
	srv.active.Store(false)
	srv.srv.Shutdown(context.Background())
} // func (j *JES) Stop()

// Run executes the JES server's main loop.
func (srv *Web) Run() {
	var (
		err     error
		swapped bool
	)

	if swapped = srv.active.CompareAndSwap(false, true); !swapped {
		srv.log.Printf("[INFO] JES appears to be running already. Toodles!\n")
		return
	}

	defer srv.log.Printf("[INFO] JES is shutting down.\n")

	// I have initially copied this from some tutorial or documentation, but
	// I am not sure if it is really necessary. OTOH, it does not appear to
	// do any harm.
	http.Handle("/", srv.router)

	if err = srv.srv.ListenAndServe(); err != nil {
		srv.log.Printf("[ERROR] The web server ran into an error: %s\n",
			err.Error())
	}
} // func (j *JES) Run()

//////////////////////////////////////////////////////////////////////////////
/// Handle requests //////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////////////
/// Web service //////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

func (srv *Web) handleSubmit(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle %s from %s\n",
		r.URL,
		r.RemoteAddr)
	var (
		err   error
		msg   string
		buf   []byte
		rbuf  bytes.Buffer
		cmd   command.Command
		job   = new(model.Job)
		reply = ajaxResponse{
			Timestamp: time.Now(),
		}
	)

	if err = r.ParseForm(); err != nil {
		msg = fmt.Sprintf("Cannot parse request form: %s", err.Error())
		srv.log.Printf("[CRITICAL] %s\n",
			msg)
		buf = errJSON(msg)
		goto SEND
	} else if _, err = io.Copy(&rbuf, r.Body); err != nil {
		msg = fmt.Sprintf("Failed to read request body: %s",
			err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		buf = errJSON(msg)
		goto SEND
	} else if err = json.Unmarshal(rbuf.Bytes(), job); err != nil {
		msg = fmt.Sprintf("Cannot parse request body: %s\n\n%s\n",
			err.Error(),
			rbuf.String())
		srv.log.Printf("[ERROR] %s\n", msg)
		buf = errJSON(msg)
		goto SEND
	}

	cmd.Verb = command.Submit
	cmd.Object = job
	srv.mon.CmdQ <- cmd

	reply.Status = true
	reply.Message = "Success"

	if buf, err = json.Marshal(&reply); err != nil {
		msg = fmt.Sprintf("Failed to serialize response: %s",
			err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		buf = errJSON(msg)
		goto SEND
	}

SEND:
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", noCache)
	w.WriteHeader(200)
	w.Write(buf) // nolint: errcheck,gosec
} // func (srv *Web) handleSubmit(w http.ResponseWriter, r *http.Request)
