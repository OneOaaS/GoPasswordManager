package main

import "testing"

func TestDBStore(t *testing.T) {
	if s, err := initDB("sqlite3", ":memory:"); err != nil {
		t.Fatal("Could not create database: ", err)
	} else {
		testStores(t, s)
	}
}
