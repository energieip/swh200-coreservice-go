package network

import (
	"encoding/json"
	"time"

	genericNetwork "github.com/energieip/common-network-go/pkg/network"
	"github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/swh200-coreservice-go/pkg/config"
	"github.com/romana/rlog"
)

const (
	EventServerSetup  = "serverSetup"
	EventServerReload = "serverReload"
	EventServerRemove = "serverRemove"
)

//ServerNetwork network object
type ServerNetwork struct {
	Iface  genericNetwork.NetworkInterface
	Events chan map[string]deviceswitch.SwitchConfig
}

//CreateServerNetwork create network server object
func CreateServerNetwork() (*ServerNetwork, error) {
	serverBroker, err := genericNetwork.NewNetwork(genericNetwork.MQTT)
	if err != nil {
		return nil, err
	}
	serverNet := ServerNetwork{
		Iface:  serverBroker,
		Events: make(chan map[string]deviceswitch.SwitchConfig),
	}
	return &serverNet, nil

}

//RemoteServerConnection connect service to server broker
func (net ServerNetwork) RemoteServerConnection(conf config.Configuration, clientID, switchMac string) error {
	cbkServer := make(map[string]func(genericNetwork.Client, genericNetwork.Message))
	cbkServer["/write/switch/"+switchMac+"/setup/config"] = net.onSetup
	cbkServer["/write/switch/"+switchMac+"/update/settings"] = net.onUpdateSetting
	cbkServer["/remove/switch/"+switchMac+"/update/settings"] = net.onRemoveSetting

	confServer := genericNetwork.NetworkConfig{
		IP:         conf.ServerIP,
		Port:       conf.ServerPort,
		ClientName: clientID,
		Callbacks:  cbkServer,
		LogLevel:   *conf.LogLevel,
	}

	for {
		rlog.Info("Try to connect to " + conf.ServerIP)
		err := net.Iface.Initialize(confServer)
		if err == nil {
			rlog.Info(clientID + " connected to server broker " + conf.ServerIP)
			return err
		}
		timer := time.NewTicker(time.Second)
		rlog.Error("Cannot connect to broker " + conf.ServerIP + " error: " + err.Error())
		rlog.Error("Try to reconnect " + conf.ServerIP + " in 1s")

		select {
		case <-timer.C:
			continue
		}
	}
}

func (net ServerNetwork) onSetup(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Info("Switch Setup: Received topic: " + msg.Topic() + " payload: " + string(payload))
	var switchConf deviceswitch.SwitchConfig
	err := json.Unmarshal(payload, &switchConf)
	if err != nil {
		rlog.Error("Cannot parse config ", err.Error())
		return
	}

	event := make(map[string]deviceswitch.SwitchConfig)
	event[EventServerSetup] = switchConf
	net.Events <- event
}

func (net ServerNetwork) onRemoveSetting(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Info("Force switch system update onRemoveSetting: Received topic: " + msg.Topic() + " payload: " + string(payload))
	var switchConf deviceswitch.SwitchConfig
	err := json.Unmarshal(payload, &switchConf)
	if err != nil {
		rlog.Error("Cannot parse config ", err.Error())
		return
	}

	event := make(map[string]deviceswitch.SwitchConfig)
	event[EventServerRemove] = switchConf
	net.Events <- event
}

func (net ServerNetwork) onUpdateSetting(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Info("Force switch system update onSwitchUpdate: Received topic: " + msg.Topic() + " payload: " + string(payload))
	var switchConf deviceswitch.SwitchConfig
	err := json.Unmarshal(payload, &switchConf)
	if err != nil {
		rlog.Error("Cannot parse config ", err.Error())
		return
	}

	event := make(map[string]deviceswitch.SwitchConfig)
	event[EventServerReload] = switchConf
	net.Events <- event
}

//Disconnect from server
func (net ServerNetwork) Disconnect() {
	net.Iface.Disconnect()
}

//SendCommand to server
func (net ServerNetwork) SendCommand(topic, content string) error {
	return net.Iface.SendCommand(topic, content)
}
