package service

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	dl "github.com/energieip/common-led-go/pkg/driverled"
	ds "github.com/energieip/common-sensor-go/pkg/driversensor"
	pkg "github.com/energieip/common-service-go/pkg/service"
	sd "github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/common-tools-go/pkg/tools"
	"github.com/energieip/swh200-coreservice-go/internal/core"
	"github.com/energieip/swh200-coreservice-go/internal/database"
	"github.com/energieip/swh200-coreservice-go/internal/network"
	"github.com/romana/rlog"
)

const (
	ActionReload = "ReloadConfig"
	ActionSetup  = "Setup"
	ActionDump   = "DumpStatus"
	ActionRemove = "remove"

	UrlStatus = "status/dump"
	UrlHello  = "setup/hello"

	TimerDump = 10
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
	groups                map[int]bool
	services              map[string]pkg.Service
	lastSystemUpgradeDate string
	friendlyName          string
}

//Initialize service
func (s *CoreService) Initialize(confFile string) error {
	hostname, _ := os.Hostname()
	clientID := "Switch" + hostname
	s.events = make(chan string)

	conf, err := pkg.ReadServiceConfig(confFile)
	if err != nil {
		rlog.Error("Cannot parse configuration file " + err.Error())
		return err
	}

	mac, ip := tools.GetNetworkInfo()
	s.ip = ip
	s.mac = strings.ToUpper(mac[9:])
	s.groups = make(map[int]bool)
	s.services = make(map[string]pkg.Service)

	os.Setenv("RLOG_LOG_LEVEL", conf.LogLevel)
	os.Setenv("RLOG_LOG_NOTIME", "yes")
	rlog.UpdateEnv()
	rlog.Info("Starting SwitchCore service")

	s.timerDump = TimerDump

	db, err := database.ConnectDatabase(conf.DB.ClientIP, conf.DB.ClientPort)
	if err != nil {
		rlog.Error("Cannot connect to database " + err.Error())
		return err
	}
	s.db = *db

	serverNet, err := network.CreateServerNetwork()
	if err != nil {
		rlog.Error("Cannot connect to broker " + conf.LocalBroker.IP + " error: " + err.Error())
		return err
	}
	s.server = *serverNet

	driversNet, err := network.CreateLocalNetwork()
	if err != nil {
		rlog.Error("Cannot connect to broker " + conf.NetworkBroker.IP + " error: " + err.Error())
		return err
	}
	s.local = *driversNet

	err = s.local.LocalConnection(*conf, clientID, s.mac)
	if err != nil {
		rlog.Error("Cannot connect to drivers broker " + conf.LocalBroker.IP + " error: " + err.Error())
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
	switchDump := sd.Switch{
		Mac:          s.mac,
		IP:           s.ip,
		IsConfigured: &s.isConfigured,
		Protocol:     "MQTTS",
	}
	dump, err := switchDump.ToJSON()
	if err != nil {
		rlog.Errorf("Could not dump switch %v status %v", s.mac, err.Error())
		return
	}

	err = s.server.SendCommand("/read/switch/"+switchDump.Mac+"/"+UrlHello, dump)
	if err != nil {
		rlog.Errorf("Could not send hello to the server %v status %v", s.mac, err.Error())
		return
	}
	rlog.Infof("Hello %v sent to the server", s.mac)
}

func (s *CoreService) sendDump() {
	status := sd.SwitchStatus{}
	status.Mac = s.mac
	status.Protocol = "MQTTS"
	status.IP = s.ip
	status.IsConfigured = &s.isConfigured
	status.FriendlyName = s.friendlyName
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

	err = s.server.SendCommand("/read/switch/"+s.mac+"/"+UrlStatus, dump)
	if err != nil {
		rlog.Errorf("Could not dump switch %v status %v", s.mac, err.Error())
		return
	}
	rlog.Infof("Status %v sent to the server", s.mac)
}

func (s *CoreService) updateConfiguration(switchConfig sd.SwitchConfig) {
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

	for grID := range switchConfig.Groups {
		_, ok := s.groups[grID]
		if !ok {
			s.groups[grID] = true
		}
	}
	if len(switchConfig.Groups) > 0 {
		url := "/write/switch/group/update/settings"
		inrec, err := json.Marshal(switchConfig.Groups)
		if err == nil {
			dump := string(inrec[:])
			s.local.SendCommand(url, dump)
		}
	}
}

func (s *CoreService) removeConfiguration(switchConfig sd.SwitchConfig) {
	for grID, group := range switchConfig.Groups {
		_, ok := s.groups[grID]
		if ok {
			delete(s.groups, grID)
		}
		dump, _ := group.ToJSON()
		url := "/remove/switch/group/update/settings"
		s.local.SendCommand(url, dump)
	}

	isConfigured := false
	for ledMac := range switchConfig.LedsConfig {
		remove := dl.LedConf{
			Mac:          ledMac,
			IsConfigured: &isConfigured,
		}
		dump, _ := remove.ToJSON()
		url := "/write/switch/led/update/settings"
		s.local.SendCommand(url, dump)
	}

	for sensorMac := range switchConfig.SensorsConfig {
		remove := ds.SensorConf{
			Mac:          sensorMac,
			IsConfigured: &isConfigured,
		}
		dump, _ := remove.ToJSON()
		url := "/write/switch/sensor/update/settings"
		s.local.SendCommand(url, dump)
	}
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

func (s *CoreService) packagesInstall(switchConfig sd.SwitchConfig) {
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

func (s *CoreService) packagesRemove(switchConfig sd.SwitchConfig) {
	pkg.RemoveServices(switchConfig.Services)
	for _, service := range switchConfig.Services {
		if _, ok := s.services[service.Name]; ok {
			delete(s.services, service.Name)
		}
	}
}

func (s *CoreService) systemUpdate(switchConfig sd.SwitchConfig) {
	core.SystemUpgrade()
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
					s.friendlyName = event.FriendlyName
					s.updateConfiguration(event)
					s.isConfigured = true

				case network.EventServerSetup:
					s.isConfigured = true
					s.friendlyName = event.FriendlyName
					s.systemUpdate(event)
					s.packagesInstall(event)
					// s.updateConfiguration(event)

				case network.EventServerRemove:
					if !s.isConfigured {
						//a reset is performed
						continue
					}
					s.packagesRemove(event)
					s.removeConfiguration(event)
				}
			}
		}
	}
	return nil
}
