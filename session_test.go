package main

import "testing"

func TestSession(t *testing.T) {
	key := []byte("allskjdlafkjzcxoivj")

	var s Session
	s.UserID = "foo"
	s.Sign(key)

	if !s.Verify(key) {
		t.Error("could not verify a valid session")
	}

	s.Mac = []byte("zaoijzv.cz;aspoa") // this hopefully isn't a valid signature...
	if s.Verify(key) {
		t.Error("incorrectly verified an invalid session")
	}
}
