package model

// Config represents the day0 configuration file for a device.
type Config struct {
	Name          string     `json:"name"`
	DeviceType    DeviceType `json:"deviceType"`
	Configuration string     `json:"configuration"`
	Locationurl   string     `json:"locationUrl"`
}
