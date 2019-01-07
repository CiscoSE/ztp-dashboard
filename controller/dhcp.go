package controller

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/CiscoSE/ztp-dashboard/model"
	"github.com/asaskevich/govalidator"
)

type DhcpController struct {
	db                   dbController
	DhcpTemplate         string
	Dhcp6Template        string
	DhcpXRHostsTemplate  string
	DhcpNXHostsTemplate  string
	Dhcp6XRHostsTemplate string
	Dhcp6NXHostsTemplate string
	interfacesCtl        interfaceController
}

type DhcpConfig struct {
	ServerIP        string
	DhcpDomain      string
	DhcpNameServers string
	DhcpSubnet      string
	DhcpNetmask     string
	Hosts           string
}

type DhcpHostConfig struct {
	HostName     string
	ClientID     string
	FixedAddress string
	FQDN         string
	BootFile     string
	ScriptFile   string
}

func (d DhcpController) GenerateConfigFiles() {
	var devices []model.Device
	dhcpHosts := ""
	dhcp6Hosts := ""

	// Open database
	session, err := d.db.OpenSession()
	if err != nil {
		go CustomLog("GenerateConfigFiles (open database): "+err.Error(), ErrorSeverity)
		return
	}
	defer session.Close()
	dbCollection := session.DB("ztpDashboard").C("device")
	err = dbCollection.Find(nil).All(&devices)
	if err != nil {
		go CustomLog("GenerateConfigFiles (read database): "+err.Error(), ErrorSeverity)
	}

	localServerIPv4, err := d.interfacesCtl.GetFirstIPv4()
	if err != nil {
		go CustomLog("GenerateConfigFiles (get IPv4 address): "+err.Error(), ErrorSeverity)
	}
	if localServerIPv4 == "" {
		go CustomLog("GenerateConfigFiles (IPv4 address empty)", ErrorSeverity)
	}
	localServerIPv6, err := d.interfacesCtl.GetFirstIPv6()
	if err != nil {
		go CustomLog("GenerateConfigFiles (get IPv6 address): "+err.Error(), ErrorSeverity)
	}
	if localServerIPv4 == "" {
		go CustomLog("GenerateConfigFiles (IPv6 address empty): "+err.Error(), ErrorSeverity)
	}

	err = scriptCtl.RemoveAllScripts()
	if err != nil {
		go CustomLog("GenerateConfigFiles (clean script directory): "+err.Error(), ErrorSeverity)
	}
	for _, item := range devices {
		if item.DeviceType.Name == "iOS-XR" {
			scriptCtl.GenerateXRZtpScript(item, govalidator.IsIPv6(item.Fixedip))
		} else if item.DeviceType.Name == "NX-OS" {
			scriptCtl.GenerateNXPoapScript(item, govalidator.IsIPv6(item.Fixedip))
		}
		var hostTemplate string
		dhcpHost := &DhcpHostConfig{}

		if govalidator.IsIPv6(item.Fixedip) {
			clientID := "00:02:00:00:00:09:"
			for _, element := range item.Serial {
				h := fmt.Sprintf("%X", element)
				clientID += h + ":"
			}
			clientID += "00"
			if item.DeviceType.Name == "iOS-XR" {
				hostTemplate = d.Dhcp6XRHostsTemplate
				dhcpHost = &DhcpHostConfig{
					HostName:     item.Hostname,
					ClientID:     clientID,
					FQDN:         item.Hostname + "." + os.Getenv("DHCP_DOMAIN"),
					BootFile:     "http://[" + localServerIPv6 + "]:" + os.Getenv("APP_WEB_PORT") + item.Image.Locationurl,
					ScriptFile:   "http://[" + localServerIPv6 + "]:" + os.Getenv("APP_WEB_PORT") + item.Config.Locationurl,
					FixedAddress: item.Fixedip,
				}
			} else if item.DeviceType.Name == "NX-OS" {
				hostTemplate = d.Dhcp6NXHostsTemplate
				dhcpHost = &DhcpHostConfig{
					HostName:     item.Hostname,
					ClientID:     clientID,
					ScriptFile:   "/tftboot/public/scripts/" + item.Serial + ".py",
					FixedAddress: item.Fixedip,
				}
			}
		} else {
			clientID := item.Serial
			if item.DeviceType.Name == "iOS-XR" {
				hostTemplate = d.DhcpXRHostsTemplate
				dhcpHost = &DhcpHostConfig{
					HostName:     item.Hostname,
					ClientID:     clientID,
					FQDN:         item.Hostname + "." + os.Getenv("DHCP_DOMAIN"),
					BootFile:     "http://" + localServerIPv4 + ":" + os.Getenv("APP_WEB_PORT") + item.Image.Locationurl,
					ScriptFile:   "http://" + localServerIPv4 + ":" + os.Getenv("APP_WEB_PORT") + "/scripts/" + item.Serial + ".sh",
					FixedAddress: item.Fixedip,
				}
			} else if item.DeviceType.Name == "NX-OS" {
				hostTemplate = d.DhcpNXHostsTemplate
				dhcpHost = &DhcpHostConfig{
					HostName:     item.Hostname,
					ClientID:     clientID,
					ScriptFile:   "public/scripts/" + item.Serial + ".py",
					FixedAddress: item.Fixedip,
				}
			}
		}

		t, err := template.ParseFiles(hostTemplate)

		if err != nil {
			go CustomLog("GenerateConfigFiles (parse hostTemplate): "+err.Error(), ErrorSeverity)

		}
		buf1 := new(bytes.Buffer)
		err = t.Execute(buf1, dhcpHost)
		if err != nil {
			go CustomLog("GenerateConfigFiles (execute hostTemplate): "+err.Error(), ErrorSeverity)
		}
		if govalidator.IsIPv6(item.Fixedip) {
			dhcp6Hosts += buf1.String()
		} else {
			dhcpHosts += buf1.String()
		}

	}

	// DHCPv4
	dhcpConfig := &DhcpConfig{
		DhcpNameServers: os.Getenv("DHCP_NAMESERVERS"),
		DhcpDomain:      os.Getenv("DHCP_DOMAIN"),
		DhcpSubnet:      os.Getenv("DHCP_SUBNET"),
		DhcpNetmask:     os.Getenv("DHCP_SUBNET_NETMASK"),
		ServerIP:        localServerIPv4,
		Hosts:           dhcpHosts,
	}
	t, err := template.ParseFiles(d.DhcpTemplate)
	if err != nil {
		go CustomLog("GenerateConfigFiles (parse dhcpTemplate): "+err.Error(), ErrorSeverity)
	}
	buf1 := new(bytes.Buffer)
	err = t.Execute(buf1, dhcpConfig)
	if err != nil {
		go CustomLog("GenerateConfigFiles (execute dhcpTemplate): "+err.Error(), ErrorSeverity)
	}
	result := buf1.String()
	err = ioutil.WriteFile(os.Getenv("DHCP_CONFIG_PATH"), []byte(strings.Replace(result, "&#34;", "\"", -1)), 0644)
	if err != nil {
		go CustomLog("GenerateConfigFiles (write dhcp.conf file): "+err.Error(), ErrorSeverity)
	}

	go CustomLog("Restarting DHCPv4 service using: "+os.Getenv("DHCP_SERVICE_RESTART_CMD"), DebugSeverity)

	_, err = exec.Command("bash", "-c", os.Getenv("DHCP_SERVICE_RESTART_CMD")).Output()
	if err != nil {
		go CustomLog("GenerateConfigFiles (restart DHCP service): "+err.Error(), ErrorSeverity)
	}

	// DHCPv6
	dhcp6Config := &DhcpConfig{
		DhcpNameServers: os.Getenv("DHCP6_NAMESERVERS"),
		DhcpDomain:      os.Getenv("DHCP6_DOMAIN"),
		DhcpSubnet:      os.Getenv("DHCP6_SUBNET"),
		DhcpNetmask:     os.Getenv("DHCP6_SUBNET_NETMASK"),
		ServerIP:        localServerIPv6,
		Hosts:           dhcp6Hosts,
	}
	t, err = template.ParseFiles(d.Dhcp6Template)
	if err != nil {
		go CustomLog("GenerateConfigFiles (Parse Dhcp6 Template): "+err.Error(), ErrorSeverity)

	}
	buf1 = new(bytes.Buffer)
	err = t.Execute(buf1, dhcp6Config)
	if err != nil {
		go CustomLog("GenerateConfigFiles (Execute Dhcp6 Template): "+err.Error(), ErrorSeverity)
	}
	result = buf1.String()
	err = ioutil.WriteFile(os.Getenv("DHCP6_CONFIG_PATH"), []byte(strings.Replace(result, "&#34;", "\"", -1)), 0644)
	if err != nil {
		go CustomLog("GenerateConfigFiles (wrote dhcp6 config file): "+err.Error(), ErrorSeverity)
	}
	go CustomLog("Restarting DHCPv6 service using:"+os.Getenv("DHCP6_SERVICE_RESTART_CMD"), DebugSeverity)

	_, err = exec.Command("bash", "-c", os.Getenv("DHCP6_SERVICE_RESTART_CMD")).Output()
	if err != nil {
		go CustomLog("GenerateConfigFiles (restart DHCP6 service): "+err.Error(), ErrorSeverity)
	}

}
