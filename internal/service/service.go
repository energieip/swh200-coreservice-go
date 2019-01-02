package service

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/common-sensor-go/pkg/driversensor"
	pkg "github.com/energieip/common-service-go/pkg/service"
	"github.com/energieip/common-switch-go/pkg/deviceswitch"
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
	Mac                   string `json:"mac"` //Switch mac address
	events                chan string
	timerDump             time.Duration          //in seconds
	IP                    string                 `json:"ip"`
	IsConfigured          bool                   `json:"isConfigured"`
	Groups                map[int]bool           `json:"groups"`
	Services              map[string]pkg.Service `json:"services"`
	LastSystemUpgradeDate string                 `json:"lastSystemUpgradeDate"`
	persistentDataPath    string
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
	s.persistentDataPath = conf.PersistentDataPath

	os.Setenv("RLOG_LOG_LEVEL", conf.LogLevel)
	os.Setenv("RLOG_LOG_NOTIME", "yes")
	rlog.UpdateEnv()
	rlog.Info("Starting SwitchCore service")

	err = s.loadPersistentData()
	if err != nil {
		s.Groups = make(map[int]bool)
		s.Services = make(map[string]pkg.Service)
		s.LastSystemUpgradeDate = core.GetLastSystemUpgradeDate()

		s.Mac, s.IP = tools.GetNetworkInfo()
		s.Mac = strings.ToUpper(strings.Replace(s.Mac, ":", "", -1))
		s.IsConfigured = false
		s.writePersistentData()
	}

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

	err = s.local.LocalConnection(*conf, clientID, s.Mac)
	if err != nil {
		rlog.Error("Cannot connect to drivers broker " + conf.LocalBroker.IP + " error: " + err.Error())
		return err
	}

	go s.server.RemoteServerConnection(*conf, clientID, s.Mac)
	rlog.Info("SwitchCore service started")
	return nil
}

//Stop service
func (s *CoreService) Stop() {
	rlog.Info("Stopping SwitchCore service")
	s.writePersistentData()
	s.server.Disconnect()
	s.local.Disconnect()
	s.db.Close()
	rlog.Info("SwitchCore service stopped")
}

func (s *CoreService) sendHello() {
	switchDump := deviceswitch.Switch{
		Mac:   s.Mac,
		IP:    s.IP,
		Topic: "switch/" + s.Mac,
		LastSystemUpgradeDate: s.LastSystemUpgradeDate,
		IsConfigured:          &s.IsConfigured,
		Protocol:              "MQTT",
	}
	dump, err := switchDump.ToJSON()
	if err != nil {
		rlog.Errorf("Could not dump switch %v status %v", s.Mac, err.Error())
		return
	}

	err = s.server.SendCommand("/read/"+switchDump.Topic+"/"+UrlHello, dump)
	if err != nil {
		rlog.Errorf("Could not send hello to the server %v status %v", s.Mac, err.Error())
		return
	}
	rlog.Infof("Hello %v sent to the server", s.Mac)
}

func (s *CoreService) sendDump() {
	status := deviceswitch.SwitchStatus{}
	status.Mac = s.Mac
	status.Protocol = "MQTT"
	status.IP = s.IP
	status.IsConfigured = &s.IsConfigured
	status.LastSystemUpgradeDate = s.LastSystemUpgradeDate
	status.Topic = "switch/" + s.Mac
	services := make(map[string]pkg.ServiceStatus)

	for _, c := range s.Services {
		component := pkg.ServiceStatus{}
		component.Name = c.Name
		component.PackageName = c.PackageName
		component.Version = c.Version
		status := component.GetServiceStatus()
		component.Status = &status
		services[component.Name] = component
	}

	status.Services = services
	status.Leds = database.GetSwitchLeds(s.db, s.Mac)
	status.Sensors = database.GetSwitchSensors(s.db, s.Mac)
	status.Groups = database.GetStatusGroup(s.db, s.Groups)

	dump, err := status.ToJSON()
	if err != nil {
		rlog.Error("Could not dump switch status ", err.Error())
		return
	}

	err = s.server.SendCommand("/read/"+status.Topic+"/"+UrlStatus, dump)
	if err != nil {
		rlog.Errorf("Could not dump switch %v status %v", s.Mac, err.Error())
		return
	}
	rlog.Infof("Status %v sent to the server", s.Mac)
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

	for grID := range switchConfig.Groups {
		_, ok := s.Groups[grID]
		if !ok {
			s.Groups[grID] = true
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
	s.writePersistentData()
}

func (s *CoreService) removeConfiguration(switchConfig deviceswitch.SwitchConfig) {
	for grID, group := range switchConfig.Groups {
		_, ok := s.Groups[grID]
		if ok {
			delete(s.Groups, grID)
		}
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
	s.writePersistentData()
}

func (s *CoreService) updateServicesConfiguration() {
	for _, srv := range s.Services {
		if srv.ConfigPath == "" {
			continue
		}
		rlog.Info("Update " + srv.Name + " configuration, path :" + srv.ConfigPath)
		err := pkg.WriteServiceConfig(srv.ConfigPath, srv.Config)
		if err != nil {
			rlog.Warn("Cannot write configuration for " + srv.ConfigPath + " err: " + err.Error())
			continue
		}
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

func (s *CoreService) packagesInstall(switchConfig deviceswitch.SwitchConfig) {
	newService := false
	for name, service := range switchConfig.Services {
		if currentState, ok := s.Services[name]; ok {
			if currentState.Version == service.Version {
				rlog.Info("Package " + name + " already in version " + service.Version + " skip it")
				continue
			}
		}
		newService = true
		rlog.Info("Install " + name + " in version " + service.Version)
		service.Install()
		version := pkg.GetPackageVersion(service.PackageName)
		if version != nil {
			service.Version = *version
		}
		s.Services[service.Name] = service
	}
	if newService {
		s.writePersistentData()
	}
}

func (s *CoreService) packagesRemove(switchConfig deviceswitch.SwitchConfig) {
	newService := false
	pkg.RemoveServices(switchConfig.Services)
	for _, service := range switchConfig.Services {
		if _, ok := s.Services[service.Name]; ok {
			delete(s.Services, service.Name)
			newService = true
		}
	}
	if newService {
		s.writePersistentData()
	}
}

func (s *CoreService) systemUpdate(switchConfig deviceswitch.SwitchConfig) {
	rlog.Info("Get " + s.LastSystemUpgradeDate + " and expect " + switchConfig.LastSystemUpgradeDate)
	if switchConfig.LastSystemUpgradeDate == s.LastSystemUpgradeDate {
		//system is already up to date
		rlog.Info("System is already up to date")
		return
	}
	core.SystemUpgrade()
	s.writePersistentData()
}

func (s *CoreService) startServices() {
	pkg.StartServices(s.Services)
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
				if s.IsConfigured {
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
						s.IsConfigured = *event.IsConfigured
					}
					if !s.IsConfigured {
						//a reset is performed
						s.writePersistentData()
						continue
					}
					//In this case reload == setup
					s.updateConfiguration(event)
					s.IsConfigured = true

				case network.EventServerSetup:
					s.IsConfigured = true
					s.systemUpdate(event)
					s.packagesInstall(event)
					s.updateServicesConfiguration()
					s.startServices()
					// s.updateConfiguration(event)

				case network.EventServerRemove:
					if !s.IsConfigured {
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
