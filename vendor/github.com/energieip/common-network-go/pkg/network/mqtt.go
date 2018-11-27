package network

import (
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/romana/rlog"
)

//Message message Payload
type Message = mqtt.Message

//Client representation
type Client = mqtt.Client

//MQTTNetwork Protocol MQTT
type MQTTNetwork struct {
	mqttClient  *mqtt.Client
	brokerEvent chan map[string]string
	callbacks   map[string]func(Client, Message)
	config      *NetworkConfig
}

//SendCommand send a command
func (p *MQTTNetwork) SendCommand(topic, payload string) error {
	if p.mqttClient == nil {
		return NewError("mqtt client not instanciate")
	}
	if token := (*p.mqttClient).Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
		rlog.Error("Cannot send a command" + token.Error().Error())
		return token.Error()
	}
	return nil
}

func onConnLostHandler(c Client, err error) {
	rlog.Error("Connection lost, reason: " + err.Error())
}

//Disconnect disconnect from MQTT
func (p *MQTTNetwork) Disconnect() {
	if p.mqttClient == nil {
		return
	}
	if (*p.mqttClient).IsConnected() {
		for topic := range p.callbacks {
			rlog.Info("Unsubscribe to topic " + topic)
			if token := (*p.mqttClient).Unsubscribe(topic); token.Wait() && token.Error() != nil {
				rlog.Error("Cannot Unsubscribe to topic" + token.Error().Error())
			}
		}
		(*p.mqttClient).Disconnect(500)
	}
}

//Initialize protocol communication
func (p *MQTTNetwork) Initialize(config NetworkConfig) error {
	rlog.Info("Plug on MQTT " + config.IP + ":" + config.Port)
	p.callbacks = config.Callbacks
	url := "tcp://" + config.IP + ":" + config.Port
	opts := mqtt.NewClientOptions().AddBroker(url).SetClientID(config.ClientName)
	switch config.LogLevel {
	case LogDebug:
		mqtt.DEBUG = log.New(os.Stdout, "DBG", 0)
		fallthrough
	case LogInfo:
		fallthrough
	case LogWarning:
		mqtt.WARN = log.New(os.Stdout, "WRN", 0)
		fallthrough
	case LogError:
		mqtt.ERROR = log.New(os.Stdout, "ERR", 0)
		fallthrough
	case LogCritical:
		mqtt.CRITICAL = log.New(os.Stdout, "CRT", 0)
	default:
		break
	}
	opts.SetKeepAlive(2 * time.Second)
	// opts.SetDefaultPublishHandler(f)
	opts.SetConnectionLostHandler(onConnLostHandler)

	opts.OnConnect = func(c Client) {
		rlog.Info("Client connected")

		//Subscribe here because when a connection is lost
		//we have to re-subscribe manually to topics
		for topic, callback := range p.callbacks {
			rlog.Info("Subscribing to topic " + topic)
			if token := c.Subscribe(topic, 0, callback); token.Wait() && token.Error() != nil {
				rlog.Error("ERROR when subscribe to " + topic + " err: " + token.Error().Error())
			}
		}
	}
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
		return token.Error()
	}

	p.mqttClient = &c
	p.config = &config

	return nil
}
