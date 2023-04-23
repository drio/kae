package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	// Config defaults
	port := 3500
	delaySecs := 5
	dbPath := "keep-an-eye.sqlite"

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage: kae [options]

Options:
  -delaySecs  number of seconds between heartbeat updates (default %d)

Environment variables:
  PORT       HTTP port to listen on (default %d)
  KAE_DB     path to SQLite 3 database (default %q)
`, delaySecs, port, dbPath)
	}
	flag.Parse()

	// Parse config from environment variables
	var err error
	if portEnv, ok := os.LookupEnv("PORT"); ok {
		port, err = strconv.Atoi(portEnv)
		if err != nil {
			exitOnError(err)
		}
	}
	if dbEnv, ok := os.LookupEnv("KAE_DB"); ok {
		dbPath = dbEnv
	}

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath))
	exitOnError(err)
	model, err := NewSQLModel(db)
	exitOnError(err)
	server, err := NewServer(model, log.Default())
	exitOnError(err)

	log.Printf("starting background job")
	go server.runBackgroundJob(bgJobOpts{
		loop: true,
		delayFn: func() {
			server.logger.Printf("runBackgrondJob: Sleeping for %d secs", delaySecs)
			time.Sleep(time.Duration(delaySecs) * time.Second)
		},
	})

	log.Printf("config: port=%d db=%q delaySecs=%d", port, dbPath, delaySecs)
	log.Printf("listening on http://localhost:%d", port)
	exitOnError(http.ListenAndServe(":"+strconv.Itoa(port), server))
	err = http.ListenAndServe(":"+strconv.Itoa(port), server)
	exitOnError(err)
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
