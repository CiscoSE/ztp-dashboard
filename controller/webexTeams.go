package controller

import (
	"bytes"
	"crypto/tls"
	"html/template"
	"net/http"
	"os"

	"github.com/CiscoSE/ztp-dashboard/model"
	"github.com/globalsign/mgo/bson"
)

// WebexTeamsController encapsulates all request to Cisco Webex teams
type WebexTeamsController struct {
	BaseURL string
	db      dbController
}

// messageTemplateParams encapsulates all variables for the message template
type messageTemplateParams struct {
	Message string
	RoomID  string
}

// SendMessage uses webex teams to send a message to the configured room ID
func (w WebexTeamsController) SendMessage(message string) {
	// Get room ID from database

	// Open database
	session, err := w.db.OpenSession()
	if err != nil {
		go CustomLog("SendMessage (open database): "+err.Error(), ErrorSeverity)
		return
	}
	defer session.Close()

	dbCollection := session.DB("ztpDashboard").C("settings")

	settingsCollection := dbCollection.Find(bson.M{})
	count, err := settingsCollection.Count()
	if err != nil {
		go CustomLog("SendMessage (read database): "+err.Error(), ErrorSeverity)
		return
	}
	if count == 0 {
		go CustomLog("No settings in database, have you configure the settings? ", DebugSeverity)
		return
	}

	var settings model.Settings
	err = settingsCollection.One(&settings)
	if err != nil {
		go CustomLog("SendMessage (read database): "+err.Error(), ErrorSeverity)
		return
	}

	// check that there is a valid roomID and token
	if settings.WebexTeamsRoomID == "" {
		go CustomLog("makeCall (Webex Room ID) No webex teams room ID configured. Cannot send message", DebugSeverity)
		return
	}
	if os.Getenv("WEBEX_BOT_TOKEN") == "" {
		go CustomLog("makeCall (Webex Token): No webex teams token configured. Cannot send message", DebugSeverity)
		return
	}
	// Create request

	// 'buf' is an io.Writter to capture the template execution output
	payload := new(bytes.Buffer)
	templateParams := &messageTemplateParams{
		Message: message,
		RoomID:  settings.WebexTeamsRoomID,
	}
	// Read the json template file
	t, err := template.ParseFiles(basePath + "/jsonTemplates/addWebexTeamsMessage.json")
	if err != nil {
		go CustomLog("SendMessage (parse template): "+err.Error(), ErrorSeverity)
		return
	}

	err = t.Execute(payload, templateParams)
	if err != nil {
		go CustomLog("SendMessage (execute template): "+err.Error(), ErrorSeverity)
		return
	}
	resp, err := w.makeCall("POST", "/v1/messages", payload.Bytes())
	if err != nil {
		go CustomLog("SendMessage (make call): "+err.Error(), ErrorSeverity)
		return
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		go CustomLog("SendMessage (make call): webex teams returned status code "+string(resp.StatusCode), ErrorSeverity)
	}
}

// Single point to make calls to webex teams
func (w WebexTeamsController) makeCall(method string, url string, payload []byte) (*http.Response, error) {

	// Validate URL format
	botToken := os.Getenv("WEBEX_BOT_TOKEN")
	callURL := w.BaseURL + url

	go CustomLog("Making call to WebexTeams -> "+method+": "+callURL, DebugSeverity)
	go CustomLog("Payload -> "+string(payload[:]), DebugSeverity)

	// Create request
	req, _ := http.NewRequest(method, callURL, bytes.NewBuffer(payload))

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+botToken)

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
