package models

import "time"

type Peer struct {
	DeviceID  string `json:"device_id"`
	SessionID string `json:"session_id"`

	DeviceName string `json:"device_name"`
	DeviceIP   string `json:"device_ip"`

	ControlPort  int `json:"control_port"`
	TransferPort int `json:"transfer_port"`

	JoinedAt time.Time `json:"joined_at"`
	LastSeen time.Time `json:"last_seen"`

	Status PeerStatus `json:"status"`
}

type PeerStatus string

const (
	PeerStatusOnline  PeerStatus = "online"
	PeerStatusOffline PeerStatus = "offline"
)
