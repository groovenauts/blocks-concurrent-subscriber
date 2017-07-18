package main

import (
	"database/sql"
	"runtime"

	_ "github.com/go-sql-driver/mysql"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
)

type SqlStore struct {
	db *sql.DB
	insertTemplate *SqlTemplate
	updateTemplate *SqlTemplate
}

func (ss *SqlStore) setup(ctx context.Context, driver, datasource string) (func() error, error) {
	log.Infof("Connecting to %v database by %v\n", driver, datasource)
	db, err := sql.Open(driver, datasource)
	if err != nil {
		log.WithFields(log.Fields{"datasource": datasource}).Errorln(err)
		return nil, err
	}
	ss.db = db
	return db.Close, nil
}

func (ss *SqlStore) save(ctx context.Context, msg *Message, f func() error) error {
	err := ss.insertLog(ctx, msg)
	if err != nil {
		return err
	}

	err = ss.transaction(func(tx *sql.Tx) error {
		err := ss.updateJob(ctx, tx, msg)
		if err != nil {
			return err
		}

		if f != nil {
			err = f()
			return err
		}

		return nil
	})
	return err
}

func (ss *SqlStore) insertLog(ctx context.Context, msg *Message) error {
	if ss.insertTemplate == nil {
		return nil
	}
	logAttrs := log.Fields(msg.buildMap())
	logAttrs["SQL"] = ss.insertTemplate.Body
	_, err := ss.db.Exec(ss.insertTemplate.Body, msg.paramValues(ss.insertTemplate.Parameters)...)
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorln("Failed to insert into pipeline_job_logs")
		return err
	}
	log.WithFields(logAttrs).Debugln("Insert into pipeline_job_logs successfully")
	return nil
}

func (ss *SqlStore) updateJob(ctx context.Context, tx *sql.Tx, msg *Message) error {
	if ss.updateTemplate == nil {
		return nil
	}
	logAttrs := log.Fields(msg.buildMap())
	logAttrs["SQL"] = ss.updateTemplate.Body
	_, err := tx.Exec(ss.updateTemplate.Body, msg.paramValues(ss.updateTemplate.Parameters)...)
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorln("Failed to update pipeline_jobs")
		return err
	}
	log.WithFields(logAttrs).Debugln("Update pipeline_jobs successfully")
	return nil
}

// Use "err" for returned variable name in order to return the error on recover.
func (ss *SqlStore) transaction(impl func(tx *sql.Tx) error) (err error) {
	tx, err := ss.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("recover in SqlStore.transaction r: ", r)
			tx.Rollback()
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()
	err = impl(tx)
	if err == nil {
		log.Debugln("Commit")
		tx.Commit()
	} else {
		log.Debugln("Rollback")
		tx.Rollback()
	}
	return err
}
