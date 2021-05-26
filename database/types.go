package database

import "database/sql"

type Execer interface {
    Exec(query string, args ...interface{}) (sql.Result, error)
}

type Queryer interface {
    Query(query string, args ...interface{}) (*sql.Rows, error)
}

type BaseRunner interface {
    Execer
    Queryer
}
