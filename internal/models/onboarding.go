package models

type OnboardingEvent struct {
	Peer      Peer
	OrgName   string
	Onboarded bool
}

type OnboardingRequest struct {
	OrgName  string `json:"org_name"`
	Passcode string `json:"passcode"`

	DeviceName string `json:"device_name"`
	DeviceIP   string `json:"device_ip"`

	ControlPort  int `json:"control_port"`
	TransferPort int `json:"transfer_port"`
}

type OnboardingResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`

	NodeID    string `json:"node_id"`
	SessionID string `json:"session_id"`

	HeartbeatInterval int `json:"heartbeat_interval"`
}
