package controller

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/CiscoSE/ztp-dashboard/model"
	"github.com/asaskevich/govalidator"
)

type DhcpController struct {
	db                 dbController
	DhcpTemplate       string
	DhcpHostsTemplate  string
	Dhcp6Template      string
	Dhcp6HostsTemplate string
}

type DhcpConfig struct {
	DhcpDomain      string
	DhcpNameServers string
	DhcpSubnet      string
	DhcpNetmask     string
	Hosts           string
}

type DhcpHostConfig struct {
	HostName     string
	ClientId     string
	FixedAddress string
	FQDN         string
	BootFile     string
	ConfigFile   string
}

func (d DhcpController) GenerateConfigFiles() {
	var devices []model.Device
	dhcpHosts := ""
	dhcp6Hosts := ""

	// Open database
	session, err := d.db.OpenSession()
	if err != nil {
		log.Fatalf("Cannot open database:" + err.Error() + "\n")
		return
	}
	defer session.Close()
	dbCollection := session.DB("ztpDashboard").C("devices")
	err = dbCollection.Find(nil).All(&devices)
	if err != nil {
		log.Fatalf("Cannot read ztpDashboard database:" + err.Error() + "\n")
	}

	for _, item := range devices {
		clientId := item.Serial
		hostTemplate := d.DhcpHostsTemplate

		if govalidator.IsIPv6(item.FixedIp) {
			clientId = "00:02:00:00:00:09:"
			for _, element := range item.Serial {
				h := fmt.Sprintf("%X", element)
				clientId += h + ":"
			}
			clientId += "00"
			hostTemplate = d.Dhcp6HostsTemplate
		}
		dhcpHost := &DhcpHostConfig{
			HostName:     item.Host,
			ClientId:     clientId,
			FQDN:         item.Host + "." + os.Getenv("DHCP_DOMAIN"),
			BootFile:     item.IpxeBootFileUrl,
			ConfigFile:   item.ConfigBootFileUrl,
			FixedAddress: item.FixedIp,
		}

		t, err := template.ParseFiles(hostTemplate)

		if err != nil {
			log.Printf("Could not get Templated parsed %v", err)

		}
		buf1 := new(bytes.Buffer)
		err = t.Execute(buf1, dhcpHost)
		if err != nil {
			log.Printf("Could not execute dhcp hosts config template: %v", err)

		}
		if govalidator.IsIPv6(item.FixedIp) {
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
		Hosts:           dhcpHosts,
	}
	t, err := template.ParseFiles(d.DhcpTemplate)
	if err != nil {
		log.Printf("Could not get Templated parsed %v", err)

	}
	buf1 := new(bytes.Buffer)
	err = t.Execute(buf1, dhcpConfig)
	if err != nil {
		log.Printf("Could not execute dhcp config template: %v", err)

	}
	result := buf1.String()
	err = ioutil.WriteFile(os.Getenv("DHCP_CONFIG_PATH"), []byte(strings.Replace(result, "&#34;", "\"", -1)), 0644)
	if err != nil {
		log.Printf("Could not write dhcp config file: %v", err)
	}
	log.Printf("Restarting DHCPv4 service using: %v", os.Getenv("DHCP_SERVICE_RESTART_CMD"))
	out, err := exec.Command("bash", "-c", os.Getenv("DHCP_SERVICE_RESTART_CMD")).Output()
	if err != nil {
		log.Printf("Could not restart dhcp service: %v", err)
	}
	fmt.Printf("%s", out)

	// DHCPv6
	dhcp6Config := &DhcpConfig{
		DhcpNameServers: os.Getenv("DHCP6_NAMESERVERS"),
		DhcpDomain:      os.Getenv("DHCP6_DOMAIN"),
		DhcpSubnet:      os.Getenv("DHCP6_SUBNET"),
		DhcpNetmask:     os.Getenv("DHCP6_SUBNET_NETMASK"),
		Hosts:           dhcp6Hosts,
	}
	t, err = template.ParseFiles(d.Dhcp6Template)
	if err != nil {
		log.Printf("Could not get dhcp6 templated parsed %v", err)

	}
	buf1 = new(bytes.Buffer)
	err = t.Execute(buf1, dhcp6Config)
	if err != nil {
		log.Printf("Could not execute dhcp6 config template: %v", err)

	}
	result = buf1.String()
	err = ioutil.WriteFile(os.Getenv("DHCP6_CONFIG_PATH"), []byte(strings.Replace(result, "&#34;", "\"", -1)), 0644)
	if err != nil {
		log.Printf("Could not write dhcp6 config file: %v", err)
	}

	log.Printf("Restarting DHCPv6 service using: %v", os.Getenv("DHCP6_SERVICE_RESTART_CMD"))
	out, err = exec.Command("bash", "-c", os.Getenv("DHCP6_SERVICE_RESTART_CMD")).Output()
	if err != nil {
		log.Printf("Could not restart dhcp6 service: %v", err)
	}
	fmt.Printf("%s", out)
}
