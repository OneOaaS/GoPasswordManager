package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"time"
)

// Session stores everything that needs to be persisted accross sessions.
type Session struct {
	UserID string    `json:"userID"`
	Time   time.Time `json:"time"`
	Mac    []byte    `json:"mac"`
}

func (s *Session) macData() []byte {
	return []byte(fmt.Sprintf("%s::%s", s.UserID, s.Time.Format(time.RFC3339Nano)))
}

func (s *Session) mac(key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(s.macData())
	return mac.Sum(nil)
}

// Sign signs the relevant fields of the Session and updates the session with
// the signature.
func (s *Session) Sign(key []byte) {
	s.Mac = s.mac(key)
}

// Verify verifies the Session's signature.
func (s *Session) Verify(key []byte) bool {
	m := s.mac(key)
	return hmac.Equal(s.Mac, m)
}
