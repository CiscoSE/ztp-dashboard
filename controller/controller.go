package controller

import (
	"html/template"
	"log"
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
	configsCtl      configController
	scriptCtl       ScriptController
	imagesCtl       imageController
)

// Startup associates controllers with templates and routes
func Startup(templates map[string]*template.Template, r *mux.Router) {

	// Handle web server mappings

	// Home & Index
	indexController.template = templates["index.html"]
	indexController.registerRoutes(r)

	homeController.template = templates["home.html"]
	homeController.registerRoutes(r)

	// Devices
	deviceCtl.deviceListTemplate = templates["devices.html"]
	deviceCtl.deviceDetailTemplate = templates["deviceDetail.html"]
	deviceCtl.registerRoutes(r)

	// Create device types if not present
	deviceCtl.checkDeviceTypes()

	// Settings
	settingsCtl.template = templates["settings.html"]
	settingsCtl.registerRoutes(r)

	// Configurations
	configsCtl.configListTemplate = templates["configs.html"]
	configsCtl.configDetailTemplate = templates["configDetail.html"]
	configsCtl.registerRoutes(r)

	// Images
	imagesCtl.imageListTemplate = templates["images.html"]
	imagesCtl.imageDetailTemplate = templates["imageDetail.html"]
	imagesCtl.registerRoutes(r)

	// Public assets and configs
	r.PathPrefix("/assets/").Handler(http.FileServer(http.Dir(basePath + "/public")))
	r.PathPrefix("/scripts/").Handler(http.FileServer(http.Dir(basePath + "/public")))
	r.PathPrefix("/configs/").Handler(http.FileServer(http.Dir(basePath + "/public")))
	r.PathPrefix("/images/").Handler(http.FileServer(http.Dir(basePath + "/public")))

	// Handle DHCP Config files
	dhcpController.DhcpTemplate = basePath + "/dhcpConfTemplates/dhcpd.conf"
	dhcpController.DhcpXRHostsTemplate = basePath + "/dhcpConfTemplates/dhcpXRHost.conf"
	dhcpController.DhcpNXHostsTemplate = basePath + "/dhcpConfTemplates/dhcpNXHost.conf"
	dhcpController.Dhcp6Template = basePath + "/dhcpConfTemplates/dhcpd6.conf"
	dhcpController.Dhcp6XRHostsTemplate = basePath + "/dhcpConfTemplates/dhcp6XRHost.conf"
	dhcpController.Dhcp6NXHostsTemplate = basePath + "/dhcpConfTemplates/dhcp6NXHost.conf"

	// Handle Day 0 script files
	dhcpController.scriptCtl.xrShellTemplate = basePath + "/shellTemplates/ztpXR.sh"
	dhcpController.scriptCtl.nxPythonTemplate = basePath + "/pythonTemplates/poapNX.py"

	// Make sure that needed directories exists
	CreateDirIfNotExist(basePath + "/public/configs")
	CreateDirIfNotExist(basePath + "/public/images")
	CreateDirIfNotExist(basePath + "/public/scripts")
	go dhcpController.GenerateConfigFiles()

}

// CreateDirIfNotExist creates directories if not present
func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatal(err.Error())
			panic(err)
		}
	}
}
