// /home/krylon/go/src/github.com/blicero/jazz/shell/cli.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-12 11:44:17 krylon>

package shell

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/Feralthedogg/go-functional/pkg/functional"
	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/logdomain"
	"github.com/blicero/jazz/model"
	prompt "github.com/c-bata/go-prompt"
	"github.com/google/shlex"
)

const pprompt = "~> "

var (
	noise     = regexp.MustCompile("[~#]")
	letterPat = regexp.MustCompile("^[a-zA-Z]")
	// farPath   = regexp.MustCompile("^[.]{0,2}/")
)

// Shell is a simple command line shell that prompts for commands to execute
// as a Job.
type Shell struct {
	log         *log.Logger
	shell       *prompt.Prompt
	j           *model.Job
	histfile    *os.File
	pathFolders []string
}

// Create creates (well, duh) and returns a fresh Shell.
func Create(completion bool) (*Shell, error) {
	var (
		err          error
		buf          bytes.Buffer
		env, history []string
		path         string
		s            = &Shell{
			j: &model.Job{
				Steps: make([]model.Step, 0),
				Env:   make(map[string]string),
			},
		}
	)

	if s.log, err = common.GetLogger(logdomain.Shell); err != nil {
		return nil, err
	} else if s.j.WorkDir, err = os.Getwd(); err != nil {
		return nil, err
	} else if s.histfile, err = os.OpenFile(
		common.HistPath,
		os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_SYNC,
		0644); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			history = make([]string, 0)
			goto ENV
		}
		s.log.Printf("[ERROR] Cannot open history file %s for reading: %s\n",
			common.HistPath,
			err.Error())
		return nil, err
	}

	if _, err = io.Copy(&buf, s.histfile); err != nil {
		s.log.Printf("[ERROR] Failed to read history file %s: %s\n",
			common.HistPath,
			err.Error())
		return nil, err
	}

	history = strings.Split(buf.String(), "\n")

	if path = os.Getenv("PATH"); path == "" {
		s.log.Printf("[ERROR] Environment variable PATH is empty\n")
	} else {
		s.pathFolders = strings.Split(path, ":")
	}

ENV:
	env = os.Environ()

	for _, v := range env {
		var (
			pieces    []string
			name, val string
		)

		pieces = strings.Split(v, "=")
		name = pieces[0]
		val = pieces[1]

		s.j.Env[name] = val
	}

	if completion {
		s.shell = prompt.New(
			s.executor,
			s.completer,
			prompt.OptionPrefix(pprompt),
			prompt.OptionHistory(history))
	} else {
		s.shell = prompt.New(
			s.executor,
			s.completerDummy,
			prompt.OptionPrefix(pprompt),
			prompt.OptionHistory(history))
	}

	return s, nil
} // func Create() (*Shell, error)

// Run runs the Shell. If everything goes well, it returns a Job that can be
// submitted to the monitor.
func (s *Shell) Run() (j *model.Job, e error) {

	defer func() {
		if x := recover(); x != nil {
			s.log.Printf("[ERROR] Panic while prompting for Job: %#v\n",
				x)
			e = x.(error)
		}
	}()

	s.shell.Run()

	return s.j, nil
}

func (s *Shell) executor(input string) {
	s.log.Printf("[TRACE] Executing: %s\n",
		input)

	var (
		err  error
		step = model.Step{
			Command: input,
		}
	)

	s.j.Steps = append(s.j.Steps, step)

	if !strings.HasSuffix(input, "\n") {
		input = input + "\n"
	}

	if _, err = s.histfile.Write([]byte(input)); err != nil {
		s.log.Printf("[ERROR] Cannot append to history file: %s\n",
			err.Error())
	}
} // func (s *Shell) executor(input string

