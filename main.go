package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	_ "modernc.org/sqlite"
)

func main() {
	// Config defaults
	port := 3500
	delaySecsDefault := 5
	dbPath := "keep-an-eye.sqlite"

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage: kae [options]

Options:
  -delaySecs  number of seconds between heartbeat updates (default %d)

Environment variables:
  PORT       HTTP port to listen on (default %d)
  KAE_DB     path to SQLite 3 database (default %q)
  KAE_USER   basic auth username (default no basic auth)
  KAE_PASS   basic auth password (default no basic auth)
`, delaySecsDefault, port, dbPath)
	}
	delaySecs := flag.Int("delaySecs", delaySecsDefault, fmt.Sprintf("default: %d", delaySecsDefault))
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
	var user, pass string
	if userEnv, ok := os.LookupEnv("KAE_USER"); ok {
		user = userEnv
	}
	if passEnv, ok := os.LookupEnv("KAE_PASS"); ok {
		pass = passEnv
	}
	if user != "" && pass == "" {
		exitOnError(errors.New("KAE_USER provided but missing KAE_PASS"))
	}
	if user == "" && pass != "" {
		exitOnError(errors.New("KAE_PASS provided but missing KAE_USER"))
	}

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath))
	exitOnError(err)
	model, err := NewSQLModel(db)
	exitOnError(err)
	am := noAuthMiddleware
	if user != "" && pass != "" {
		am = middleware.BasicAuth("kae site", map[string]string{user: pass})
	}
	server, err := NewServer(ServerOpts{
		model:          model,
		logger:         log.Default(),
		authMiddleware: am,
	})
	exitOnError(err)

	log.Printf("starting background job")
	go server.runBackgroundJob(bgJobOpts{
		loop: true,
		delayFn: func() {
			server.logger.Printf("runBackgrondJob: Sleeping for %d secs", *delaySecs)
			time.Sleep(time.Duration(*delaySecs) * time.Second)
		},
	})

	log.Printf("config: port=%d db=%q delaySecs=%d", port, dbPath, *delaySecs)
	log.Printf("listening on http://:%d", port)
	exitOnError(http.ListenAndServe(":"+strconv.Itoa(port), server))
	err = http.ListenAndServe(":"+strconv.Itoa(port), server)
	exitOnError(err)
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
