package service

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	group "github.com/energieip/common-group-go/pkg/groupmodel"
	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/common-sensor-go/pkg/driversensor"
	pkg "github.com/energieip/common-service-go/pkg/service"
	"github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/common-tools-go/pkg/tools"
	"github.com/energieip/swh200-coreservice-go/internal/core"
	"github.com/energieip/swh200-coreservice-go/internal/database"
	"github.com/energieip/swh200-coreservice-go/internal/network"
	"github.com/energieip/swh200-coreservice-go/pkg/config"
	"github.com/romana/rlog"
)

const (
	ActionReload = "ReloadConfig"
	ActionSetup  = "Setup"
	ActionDump   = "DumpStatus"
	ActionRemove = "remove"

	UrlStatus = "status/dump"
	UrlHello  = "setup/hello"
)

//CoreService content
type CoreService struct {
	server                network.ServerNetwork //Remote server
	local                 network.LocalNetwork  //local broker for drivers and services
	db                    database.Database
	mac                   string //Switch mac address
	events                chan string
	timerDump             time.Duration //in seconds
	ip                    string
	isConfigured          bool
	groups                map[int]group.GroupRuntime
	services              map[string]pkg.Service
	lastSystemUpgradeDate string
}

//Initialize service
func (s *CoreService) Initialize(confFile string) error {
	s.groups = make(map[int]group.GroupRuntime)
	s.services = make(map[string]pkg.Service)
	s.lastSystemUpgradeDate = core.GetLastSystemUpgradeDate()
	hostname, err := os.Hostname()
	if err != nil {
		rlog.Error("Cannot read hostname " + err.Error())
		return err
	}
	clientID := "Switch" + hostname
	s.mac, s.ip = tools.GetNetworkInfo()
	s.mac = strings.ToUpper(strings.Replace(s.mac, ":", "", -1))
	s.events = make(chan string)

	conf, err := config.ReadConfig(confFile)
	if err != nil {
		rlog.Error("Cannot parse configuration file " + err.Error())
		return err
	}
	os.Setenv("RLOG_LOG_LEVEL", *conf.LogLevel)
	os.Setenv("RLOG_LOG_NOTIME", "yes")
	rlog.UpdateEnv()
	rlog.Info("Starting SwitchCore service")

	s.isConfigured = false
	s.timerDump = conf.TimerDump

	db, err := database.ConnectDatabase(conf.DatabaseIP, conf.DatabasePort)
	if err != nil {
		rlog.Error("Cannot connect to database " + err.Error())
		return err
	}
	s.db = *db

	serverNet, err := network.CreateServerNetwork()
	if err != nil {
		rlog.Error("Cannot connect to broker " + conf.ServerIP + " error: " + err.Error())
		return err
	}
	s.server = *serverNet

	driversNet, err := network.CreateLocalNetwork()
	if err != nil {
		rlog.Error("Cannot connect to broker " + conf.ServerIP + " error: " + err.Error())
		return err
	}
	s.local = *driversNet

	err = s.local.LocalConnection(*conf, clientID, s.mac)
	if err != nil {
		rlog.Error("Cannot connect to drivers broker " + conf.DriversIP + " error: " + err.Error())
		return err
	}

	go s.server.RemoteServerConnection(*conf, clientID, s.mac)
	rlog.Info("SwitchCore service started")
	return nil
}

//Stop service
func (s *CoreService) Stop() {
	rlog.Info("Stopping SwitchCore service")
	s.server.Disconnect()
	s.local.Disconnect()
	s.db.Close()
	rlog.Info("SwitchCore service stopped")
}

func (s *CoreService) sendHello() {
	switchDump := deviceswitch.Switch{
		Mac:   s.mac,
		IP:    s.ip,
		Topic: "switch/" + s.mac,
		LastSystemUpgradeDate: s.lastSystemUpgradeDate,
		IsConfigured:          &s.isConfigured,
		Protocol:              "MQTT",
	}
	dump, err := switchDump.ToJSON()
	if err != nil {
		rlog.Errorf("Could not dump switch %v status %v", s.mac, err.Error())
		return
	}

	err = s.server.SendCommand("/read/"+switchDump.Topic+"/"+UrlHello, dump)
	if err != nil {
		rlog.Errorf("Could not send hello to the server %v status %v", s.mac, err.Error())
		return
	}
	rlog.Infof("Hello %v sent to the server", s.mac)
}

