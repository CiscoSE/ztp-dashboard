package model

type Device struct {
	Host              string `json:"host"`
	Serial            string `json:"serial"`
	FixedIp           string `json:"fixedIp"`
	IpxeBootFileUrl   string `json:"ipxeBootfileUrl"`
	ConfigBootFileUrl string `json:"configBootfileUrl"`
}

