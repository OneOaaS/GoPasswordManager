package main

type Config struct {
	CookieSecret []byte
	CookieName   string
	Dev          bool
	DB           struct {
		Driver string
		DSN    string
	}
}
