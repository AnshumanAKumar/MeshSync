package models

import "time"

type BootstrapNode struct {
	OrgName     string
	BootstrapIP string
	ControlPort int
	Connected   bool
	LastSeen    time.Time
}