func (s *CoreService) sendDump() {
	status := deviceswitch.SwitchStatus{}
	status.Mac = s.mac
	status.Protocol = "MQTT"
	status.IP = s.ip
	status.IsConfigured = &s.isConfigured
	status.LastSystemUpgradeDate = s.lastSystemUpgradeDate
	status.Topic = "switch/" + s.mac
	services := make(map[string]pkg.ServiceStatus)

	for _, c := range s.services {
		component := pkg.ServiceStatus{}
		component.Name = c.Name
		component.PackageName = c.PackageName
		component.Version = c.Version
		status := component.GetServiceStatus()
		component.Status = &status
		services[component.Name] = component
	}

	status.Services = services
	status.Leds = database.GetSwitchLeds(s.db, s.mac)
	status.Sensors = database.GetSwitchSensors(s.db, s.mac)
	status.Groups = database.GetStatusGroup(s.db, s.groups)

	dump, err := status.ToJSON()
	if err != nil {
		rlog.Error("Could not dump switch status ", err.Error())
		return
	}

	err = s.server.SendCommand("/read/"+status.Topic+"/"+UrlStatus, dump)
	if err != nil {
		rlog.Errorf("Could not dump switch %v status %v", s.mac, err.Error())
		return
	}
	rlog.Infof("Status %v sent to the server", s.mac)
}

func (s *CoreService) updateConfiguration(switchConfig deviceswitch.SwitchConfig) {
	for _, led := range switchConfig.LedsSetup {
		url := "/write/switch/led/setup/config"
		ledDump, _ := led.ToJSON()
		s.local.SendCommand(url, ledDump)
	}
	for _, led := range switchConfig.LedsConfig {
		url := "/write/switch/led/update/settings"
		ledDump, _ := led.ToJSON()
		s.local.SendCommand(url, ledDump)
	}

	for _, sensor := range switchConfig.SensorsSetup {
		url := "/write/switch/sensor/setup/config"
		sensorDump, _ := sensor.ToJSON()
		s.local.SendCommand(url, sensorDump)
	}
	for _, sensor := range switchConfig.SensorsConfig {
		url := "/write/switch/sensor/update/settings"
		sensorDump, _ := sensor.ToJSON()
		s.local.SendCommand(url, sensorDump)
	}

	deleteGroups := make(map[int]*group.GroupRuntime)
	updateGroups := make(map[int]group.GroupRuntime)
	addGroups := make(map[int]group.GroupRuntime)
	for grID, gr := range switchConfig.Groups {
		val, ok := s.groups[grID]
		if ok {
			newGr := core.UpdateGroup(s.db, s.mac, val, gr)
			if newGr != nil {
				s.groups[grID] = *newGr
				updateGroups[grID] = *newGr
			} else {
				deleteGroups[grID] = nil
				delete(s.groups, grID)
			}
		} else {
			newGr := core.CreateGroup(s.db, s.mac, gr)
			if newGr != nil {
				s.groups[grID] = *newGr
				addGroups[grID] = s.groups[grID]
			}
		}
	}
	if len(addGroups) > 0 {
		url := "/write/switch/group/setup/config"
		inrec, err := json.Marshal(addGroups)
		if err == nil {
			dump := string(inrec[:])
			err = s.local.SendCommand(url, dump)
			if err != nil {
				rlog.Error("Cannot send added config to group service " + err.Error())
			} else {
				rlog.Info("Configuration added has been sent to the group service")
			}
		}
	}
	if len(updateGroups) > 0 {
		url := "/write/switch/group/update/settings"
		inrec, err := json.Marshal(updateGroups)
		if err == nil {
			dump := string(inrec[:])
			err = s.local.SendCommand(url, dump)
			if err != nil {
				rlog.Error("Cannot send updated config to group service " + err.Error())
			} else {
				rlog.Info("Configuration updated has been sent to the group service")
			}
		}
	}
	if len(deleteGroups) > 0 {
		url := "/remove/switch/group/update/settings"
		inrec, err := json.Marshal(deleteGroups)
		if err == nil {
			dump := string(inrec[:])
			err = s.local.SendCommand(url, dump)
			if err != nil {
				rlog.Error("Cannot send remove config to group service " + err.Error())
			} else {
				rlog.Info("Configuration remove has been sent to the group service")
			}
		}
	}
}

