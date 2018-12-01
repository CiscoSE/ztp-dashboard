package controller

import (
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var basePath = os.Getenv("GOPATH") + "/src/github.com/CiscoSE/ztp-dashboard"

var (
	indexController index
	homeController  home
	deviceCtl       deviceController
	dhcpController  DhcpController
	settingsCtl     settingsController
	deviceDetailCtl deviceDetailController
)

func Startup(templates map[string]*template.Template, r *mux.Router) {

	// Handle web server

	indexController.template = templates["index.html"]
	indexController.registerRoutes(r)

	homeController.template = templates["home.html"]
	homeController.registerRoutes(r)

	deviceDetailCtl.template = templates["deviceDetail.html"]
	deviceDetailCtl.registerRoutes(r)

	deviceCtl.template = templates["devices.html"]
	deviceCtl.registerRoutes(r)

	settingsCtl.template = templates["settings.html"]
	settingsCtl.registerRoutes(r)

	r.PathPrefix("/assets/").Handler(http.FileServer(http.Dir(basePath + "/public")))

	// Handle DHCP Config files
	dhcpController.DhcpTemplate = basePath + "/dhcpConfTemplates/dhcpd.conf"
	dhcpController.DhcpHostsTemplate = basePath + "/dhcpConfTemplates/dhcpHost.conf"
	dhcpController.Dhcp6Template = basePath + "/dhcpConfTemplates/dhcpd6.conf"
	dhcpController.Dhcp6HostsTemplate = basePath + "/dhcpConfTemplates/dhcp6Host.conf"
	go dhcpController.GenerateConfigFiles()

}
