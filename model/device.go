package model

// Device identifies the attributes for the network device
type Device struct {
	Hostname   string     `json:"hostname"`
	Serial     string     `json:"serial"`
	Fixedip    string     `json:"fixedIp"`
	Image      Image      `json:"image"`
	Config     Config     `json:"config"`
	DeviceType DeviceType `json:"deviceType"`
}

// DeviceType identifies if the device is NX or XR type
type DeviceType struct {
	Name string `json:"name"`
}
