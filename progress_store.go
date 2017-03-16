package main

import (
	"database/sql"
	"fmt"
	"runtime"

	_ "github.com/go-sql-driver/mysql"

	"golang.org/x/net/context"
)

const (
	SQL_UPDATE_JOBS = "UPDATE pipeline_jobs SET progress = ? WHERE job_message_id = ? AND progress < ?"
	SQL_INSERT_LOGS = "INSERT INTO pipeline_job_logs (pipeline, job_message_id, publish_time, progress, completed, log_level, log_message) VALUES (?, ?, ?, ?, ?, ?, ?)"
)

type SqlStore struct {
	db *sql.DB
}

func (ss *SqlStore) setup(ctx context.Context, driver, datasource string) (func() error, error) {
	fmt.Printf("Connecting to %v database by %v\n", driver, datasource)
	db, err := sql.Open(driver, datasource)
	if err != nil {
		fmt.Println("Failed to get database connection for ", datasource, " cause of ", err)
		return nil, err
	}
	ss.db = db
	return db.Close, nil
}

func (ss *SqlStore) save(ctx context.Context, pipeline string, msg *Message, f func() error) error {
	err := ss.transaction(func(tx *sql.Tx) error {
		_, err := tx.Exec(SQL_UPDATE_JOBS, msg.progress, msg.msg_id, msg.progress)
		if err != nil {
			fmt.Printf("Failed to update pipeline_jobs job_message_id: %v to progress: %v cause of %v", msg.msg_id, msg.progress, err)
			return err
		}

		_, err = tx.Exec(SQL_INSERT_LOGS, pipeline, msg.msg_id, msg.publishTime, msg.progress, msg.completedInt(), msg.level, msg.data)
		if err != nil {
			fmt.Printf("Failed to insert pipeline_job_logs pipeline: %v, job_message_id: %v, progress: %v cause of %v", pipeline, msg.msg_id, msg.progress, err)
			return err
		}

		if f != nil {
			err = f()
			return err
		}

		return nil
	})
	if err != nil {
		fmt.Printf("Failed to begin a transaction job_message_id: %v to progress: %v cause of %v", msg.msg_id, msg.progress, err)
	}
	return err
}

// Use "err" for returned variable name in order to return the error on recover.
func (ss *SqlStore) transaction(impl func(tx *sql.Tx) error) (err error) {
	tx, err := ss.db.Begin()
	if err != nil {
		fmt.Printf("Failed to begin transaction cause of %v\n", err)
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recover in SqlStore.transaction r: %v\n", r)
			tx.Rollback()
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()
	err = impl(tx)
	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	return err
}
