// /home/krylon/go/src/github.com/blicero/jazz/shell/pragma.go
// -*- mode: go; coding: utf-8; -*-
// Created on 12. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-12 13:01:22 krylon>

package shell

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/blicero/jazz/common"
	"github.com/google/shlex"
)

func (s *Shell) processPragma(p string) error {
	var (
		err    error
		tokens []string
	)

	if tokens, err = shlex.Split(p); err != nil {
		return err
	} else if len(tokens) == 0 {
		s.log.Printf("[TRACE] No tokens in pragma %q\n",
			p)
		return nil
	}

	switch strings.ToLower(tokens[0]) {
	case "env":
		return s.processEnv(tokens[1:])
	case "nice":
		if len(tokens) != 2 {
			err = fmt.Errorf("invalid number of tokens: %d (!= 2)",
				len(tokens))
			s.log.Printf("[ERROR] %s\n",
				err.Error())
			return err
		}
		var nice int64

		if nice, err = strconv.ParseInt(tokens[1], 10, 8); err != nil {
			s.log.Printf("[ERROR] Cannot parse nice value %q: %s\n",
				tokens[1],
				err.Error())
			return err
		} else if nice < 0 && os.Geteuid() != 0 {
			err = fmt.Errorf("only root can set negative nice values (%d)",
				nice)
			s.log.Printf("[ERROR] %s\n",
				err.Error())
			return err
		}

		s.j.Niceness = nice
	case "start":
		return s.processStartTime(tokens[1:])
	case "deadline":
		return s.processDeadline(tokens[1:])
	default:
		s.log.Printf("[TRACE] Unknown pragma %q\n",
			tokens[0])
	}

	// ...

	return nil
} // func (s *Shell) processPragma(p string) error

func (s *Shell) processEnv(tokens []string) error {
	for _, envvar := range tokens[1:] {
		var (
			err             error
			pieces          []string
			key, val, exval string
			ok              bool
		)

		pieces = strings.Split(envvar, "=")

		if len(pieces) != 2 {
			err = fmt.Errorf(
				"unexpected number of pieces in %q: %d (expected 2)",
				envvar,
				len(pieces))
			s.log.Printf("[INFO] \n",
				err.Error())
			return err
		}

		key, val = pieces[0], pieces[1]

		if exval, ok = s.j.Env[key]; ok {
			err = fmt.Errorf("environment variable %s already exists: %q\n",
				key,
				exval)
			s.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}

		s.j.Env[key] = strings.Trim(val, `"`)
	}

	return nil
} // func (s *Shell) processEnv(tokens []string) error

func (s *Shell) processStartTime(tokens []string) error {
	var (
		err     error
		tstring string
		start   time.Time
	)

	tstring = strings.Join(tokens, " ")

	if start, err = dateparse.ParseLocal(tstring); err != nil {
		s.log.Printf("[ERROR] Cannot parse datetime %q: %s\n",
			tstring,
			err.Error())
		return err
	} else if start.Before(time.Now().Truncate(time.Second)) {
		err = fmt.Errorf("start time is already %s in the past: %s",
			time.Since(start),
			start.Format(common.TimestampFormat))
		return err
	}

	s.j.ScheduledStart = start

	return nil
} // func (s *Shell) processStartTime(tokens []string) error

func (s *Shell) processDeadline(tokens []string) error {
	var (
		err      error
		tstring  string
		deadline time.Time
	)

	tstring = strings.Join(tokens, " ")

	if deadline, err = dateparse.ParseLocal(tstring); err != nil {
		s.log.Printf("[ERROR] Cannot parse datetime %q: %s\n",
			tstring,
			err.Error())
		return err
	} else if deadline.Before(time.Now().Truncate(time.Second)) {
		err = fmt.Errorf("deadline expired %s ago (%s)",
			time.Since(deadline),
			deadline.Format(common.TimestampFormat))
		s.log.Printf("[ERROR] %s\n", err.Error())
		return err
	}

	s.j.Deadline = deadline
	return nil
} // func (s *Shell) processDeadline(tokens []string) error
