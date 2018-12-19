package controller

import (
	"os"

	"github.com/globalsign/mgo"
)

type dbController struct {
}

func (d dbController) OpenSession() (*mgo.Session, error) {
	// Open database
	session, err := mgo.Dial(os.Getenv("DB_URI"))
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Monotonic, true)
	return session, err
}
