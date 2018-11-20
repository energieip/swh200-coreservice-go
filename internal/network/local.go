package network

import (
	genericNetwork "github.com/energieip/common-network-go/pkg/network"
	"github.com/energieip/swh200-coreservice-go/pkg/config"
)

//LocalNetwork network object
type LocalNetwork struct {
	Iface genericNetwork.NetworkInterface
}

//CreateLocalNetwork create network server object
func CreateLocalNetwork() (*LocalNetwork, error) {
	driverBroker, err := genericNetwork.NewNetwork(genericNetwork.MQTT)
	if err != nil {
		return nil, err
	}
	driversNet := LocalNetwork{
		Iface: driverBroker,
	}
	return &driversNet, nil

}

//LocalConnection connect service to drivers and services broker
func (net LocalNetwork) LocalConnection(conf config.Configuration, clientID, switchMac string) error {
	cbkLocal := make(map[string]func(genericNetwork.Client, genericNetwork.Message))

	confLocal := genericNetwork.NetworkConfig{}
	confLocal.IP = conf.DriversIP
	confLocal.Port = conf.DriversPort
	confLocal.ClientName = clientID
	confLocal.Callbacks = cbkLocal
	confLocal.LogLevel = *conf.LogLevel

	return net.Iface.Initialize(confLocal)
}

//Disconnect from drivers broker
func (net LocalNetwork) Disconnect() {
	net.Iface.Disconnect()
}

//SendCommand to driver brokers
func (net LocalNetwork) SendCommand(topic, content string) error {
	return net.Iface.SendCommand(topic, content)
}
