package models

import "time"

type ClusterSession struct {
	SessionID string
	NodeID    string

	BootstrapAddress string

	Connected     bool
	LastHeartbeat time.Time
}
