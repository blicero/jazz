// /home/krylon/go/src/github.com/blicero/jazz/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-10-02 11:16:47 krylon>

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/blicero/jazz/client"
	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/database"
	"github.com/blicero/jazz/logdomain"
	"github.com/blicero/jazz/model"
	"github.com/blicero/jazz/monitor"
	"github.com/blicero/jazz/shell"
	"github.com/blicero/jazz/web"
)

var glog *log.Logger

func main() {
	fmt.Printf("%s %s\nbuilt on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	var (
		err              error
		doComplete       bool
		baseDir, webAddr string
		srv              *web.Web
		mon              *monitor.Monitor
		command          string = "list"
		wsc              *client.Client
	)

	flag.StringVar(
		&baseDir,
		"basedir",
		common.BaseDir,
		"directory for semi-private files, logs, job queue, spooling, etc.",
	)

	flag.StringVar(
		&webAddr,
		"addr",
		common.WebAddr,
		"address of the web server to listen on or talk to",
	)

	flag.BoolVar(
		&doComplete,
		"complete",
		true,
		"enable auto-completion",
	)

	flag.Parse()

	if err = common.SetBaseDir(baseDir); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"cannot initialize environment - %s\n",
			err.Error())
		os.Exit(1)
	} else if flag.NArg() >= 1 {
		command = flag.Arg(0)
	} else {
		fmt.Fprintf(
			os.Stderr,
			"What do you want me to DO???\n")
	}

	if glog, err = common.GetLogger(logdomain.Main); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Cannot initialize Logger: %s\n",
			err.Error())
		os.Exit(1)
	}

	switch strings.ToLower(command) {
	case "submit", "shell":
		var (
			s *shell.Shell
			j *model.Job
		)

		if wsc, err = client.New(webAddr); err != nil {
			glog.Printf("[ERROR] Cannot create WS client for %s: %s\n",
				webAddr,
				err.Error())
			os.Exit(1)
		}

		common.Interactive.Store(true)

		if s, err = shell.Create(doComplete); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Failed to create Shell: %s\n",
				err.Error())
			os.Exit(1)
		} else if j, err = s.Run(); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Error prompting for Job: %s\n",
				err.Error())
			os.Exit(1)
		} else if err = wsc.JobSubmit(j); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Failed to submit Job to Job queue %s: %s\n",
				webAddr,
				err.Error())
			os.Exit(1)
		}

		fmt.Printf("Submit Job:\n%s", j.PrettyPrint())

		os.Exit(0)
	case "list":
		var (
			db   *database.Database
			jobs []*model.Job
		)

		if db, err = database.Open(common.DbPath); err != nil {
			if err.Error() == "timeout" {
				// This most likely means the database is locked
				// because the server is running.
				// In this case, we should try the web service.
				if wsc, err = client.New(webAddr); err != nil {
					glog.Printf("[CRITICAL] Cannot create WS Client: %s\n", err.Error())
					os.Exit(1)
				} else if jobs, err = wsc.QueryQueue(); err != nil {
					glog.Printf("[ERROR] Query for Job Queue failed: %s\n",
						err.Error())
					os.Exit(1)
				}
			} else if jobs, err = db.JobGetAll(); err != nil {
			} else {
				fmt.Fprintf(
					os.Stderr,
					"Cannot list jobs: %s\n",
					err.Error())
				os.Exit(1)
			}
		}

		if len(jobs) == 0 {
			fmt.Println("Job Queue is empty")
		} else {
			for idx, job := range jobs {
				fmt.Printf("%03d Job %d - %s - %s\n",
					idx,
					job.ID,
					job.Name,
					job.WorkDir)
			}
		}
	case "serve", "daemon", "queue":
		fmt.Printf("Run daemon at %s\n",
			webAddr)
		if mon, err = monitor.Create(); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Error creating Monitor: %s\n",
				err.Error(),
			)
			os.Exit(1)
		} else if srv, err = web.Create(webAddr, mon); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Error creating Web server: %s\n",
				err.Error(),
			)
			os.Exit(1)
		}

		sigQ := make(chan os.Signal, 1)
		signal.Notify(sigQ, os.Interrupt, syscall.SIGTERM)

		ticker := time.NewTicker(common.TickInterval)
		defer ticker.Stop()

		mon.Start()
		go srv.Run()

		for {
			select {
			case <-ticker.C:
				// foo
			case s := <-sigQ:
				fmt.Fprintf(
					os.Stderr,
					"Okay, okay, I'm quitting: %s\n",
					s)
				os.Exit(0)
			}
		}
	default:
		fmt.Fprintf(
			os.Stderr,
			"What do you mean I should %s? I'm quitting\n",
			command)
		os.Exit(0)
	}
} // func main()
