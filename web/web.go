// /home/krylon/go/src/github.com/blicero/jazz/web/web.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-10-01 10:57:56 krylon>

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
	"strconv"
	"sync/atomic"
	"time"

	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/logdomain"
	"github.com/blicero/jazz/model"
	"github.com/blicero/jazz/monitor"
	"github.com/gorilla/mux"
)

const (
	cacheControl = "max-age=120, public"
	noCache      = "no-store, max-age=0"
	tmplFolder   = "assets/templates"
)

// nolint: unused
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

// Create creates and returns a new Web server instance.
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

	if srv.log, err = common.GetLogger(logdomain.Web); err != nil {
		return nil, err
	}

	srv.srv.Addr = addr
	srv.srv.ErrorLog = srv.log
	srv.srv.Handler = srv.router

	srv.router.HandleFunc("/ws/job/new", srv.handleSubmit)
	srv.router.HandleFunc("/ws/job/all", srv.handleQueryQueuedJobs)

	// ...

	return srv, nil
} // func Create(addr string, mon *monitor.Monitor) (*Web, error)

// IsActive returns the value of the Web server's active flag.
func (srv *Web) IsActive() bool {
	return srv.active.Load()
} // func (srv *Web) IsActive() bool

// Stop tells the Web server to stop.
func (srv *Web) Stop() {
	srv.active.Store(false)
	srv.srv.Shutdown(context.Background()) // nolint: errcheck
} // func (srv *Web) Stop()

// Run executes the Web server's main loop.
func (srv *Web) Run() {
	var (
		err     error
		swapped bool
	)

	if swapped = srv.active.CompareAndSwap(false, true); !swapped {
		srv.log.Printf("[INFO] Web server appears to be running already. Toodles!\n")
		return
	}

	defer srv.log.Printf("[INFO] Web server is shutting down.\n")

	// I have initially copied this from some tutorial or documentation, but
	// I am not sure if it is really necessary. OTOH, it does not appear to
	// do any harm.
	http.Handle("/", srv.router)

	if err = srv.srv.ListenAndServe(); err != nil {
		srv.log.Printf("[ERROR] The web server ran into an error: %s\n",
			err.Error())
	}
} // func (srv *Web) Run()

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
		job   = new(model.Job)
		reply = model.WebResponse{
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
	} else if err = srv.mon.SubmitJob(job); err != nil {
		msg = fmt.Sprintf("Failed to submit Job: %s\n",
			err.Error())
		srv.log.Printf("[ERROR] %s\n", err.Error())
		reply.Message = msg
	} else {
		reply.Status = true
		reply.Message = strconv.FormatInt(job.ID, 10)
	}

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

func (srv *Web) handleQueryQueuedJobs(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handle %s from %s\n",
		r.URL,
		r.RemoteAddr)
	var (
		err   error
		jobs  []*model.Job
		msg   string
		buf   []byte
		reply = model.WebResponse{
			Timestamp: time.Now(),
		}
	)

	if jobs, err = srv.mon.GetQueuedJobs(); err != nil {
		reply.Message = fmt.Sprintf("Failed to get queued Jobs from Monitor: %s",
			err.Error())
		srv.log.Printf("[ERROR] %s\n", reply.Message)
		goto SEND
	} else if buf, err = json.Marshal(jobs); err != nil {
		reply.Message = fmt.Sprintf("Failed to serialize Jobs: %s",
			err.Error())
		srv.log.Printf("[ERROR] %s\n",
			reply.Message)
		goto SEND
	}

	reply.Payload = string(buf)

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
} // func (srv *Web) handleQueryQueuedJobs(w http.ResponseWriter, r *http.Request)
