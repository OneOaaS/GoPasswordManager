package main

import "time"

// Session stores everything that needs to be persisted accross sessions.
type Session struct {
	UserID string    `json:"userID"`
	Time   time.Time `json:"time"`
}
