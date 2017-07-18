package main

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

var InsertRegexp = regexp.MustCompile(`(?i)\s*insert\s`)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "USAGE: kick DATA_SOURCE SQL")
		fmt.Fprintln(os.Stderr, "    DATA_SOURCE: root:@/blocks_subscriber_example1")
		fmt.Fprintln(os.Stderr, "            SQL: INSERT INTO pipeline_jobs (pipeline, progress, created_at, updated_at) VALUES ('pipeline01', 0, NOW(), NOW())")
		fmt.Fprintln(os.Stderr, "kick outputs the result into stdout")
		os.Exit(1)
		return
	}

	datasource := os.Args[1]
	statement := os.Args[2]

	driver := "mysql"
	db, err := sql.Open(driver, datasource)
	ShowAndExitIfError(err, 26)
	defer db.Close()

	r, err := db.Exec(statement)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Statement: %v\n", statement)
	}
	ShowAndExitIfError(err, 30)

	if InsertRegexp.MatchString(statement) {
		lastId, err := r.LastInsertId()
		ShowAndExitIfError(err, 33)
		fmt.Printf("%v\n", lastId)
	} else {
		rows, err := r.RowsAffected()
		ShowAndExitIfError(err, 35)
		fmt.Printf("%v\n", rows)
	}
}

func ShowAndExitIfError(err error, line int) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred at %v\n%+v\n", line, err)
		os.Exit(1)
	}
}
