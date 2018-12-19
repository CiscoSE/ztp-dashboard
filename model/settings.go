package model

// Settings represents the global configurations of the app
type Settings struct {
	SituationMgrURL  string `json:"situationMgrURL"`
	WebexTeamsRoomID string `json:"webexTeamsRoomID"`
}
