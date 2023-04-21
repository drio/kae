package main

import (
	"testing"
)

func TestRSS(t *testing.T) {
	t.Run("XXXXXXXXXXXXXXXXX", func(t *testing.T) {
		got := 1
		want := 1
		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})
}

func initApp(t *testing.T) {
	// db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "db"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
