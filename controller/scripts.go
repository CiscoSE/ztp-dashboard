package controller

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/CiscoSE/ztp-dashboard/model"
)

// ScriptController generates shell script files for XR and NX
type ScriptController struct {
	xrShellTemplate  string
	nxPythonTemplate string
	interfacesCtl    interfaceController
}

type xrZtpConfig struct {
	ServerURL string
	ConfigURL string
}

type nxPoapConfig struct {
	ServerIP   string
	ConfigName string
	ImageName  string
}

// GenerateNXPoapScript creates the day0 script for nexus devices
func (s ScriptController) GenerateNXPoapScript(device model.Device, isIPv6 bool) {
	var err error
	var serverIP string

	if isIPv6 {
		serverIP, err = s.interfacesCtl.GetFirstIPv6()
	} else {
		serverIP, err = s.interfacesCtl.GetFirstIPv4()
	}
	if err != nil {
		go CustomLog("GenerateNXPoapScript (get ServerIP): "+err.Error(), ErrorSeverity)
		return
	}
	if serverIP == "" {
		go CustomLog("GenerateNXPoapScript (No local server IP. Cannot build POAP script)", ErrorSeverity)
		return
	}
	poapConfig := &nxPoapConfig{
		ServerIP:   serverIP,
		ImageName:  device.Image.Name,
		ConfigName: device.Config.Name + ".conf",
	}
	t, err := template.ParseFiles(s.nxPythonTemplate)

	if err != nil {
		go CustomLog("GenerateNXPoapScript (Parse nxPythonTemplate): "+err.Error(), ErrorSeverity)
	}
	buf1 := new(bytes.Buffer)
	err = t.Execute(buf1, poapConfig)
	if err != nil {
		go CustomLog("GenerateNXPoapScript (Execute nxPythonTemplate): "+err.Error(), ErrorSeverity)
	}
	result := buf1.String()
	err = ioutil.WriteFile(basePath+"/public/scripts/"+device.Serial+".py", []byte(strings.Replace(result, "&#34;", "\"", -1)), 0644)
	if err != nil {
		go CustomLog("GenerateNXPoapScript (Write"+device.Serial+".py into disk): "+err.Error(), ErrorSeverity)
	}
}

// GenerateXRZtpScript creates the shell script to be used by XR devices
func (s ScriptController) GenerateXRZtpScript(device model.Device, isIPv6 bool) {
	var err error
	var serverIP string

	if isIPv6 {
		serverIP, err = s.interfacesCtl.GetFirstIPv6()
	} else {
		serverIP, err = s.interfacesCtl.GetFirstIPv4()
	}

	if err != nil {
		go CustomLog("GenerateXRZtpScript (get ServerIP): "+err.Error(), ErrorSeverity)
		return
	}
	if serverIP == "" {
		go CustomLog("GenerateXRZtpScript (No local server IP. Cannot build POAP script)", ErrorSeverity)
		return
	}
	shellConfig := &xrZtpConfig{
		ServerURL: "http://" + serverIP + ":" + os.Getenv("APP_WEB_PORT"),
		ConfigURL: device.Config.Locationurl,
	}

	t, err := template.ParseFiles(s.xrShellTemplate)

	if err != nil {
		go CustomLog("GenerateXRZtpScript (Parse xrShell Template): "+err.Error(), ErrorSeverity)
		return
	}
	buf1 := new(bytes.Buffer)
	err = t.Execute(buf1, shellConfig)
	if err != nil {
		go CustomLog("GenerateXRZtpScript (Execute xrShell Template): "+err.Error(), ErrorSeverity)
		return
	}
	result := buf1.String()
	err = ioutil.WriteFile(basePath+"/public/scripts/"+device.Serial+".sh", []byte(strings.Replace(result, "&#34;", "\"", -1)), 0644)
	if err != nil {
		go CustomLog("GenerateXRZtpScript (write "+device.Serial+".sh into disk): "+err.Error(), ErrorSeverity)
	}
}

// RemoveAllScripts deletes all scripts from the web server
func (s ScriptController) RemoveAllScripts() error {
	d, err := os.Open(basePath + "/public/scripts/")
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(basePath+"/public/scripts/", name))
		if err != nil {
			return err
		}
	}
	return nil
}
