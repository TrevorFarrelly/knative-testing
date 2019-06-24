package main

import (
  "log"
  _ "github.com/go-sql-driver/mysql"
  "database/sql"
  "database/sql/driver"
)

const (
  socket = "/usr/local/google/home/trevorfarrelly/go/src/github.com/TrevorFarrelly/knative-testing/mysql-experiment/trevorfarrelly-knative-2019:us-west1-b:flaky-testing"
)

func connect() *sql.DB {
  conn, err := sql.Open("mysql", socket)
  if err != nil {
    log.Fatalf("Error getting connection: %v\n", err)
  }
  if conn.Ping() == driver.ErrBadConn {
    log.Fatalf("Could not connect to database\n")
  }
  return conn
}

func start(db *sql.DB, tablename string) {
  if _, err := db.Exec("CREATE TABLE IF NOT EXIST ? (job VARCHAR(255), name VARCHAR(255), isflaky BIT, PRIMARY KEY job)", tablename); err != nil {
    log.Printf("unable to create table %s\n", tablename)
  }

}

func main() {
  db := connect()
  start(db, "flakystatus")
}
