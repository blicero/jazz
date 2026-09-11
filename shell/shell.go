// /home/krylon/go/src/github.com/blicero/jazz/shell/cli.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-11 11:32:10 krylon>

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
	noise   = regexp.MustCompile("[~#]")
	farPath = regexp.MustCompile("^[.]{0,2}/")
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
func Create() (*Shell, error) {
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

	s.shell = prompt.New(
		s.executor,
		s.completer,
		prompt.OptionPrefix(pprompt),
		prompt.OptionHistory(history))

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
		err         error
		line        string
		tokens      []string
		suggestions = make([]prompt.Suggest, 0)
	)

	line = d.CurrentLine()
	s.log.Printf("[DEBUG] Complete current line: %s\n",
		line)
	if tokens, err = shlex.Split(line); err != nil {
		s.log.Printf("[ERROR] Cannot tokenize input (%s): %s\n",
			line,
			err.Error())
		return suggestions
	} else if len(tokens) == 0 {
		return suggestions
	} else if len(tokens) == 1 && !farPath.MatchString(tokens[0]) {
		// complete name of executable
		return s.completerExecutables(d)
	}

	var (
		word = d.GetWordBeforeCursor()
		idx  = slices.Index(tokens, word)
	)

	s.log.Printf("[TRACE] word = %s idx = %d, tokens = %#v\n",
		word,
		idx,
		tokens)

	if word == "" {
		return suggestions
	}

	// If it's not the first word, attempt to complete a filename.
	var (
		fh       *os.File
		cwd, dir string
		files    []string
	)

	if cwd, err = os.Getwd(); err != nil {
		s.log.Printf("[ERROR] Cannot query current directory from OS: %s\n",
			err.Error())
		return suggestions
	}

	// If the input is an absolute or relative path, we need to chase that down
	if filepath.IsLocal(word) {
		dir = cwd
	} else if farPath.MatchString(word) {
		dir = filepath.Dir(word)
	} else {
		dir = filepath.Clean(filepath.Join(cwd, word))
	}

	if fh, err = os.Open(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return suggestions
		}

		s.log.Printf("[ERROR] Cannot open directory %s: %s\n",
			dir,
			err.Error())
		return suggestions
	}

	defer fh.Close() // nolint: errcheck

	if files, err = fh.Readdirnames(-1); err != nil {
		s.log.Printf("[ERROR] Cannot from contents of directory %s: %s\n",
			cwd,
			err.Error())
		return suggestions
	}

	for _, file := range files {
		if strings.HasPrefix(file, word) && !noise.MatchString(file) {
			s := prompt.Suggest{
				Text:        file,
				Description: filepath.Join(cwd, file),
			}
			suggestions = append(suggestions, s)
		}
	}

	return suggestions
} // func (s *Shell) completer(d prompt.Document) []prompt.Suggest

// completerExecutables looks up all executable files in the folders from PATH
// and returns the ones that match the text typed so far.
func (s *Shell) completerExecutables(d prompt.Document) []prompt.Suggest {
	var (
		err         error
		sugg        = []prompt.Suggest{}
		word        string
		executables []string
	)

	for _, dir := range s.pathFolders {
		var files []string

		if files, err = s.readFolder(dir); err != nil {
			s.log.Printf("[ERROR] Cannot read %s: %s\n",
				dir,
				err.Error())
			return sugg
		}

		executables = append(executables, files...)
	}

	word = d.GetWordBeforeCursor()

	s.log.Printf("[TRACE] Looking for executables like %q\n",
		word)

	for _, ex := range executables {
		var base = filepath.Base(ex)

		if noise.MatchString(base) {
			continue
		} else if strings.HasPrefix(base, word) {
			s := prompt.Suggest{
				Text:        base,
				Description: ex,
			}

			sugg = append(sugg, s)
		}
	}

	return sugg
} // func (s *Shell) completerExecutables(d prompt.Document) []prompt.Suggest

func (s *Shell) readFolder(path string) ([]string, error) {
	var (
		err   error
		fh    *os.File
		info  []os.FileInfo
		files []string
	)

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
		return i.Mode().IsRegular() &&
			i.Mode().Perm()&0111 == 0111
	})

	files = functional.Map(func(i os.FileInfo) string {
		return filepath.Join(path, i.Name())
	}, info)

	return files, nil
} // func (s *Shell) readFolder(path string) ([]string, error)