func (s *Shell) completer(d prompt.Document) []prompt.Suggest {
	var (
		err           error
		line, word    string
		tokens, files []string
		sugg          []prompt.Suggest
		widx          int
	)

	line = d.CurrentLine()
	if tokens, err = shlex.Split(line); err != nil {
		s.log.Printf("[ERROR] Cannot tokenize line %q: %s\n",
			line,
			err.Error())
		return nil
	}

	word = d.GetWordBeforeCursor()
	widx = slices.Index(tokens, word)

	s.log.Printf("[TRACE] Current line: %q, word #%d: %q\n",
		line,
		widx,
		word)

	if widx == 0 {
		return s.completerExecutables(d)
	} else if files, err = s.readFolder(word); err != nil {
		return nil
	}

	sugg = make([]prompt.Suggest, len(files))

	for i, f := range files {
		sugg[i] = prompt.Suggest{
			Text: f,
		}
	}

	return sugg
} // func (s *Shell) completer(d prompt.Document) []prompt.Suggest

// completerExecutables looks up all executable files in the folders from PATH
// and returns the ones that match the text typed so far.
func (s *Shell) completerExecutables(d prompt.Document) []prompt.Suggest {
	var (
		err      error
		binaries = make([]prompt.Suggest, 0)
		word     = d.GetWordBeforeCursor()
	)

	if !letterPat.MatchString(word) {
		return binaries
	}

	for _, dir := range s.pathFolders {
		var (
			files []string
		)

		if files, err = s.readFolder(dir); err != nil {
			continue
		}

		for _, f := range files {
			var base = filepath.Base(f)
			if strings.HasPrefix(base, word) {
				// s.log.Printf("[TRACE] Consider %s\n",
				// 	f)
				binaries = append(binaries,
					prompt.Suggest{
						Text:        base,
						Description: f,
					})
			}
		}
	}

	return binaries
} // func (s *Shell) completerExecutables(d prompt.Document) []prompt.Suggest

func (s *Shell) completerDummy(d prompt.Document) []prompt.Suggest {
	return make([]prompt.Suggest, 0)
} // func (s *Shell) completerDummy(d prompt.Document) []prompt.Suggest

func (s *Shell) readFolder(path string) ([]string, error) {
	var (
		err   error
		fh    *os.File
		info  []os.FileInfo
		files []string
	)

	s.log.Printf("[TRACE] Attempt to find completions for %q in current folder\n",
		path)

	if fh, err = os.Open(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}

		return nil, err
	}

	defer fh.Close() // nolint: errcheck

	if info, err = fh.Readdir(-1); err != nil {
		return nil, err
	}

	info = slices.DeleteFunc(info, func(i os.FileInfo) bool {
		return !i.Mode().IsRegular() ||
			noise.MatchString(i.Name()) ||
			(i.Mode().Perm()&0111 != 0111)
	})

	files = functional.Map(func(i os.FileInfo) string {
		return filepath.Join(path, i.Name())
	}, info)

	return files, nil
} // func (s *Shell) readFolder(path string) ([]string, error)

// func (s *Shell) glob(word string) ([]string, error) {
// 	var (
// 		err   error
// 		dh    *os.File
// 		files []string
// 		finfo []os.FileInfo
// 	)

// 	if dh, err = os.Open("."); err != nil {
// 		s.log.Printf("[ERROR] Cannot open current directory: %s\n",
// 			err.Error())
// 		return nil, err
// 	}

// 	defer dh.Close() // nolint: errcheck

// 	if finfo, err = dh.Readdir(-1); err != nil {
// 		s.log.Printf("[ERROR] Cannot read current directory: %s\n",
// 			err.Error())
// 		return nil, err
// 	}

// 	files = make([]string, 0, len(finfo))

// 	for _, info := range finfo {
// 		var name = info.Name()
// 		if !(info.Mode().IsRegular() || info.Mode().IsDir()) {
// 			continue
// 		} else if noise.MatchString(name) {
// 			continue
// 		} else if strings.HasPrefix(name, word) {
// 			files = append(files, name)
// 		}
// 	}

// 	return files, nil
// } // func (s *Shell) glob(word string) ([]string, error)
