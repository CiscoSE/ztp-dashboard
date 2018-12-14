package model

// Image identifies the operating system to be installed on devices.
type Image struct {
	Name        string     `json:"name"`
	DeviceType  DeviceType `json:"deviceType"`
	Locationurl string     `json:"locationUrl"`
}
