// /home/krylon/go/src/github.com/blicero/jazz/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-08 11:45:48 krylon>

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/jes"
	"github.com/blicero/jazz/model"
	"github.com/blicero/jazz/shell"
)

func main() {
	fmt.Printf("%s %s\nbuilt on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	var (
		err      error
		doSubmit bool
		baseDir  string
		watcher  *jes.JES
	)

	flag.StringVar(
		&baseDir,
		"basedir",
		common.BaseDir,
		"directory for semi-private files, logs, job queue, spooling, etc.",
	)

	flag.BoolVar(
		&doSubmit,
		"submit",
		false,
		"prompt for a new Job to submit",
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

		if s, err = shell.Create(); err != nil {
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
		}

		fmt.Println("Submit Job:")
		for _, step := range j.Steps {
			fmt.Printf(">> %s\n",
				step.Command)
		}

		// do something!
		fmt.Println("Sleepy now")
		time.Sleep(time.Second * 10)

		// later:
		os.Exit(0)
	}

	sigQ := make(chan os.Signal, 1)
	signal.Notify(sigQ, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(common.TickInterval)
	defer ticker.Stop()

	_ = watcher.Start()

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
}
