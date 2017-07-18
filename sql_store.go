package main

import (
	"database/sql"
	"runtime"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
)

const (
	SQL_UPDATE_JOBS = "UPDATE pipeline_jobs SET progress = ?, updated_at = ? WHERE job_message_id = ? AND progress < ?"
	SQL_INSERT_LOGS = "INSERT INTO pipeline_job_logs (pipeline, job_message_id, publish_time, progress, completed, log_level, log_message) VALUES (?, ?, ?, ?, ?, ?, ?)"
)

type SqlStore struct {
	db *sql.DB
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

func (ss *SqlStore) save(ctx context.Context, pipeline string, msg *Message, f func() error) error {
	err := ss.insertLog(ctx, pipeline, msg)
	if err != nil {
		return err
	}

	err = ss.transaction(func(tx *sql.Tx) error {
		err := ss.updateJob(ctx, tx, pipeline, msg)
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

func (ss *SqlStore) insertLog(ctx context.Context, pipeline string, msg *Message) error {
	logAttrs := log.Fields(msg.buildMap())

	logAttrs["SQL"] = SQL_INSERT_LOGS
	_, err := ss.db.Exec(SQL_INSERT_LOGS, pipeline, msg.msg_id, msg.publishTime, msg.progress, msg.completedInt(), msg.level, msg.data)
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorln("Failed to insert into pipeline_job_logs")
		return err
	}
	log.WithFields(logAttrs).Debugln("Insert into pipeline_job_logs successfully")
	return nil
}

func (ss *SqlStore) updateJob(ctx context.Context, tx *sql.Tx, pipeline string, msg *Message) error {
	logAttrs := log.Fields(msg.buildMap())
	logAttrs["SQL"] = SQL_UPDATE_JOBS
	_, err := tx.Exec(SQL_UPDATE_JOBS, msg.progress, time.Now(), msg.msg_id, msg.progress)
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
