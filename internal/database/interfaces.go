package database

import (
	"database/sql"
)

// DBTX interface combines sql.DB and sql.Tx methods
type DBTX interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}
