package controller

import (
	"net"
	"time"

	"github.com/CiscoSE/ztp-dashboard/model"
	"github.com/globalsign/mgo/bson"

	"github.com/tatsushid/go-fastping"
)

type TestController struct {
	db dbController
}

// TestDevice executes an ping command. Needs root priveledge
func (t TestController) TestDevice(device model.Device) {
	deviceReplied := false
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", device.Fixedip)
	if err != nil {
		go CustomLog("TestDevice (resolve address): "+err.Error(), ErrorSeverity)
	}

	session, err := t.db.OpenSession()
	if err != nil {
		go CustomLog("TestDevice (open database): "+err.Error(), ErrorSeverity)
		return
	}
	defer session.Close()
	dbCollection := session.DB("ztpDashboard").C("device")

	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		go CustomLog("Response from "+device.Fixedip+" received", DebugSeverity)
		// Only do update if device status is different from desired
		if device.Status != "Reachable" {
			go CustomLog("TestDevice: Updating device "+device.Serial+" status to 'Reachable'", DebugSeverity)
			device.Status = "Reachable"
			dbCollection.Update(bson.M{"fixedip": device.Fixedip}, &device)

			// Send notification
			go WebexTeamsCtl.SendMessage("Device " + device.Hostname + " (serial " + device.Serial + ") is reachable. Test succeded")
		}
		//deviceReplied = true
	}
	p.OnIdle = func() {
		if !deviceReplied {
			go CustomLog("TestDevice (Idle): Cannot get a response from "+device.Fixedip, ErrorSeverity)
			// Only do update if device status is different from desired
			if device.Status != "Unreachable" {
				go CustomLog("TestDevice: Updating device "+device.Serial+" status to 'Unreachable'", DebugSeverity)
				device.Status = "Unreachable"
				dbCollection.Update(bson.M{"fixedip": device.Fixedip}, &device)

				// Send notification
				go WebexTeamsCtl.SendMessage("Device " + device.Hostname + " (serial " + device.Serial + ") unreachable. Test failed")
			}
		}
	}
	err = p.Run()
	if err != nil {
		go CustomLog("MakePing (Run): Cannot run ping to "+device.Fixedip+": "+err.Error(), ErrorSeverity)
	}
}
