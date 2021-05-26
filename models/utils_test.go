package models

import (
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestParseTime(t *testing.T) {
	db, err := setupdb(t)
	if err != nil {
		t.Fatalf("cannot setup database %v", err)
	}
	_, err = db.Exec("create table foo(id integer, date datetime)")
	if err != nil {
		t.Fatalf("cannot create table foo, %v", err)
	}
	instant := time.Now()
	_, err = db.Exec("insert into foo (id, date) values (?, ?)", 0, instant)
	if err != nil {
		t.Fatalf("cannot insert into foo, %v", err)
	}
	var result DateFromStr
	err = db.QueryRow("select max(date) from foo").Scan(&result)
	if err != nil {
		t.Fatalf("cannot scan max(date), %v", err)
	}
	if !result.T.Equal(instant) {
		t.Errorf(
			"date from db (%v) is not equal to inserted date (%v)",
			result.T.Format(DATE_SQLITE_FORMAT),
			instant.Format(DATE_SQLITE_FORMAT),
		)
	}
}
