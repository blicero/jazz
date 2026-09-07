// /home/krylon/go/src/github.com/blicero/jazz/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-04 13:28:13 krylon>

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
)

func main() {
	fmt.Printf("%s %s\nbuilt on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	var (
		err     error
		baseDir string
		jesDir  string
		watcher *jes.JES
	)

	flag.StringVar(
		&baseDir,
		"basedir",
		common.BaseDir,
		"directory for semi-private files, logs, job queue, spooling, etc.",
	)
	flag.StringVar(
		&jesDir,
		"jsa",
		fmt.Sprintf("/tmp/jazz.%s.d",
			os.Getenv("USER")),
		"directory for job entry",
	)

	flag.Parse()

	if err = common.SetBaseDir(baseDir); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"cannot initialize environment - %s\n",
			err.Error())
		os.Exit(1)
	} else if watcher, err = jes.Create(jesDir); err != nil {
		fmt.Fprintf(os.Stderr, "cannot create JES monitor in %s: %s\n",
			jesDir,
			err.Error())
		os.Exit(1)
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
