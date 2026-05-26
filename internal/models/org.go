package models

import "time"

type Org struct {
	ID        string
	Name      string
	Passcode  string
	ExpiresAt time.Time
}
