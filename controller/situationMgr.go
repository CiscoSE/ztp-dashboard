package controller

import (
	"bytes"
	"crypto/tls"
	"html/template"
	"log"
	"net/http"

	"github.com/CiscoSE/ztp-dashboard/model"
	"github.com/globalsign/mgo/bson"
)

// SituationMgrController encapsulates all request to Cisco situation manager
type SituationMgrController struct {
	db           dbController
	InterfaceCtl interfaceController
}

// eventTemplateParams encapsulates all variables for the event template
type eventTemplateParams struct {
	Description string
	Source      string
}

// SendEvent uses register a new event into situation manager with the given description
// All logs are printed directly since the debug handler uses this method
func (s SituationMgrController) SendEvent(pDescription string) {
	// Create request

	ip, err := s.InterfaceCtl.GetFirstIPv4()
	if err != nil {
		log.Print(ErrorSeverity + " (SendEvent): Cannot retrieve IPv4 address. Trying with IPv6: " + err.Error())
	}
	if ip == "" {

		log.Print(ErrorSeverity + " (SendEvent): Empty IPv4 address returned. Trying with IPv6 " + err.Error())
		ip, err = s.InterfaceCtl.GetFirstIPv6()
		if err != nil {
			log.Print(ErrorSeverity + " (SendEvent): Cannot retrieve IPv6 address. No IP addresses found. " + err.Error())
			return
		}
	}

	// 'buf' is an io.Writter to capture the template execution output
	payload := new(bytes.Buffer)
	templateParams := &eventTemplateParams{
		Description: pDescription,
		Source:      ip,
	}
	// Read the json template file
	t, err := template.ParseFiles(basePath + "/jsonTemplates/addSituationMgrEvent.json")
	if err != nil {
		log.Print(ErrorSeverity + " (SendEvent): Cannot read event template. " + err.Error())
		return
	}

	err = t.Execute(payload, templateParams)
	if err != nil {
		log.Print(ErrorSeverity + " (SendEvent): Cannot create event payload. " + err.Error())
		return
	}
	s.makeCall("POST", "", payload.Bytes())
}

// Single point to make calls to webex teams
func (s SituationMgrController) makeCall(method string, url string, payload []byte) (*http.Response, error) {
	// Get Situation Manager URL from database

	// Open database
	session, err := s.db.OpenSession()
	if err != nil {
		log.Print(ErrorSeverity + " (makeCall): Cannot open database. " + err.Error())
		return nil, err
	}
	defer session.Close()

	dbCollection := session.DB("ztpDashboard").C("settings")

	settingsCollection := dbCollection.Find(bson.M{})
	count, err := settingsCollection.Count()
	if err != nil {
		log.Print(ErrorSeverity + " (makeCall): Cannot read database. " + err.Error())
		return nil, err
	}
	if count == 0 {
		CustomLog("(makeCall): No settings in database, have you configure the settings?", DebugSeverity)
		return nil, nil
	}

	var settings model.Settings
	err = settingsCollection.One(&settings)
	if err != nil {
		log.Print(ErrorSeverity + " (makeCall): Cannot parse settings from database:" + err.Error())
		return nil, err
	}

	// Check if settings have been correctly configured
	if settings.SituationMgrURL == "" {
		CustomLog("makeCall (SituationMgrURL): Cannot send event, no SitMgr URL configured", DebugSeverity)
		return nil, nil
	}
	// Send the request
	callURL := settings.SituationMgrURL

	CustomLog("Making call to situation manager -> "+method+": "+callURL, DebugSeverity)
	CustomLog("Payload -> "+string(payload[:]), DebugSeverity)

	// Create request
	req, _ := http.NewRequest(method, callURL, bytes.NewBuffer(payload))

	// Add headers
	req.Header.Set("Content-Type", "application/json")

	// Create transport that allows https connections with self signed  certificates
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Set up client
	client := &http.Client{Transport: tr}

	// Do request
	resp, err := client.Do(req)

	// Return the results
	return resp, err
}
