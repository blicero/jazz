// /home/krylon/go/src/github.com/blicero/jazz/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 31. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-19 12:18:10 krylon>

package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/jes"
	"github.com/blicero/jazz/model"
	"github.com/blicero/jazz/monitor"
	"github.com/blicero/jazz/monitor/command"
	"github.com/blicero/jazz/shell"
)

func main() {
	fmt.Printf("%s %s\nbuilt on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	var (
		err                  error
		doSubmit, doComplete bool
		baseDir, sockPath    string
		watcher              *jes.JES
		mon                  *monitor.Monitor
	)

	flag.StringVar(
		&baseDir,
		"basedir",
		common.BaseDir,
		"directory for semi-private files, logs, job queue, spooling, etc.",
	)

	flag.StringVar(
		&sockPath,
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
		} else if err = submitJob(sockPath, j); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Failed to submit Job to Job queue %s: %s\n",
				sockPath,
				err.Error())
			os.Exit(1)
		}

		fmt.Printf("Submit Job:\n%s", j.PrettyPrint())

		os.Exit(0)
	} else if watcher, err = jes.Create(sockPath); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Error creating JES: %s\n",
			err.Error(),
		)
		os.Exit(1)
	} else if mon, err = monitor.Create(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Error creating Monitor: %s\n",
			err.Error(),
		)
		os.Exit(1)
	}

	sigQ := make(chan os.Signal, 1)
	signal.Notify(sigQ, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(common.TickInterval)
	defer ticker.Stop()

	mon.Start()
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
		case job := <-watcher.JobQ:
			cmd := command.Command{
				Verb:   command.Submit,
				Object: job,
			}
			go func() {
				mon.CmdQ <- cmd
			}()
		}
	}
} // func main()

// TODO switch to HTTP interface, and maybe roll that back into the Shell?
func submitJob(path string, j *model.Job) error {
	var (
		err     error
		sock    *net.UnixConn
		enc     *gob.Encoder
		buf     bytes.Buffer
		n, oobn int
		addr    = net.UnixAddr{
			Name: path,
			Net:  "unixgram",
		}
	)

	if sock, err = net.DialUnix("unixgram", nil, &addr); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to connect to socket %s: %s\n",
			path,
			err.Error())
		return err
	}

	enc = gob.NewEncoder(&buf)

	if err = enc.Encode(j); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to serialize Job: %s\n",
			err.Error(),
		)
		return err
	} else if n, oobn, err = sock.WriteMsgUnix(buf.Bytes(), nil, &addr); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to submit Job: %s\n",
			err.Error())
		return err
	} else if n != len(buf.Bytes()) { // CANTHAPPEN
		fmt.Fprintf(
			os.Stderr,
			"Error: We only sent %d of %d bytes\n",
			n,
			len(buf.Bytes()))
	} else if oobn != 0 { // CANTHAPPEN
		fmt.Fprintf(
			os.Stderr,
			"We didn't send any oob data, but %d bytes were sent anyway?!\n",
			oobn,
		)

	}

	return nil
} // func submitJob(j *model.Job) error
