// /home/krylon/go/src/github.com/blicero/jazz/database/database.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 09. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-09-02 12:06:23 krylon>

// Package database provides persistence for the Job Queue.
package database

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"strconv"
	"sync"

	"github.com/blicero/jazz/common"
	"github.com/blicero/jazz/logdomain"
	"github.com/blicero/jazz/model"
	bolt "go.etcd.io/bbolt"
)

const bucketName = "jobqueue"

var (
	// ErrNotFound indicates an inexistant key.
	ErrNotFound = errors.New("ID was not found in database")
	// ErrExistentialCrisis indicates something went *deeply* wrong.
	ErrExistentialCrisis = errors.New("i am in an existential crisis right now")
	openLock             sync.Mutex
)

// Database deals with persistent data.
type Database struct {
	path string
	log  *log.Logger
	db   *bolt.DB
}

// Open opens a Database at the given filename <path>. The file does not
// need to exist prior to opening.
func Open(path string) (*Database, error) {
	openLock.Lock()
	defer openLock.Unlock()

	var (
		err error
		db  = &Database{path: path}
	)

	if db.log, err = common.GetLogger(logdomain.Database); err != nil {
		return nil, err
	} else if db.db, err = bolt.Open(path, 0600, nil); err != nil {
		db.log.Printf("[CRITICAL] Failed to open database at %s: %s\n",
			path,
			err.Error())
		return nil, err
	}

	return db, nil
} // func Open(path string) (*Database, error)

// JobAdd saves a Job to the database.
func (db *Database) JobAdd(j *model.Job) error {
	var (
		err error
	)

	// if err = enc.Encode(j); err != nil {
	// 	db.log.Printf("[ERROR] Failed to serialize Job %d: %s\n",
	// 		j.ID,
	// 		err.Error())
	// 	return err
	// }

	// //keybuf = []byte(strconv.FormatInt(j.ID, 10))
	// valbuf = encbuf.Bytes()

	err = db.db.Update(func(tx *bolt.Tx) error {
		var (
			ex             error
			id             uint64
			bucket         *bolt.Bucket
			keybuf, valbuf []byte
			encbuf         bytes.Buffer
			enc            = gob.NewEncoder(&encbuf)
		)

		if bucket, ex = tx.CreateBucketIfNotExists([]byte(bucketName)); ex != nil {
			db.log.Printf("[CRITICAL] Failed to create Bucket %s: %s\n",
				bucketName,
				err.Error())
			return ex
		} else if id, ex = bucket.NextSequence(); ex != nil {
			db.log.Printf("[ERROR] Cannot get ID for new Job: %s\n",
				ex.Error())
			return ex
		}

		j.ID = int64(id)

		if ex = enc.Encode(j); ex != nil {
			db.log.Printf("[ERROR] Failed to serialize Job %d: %s\n",
				j.ID,
				ex.Error())
			return err
		}

		keybuf = []byte(strconv.FormatInt(j.ID, 10))
		valbuf = encbuf.Bytes()

		if ex = bucket.Put(keybuf, valbuf); ex != nil {
			db.log.Printf("[ERROR] Failed to save Job %d: %s\n",
				j.ID,
				err.Error())
			return ex
		}

		return nil
	})

	return err
} // func (db *Database) JobAdd(j *model.Job) error

// JobGet retrieves a single Job by its ID.
func (db *Database) JobGet(id int64) (*model.Job, error) {
	var (
		err            error
		keybuf, valbuf []byte
		decbuf         *bytes.Buffer
		dec            *gob.Decoder
		j              *model.Job
	)

	keybuf = []byte(strconv.FormatInt(id, 10))

	err = db.db.View(func(tx *bolt.Tx) error {
		var bucket *bolt.Bucket

		if bucket = tx.Bucket([]byte(bucketName)); bucket == nil {
			db.log.Printf("[CRITICAL] Bucket %s does not exist?!?\n",
				bucketName)
			return ErrExistentialCrisis
		} else if valbuf = bucket.Get(keybuf); valbuf == nil {
			db.log.Printf("[DEBUG] Job %d does not exist in datbase\n",
				id)
			return ErrNotFound
		}

		return nil
	})

	if err != nil {
		db.log.Printf("[ERROR] Failed to lookup Job %d: %s\n",
			id,
			err.Error())
		return nil, err
	}

	decbuf = bytes.NewBuffer(valbuf)
	dec = gob.NewDecoder(decbuf)
	j = new(model.Job)

	if err = dec.Decode(j); err != nil {
		db.log.Printf("[ERROR] Failed to de-serialize Job %d: %s\n",
			id,
			err.Error())
		return nil, err
	}

	return j, nil
} // func (db *Database) JobGet(id int64) (*model.Job, error)

// JobDelete removes a Job from the Datbase.
func (db *Database) JobDelete(j *model.Job) error {
	var (
		err    error
		keybuf []byte
	)

	keybuf = []byte(strconv.FormatInt(j.ID, 10))

	err = db.db.Update(func(tx *bolt.Tx) error {
		var bucket = tx.Bucket([]byte(bucketName))
		return bucket.Delete(keybuf)
	})

	return err
} // func (db *Database) JobDelete(j *model.Job) error

// JobGetAll loads all Jobs from the Database.
func (db *Database) JobGetAll() ([]*model.Job, error) {
	var (
		err  error
		jobs []*model.Job
	)

	jobs = make([]*model.Job, 0)

	err = db.db.View(func(tx *bolt.Tx) error {
		var (
			ex     error
			bucket *bolt.Bucket
			cur    *bolt.Cursor
		)

		bucket = tx.Bucket([]byte(bucketName))
		cur = bucket.Cursor()

		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			var (
				decbuf = bytes.NewBuffer(v)
				dec    = gob.NewDecoder(decbuf)
				j      = new(model.Job)
			)

			if ex = dec.Decode(j); ex != nil {
				db.log.Printf("[ERROR] Failed to restore Job %s: %s\n",
					k,
					err.Error())
				continue
			}

			jobs = append(jobs, j)
		}

		return nil
	})

	return jobs, err
} // func (db *Database) JobGetAll() ([]*model.Job, error)
