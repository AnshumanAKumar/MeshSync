package discovery

type Advertisement struct {
	Type        string `json:"type"`
	OrgName     string `json:"org_name"`
	Passcode    string `json:"passcode"`
	BootstrapIP string `json:"bootstrap_ip"`
	ControlPort int    `json:"control_port"`
}
