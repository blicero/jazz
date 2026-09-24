// /home/krylon/go/src/github.com/blicero/jazz/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-21 21:15:58 krylon>

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/model"
	"github.com/blicero/jazz/monitor"
	"github.com/blicero/jazz/shell"
	"github.com/blicero/jazz/web"
)

func main() {
	fmt.Printf("%s %s\nbuilt on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	var (
		err                  error
		doSubmit, doComplete bool
		baseDir, webAddr     string
		srv                  *web.Web
		mon                  *monitor.Monitor
	)

	flag.StringVar(
		&baseDir,
		"basedir",
		common.BaseDir,
		"directory for semi-private files, logs, job queue, spooling, etc.",
	)

	flag.StringVar(
		&webAddr,
		"socket",
		common.SockPath,
		"path for the socket for submitting Jobs",
	)

	flag.BoolVar(
		&doSubmit,
		"submit",
		false,
		"prompt for a new Job to submit",
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
	}

	if doSubmit {
		var (
			s *shell.Shell
			j *model.Job
		)

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
		} else if err = submitJob(webAddr, j); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Failed to submit Job to Job queue %s: %s\n",
				webAddr,
				err.Error())
			os.Exit(1)
		}

		fmt.Printf("Submit Job:\n%s", j.PrettyPrint())

		os.Exit(0)
	} else if mon, err = monitor.Create(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Error creating Monitor: %s\n",
			err.Error(),
		)
		os.Exit(1)
	} else if srv, err = web.Create(webAddr, mon); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Error creating JES: %s\n",
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
} // func main()