func (s *CoreService) removeConfiguration(switchConfig deviceswitch.SwitchConfig) {
	for _, group := range switchConfig.Groups {
		dump, _ := group.ToJSON()
		url := "/remove/switch/group/update/settings"
		s.local.SendCommand(url, dump)
	}

	isConfigured := false
	for ledMac := range switchConfig.LedsConfig {
		remove := driverled.LedConf{
			Mac:          ledMac,
			IsConfigured: &isConfigured,
		}
		dump, _ := remove.ToJSON()
		url := "/write/switch/led/update/settings"
		s.local.SendCommand(url, dump)
	}

	for sensorMac := range switchConfig.SensorsConfig {
		remove := driversensor.SensorConf{
			Mac:          sensorMac,
			IsConfigured: &isConfigured,
		}
		dump, _ := remove.ToJSON()
		url := "/write/switch/sensor/update/settings"
		s.local.SendCommand(url, dump)
	}
}

func (s *CoreService) updateServicesConfiguration() {
	rlog.Debug("Update services configuration")
}

func (s *CoreService) cronDump() {
	timerDump := time.NewTicker(s.timerDump * time.Second)
	for {
		select {
		case <-timerDump.C:
			s.events <- ActionDump
		}
	}
}

func (s *CoreService) packagesInstall(switchConfig deviceswitch.SwitchConfig) {
	for name, service := range switchConfig.Services {
		if currentState, ok := s.services[name]; ok {
			if currentState.Version == service.Version {
				rlog.Info("Package " + name + " already in version " + service.Version + " skip it")
				continue
			}
		}

		rlog.Info("Install " + name + " in version " + service.Version)
		service.Install()
		version := pkg.GetPackageVersion(service.PackageName)
		if version != nil {
			service.Version = *version
		}
		s.services[service.Name] = service
	}
}

func (s *CoreService) packagesRemove(switchConfig deviceswitch.SwitchConfig) {
	pkg.RemoveServices(switchConfig.Services)
	for _, service := range switchConfig.Services {
		if _, ok := s.services[service.Name]; ok {
			delete(s.services, service.Name)
		}
	}
}

func (s *CoreService) systemUpdate(switchConfig deviceswitch.SwitchConfig) {
	rlog.Info("Get " + s.lastSystemUpgradeDate + " and expect " + switchConfig.LastSystemUpgradeDate)
	if switchConfig.LastSystemUpgradeDate == s.lastSystemUpgradeDate {
		//system is already up to date
		rlog.Info("System is already up to date")
		return
	}
	core.SystemUpgrade()
}

func (s *CoreService) startServices() {
	pkg.StartServices(s.services)
}

//Run service mainloop
func (s *CoreService) Run() error {
	s.sendHello()
	go s.cronDump()
	for {
		select {
		case serviceEvent := <-s.events:
			switch serviceEvent {
			case ActionDump:
				if s.isConfigured {
					s.sendDump()
				} else {
					s.sendHello()
				}
			}

		case serverEvents := <-s.server.Events:
			for eventType, event := range serverEvents {
				switch eventType {
				case network.EventServerReload:
					if event.IsConfigured != nil {
						s.isConfigured = *event.IsConfigured
					}
					if !s.isConfigured {
						//a reset is performed
						continue
					}
					//In this case reload == setup
					fallthrough

				case network.EventServerSetup:
					s.systemUpdate(event)
					s.packagesInstall(event)
					s.updateConfiguration(event)
					s.updateServicesConfiguration()
					s.startServices()
					s.isConfigured = true

				case network.EventServerRemove:
					s.packagesRemove(event)
					s.removeConfiguration(event)
				}
			}
		}
	}
	return nil
}
