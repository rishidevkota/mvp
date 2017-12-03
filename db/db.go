package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)


var db *sql.DB
var dbDriverName string

var stmts = make(map[string]*sql.Stmt)

type Row struct {
	*sql.Row
}

type Rows struct {
	*sql.Rows
}

func init() {
	mdb, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	db=mdb
}

func patch(args []interface{}) []interface{} {
	var pArgs []interface{}
	for _, arg := range args {
		switch v := arg.(type) {
		case bool:
			if v {
				pArgs = append(pArgs, 1)
			} else {
				pArgs = append(pArgs, 0)
			}
		default:
			pArgs = append(pArgs, v)
		}
	}
	return pArgs
}

func makeStmt(query string) *sql.Stmt {
	stmt, ok := stmts[query]
	if !ok {
		stmt, err := db.Prepare(query)
		if err != nil {
			log.Panicf("[ERROR] Error making stmt: %s. Err msg: %s\n", query, err)
		}
		stmts[query] = stmt
		return stmt
	}
	return stmt

}

func QueryRow(query string, args ...interface{}) *Row {
	stmt := makeStmt(query)
	pArgs := patch(args)
	return &Row{stmt.QueryRow(pArgs...)}
}

func Query(query string, args ...interface{}) *Rows {
	stmt := makeStmt(query)
	pArgs := patch(args)
	rows, err := stmt.Query(pArgs...)
	if err != nil {
		log.Panicf("[ERROR] Error with SQL query '%s': %s\n", query, err)
	}
	return &Rows{rows}
}

func (r *Row) Scan(args ...interface{}) error {
	err := r.Row.Scan(args...)
	switch {
	case err == sql.ErrNoRows:
		return err
	case err != nil:
		log.Panicf("[ERROR] Error scanning row: %s\n", err)
	}
	return nil
}

func (rs *Rows) Scan(args ...interface{}) error {
	err := rs.Rows.Scan(args...)
	switch {
	case err == sql.ErrNoRows:
		return err
	case err != nil:
		log.Panicf("[ERROR] Error scanning rows: %s\n", err)
	}
	return nil
}

func Exec(query string, args ...interface{}) sql.Result {
	pArgs := patch(args)
	stmt := makeStmt(query)
	result, err := stmt.Exec(pArgs...) 
	if err != nil {
		log.Panicf("[ERROR] Error executing %s. Err msg: %s\n", query, err)
	}

	return result
}