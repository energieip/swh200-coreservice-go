package driversensor

import (
	"encoding/json"
)

const (
	DbName    = "status"
	TableName = "sensors"

	UrlHello   = "setup/hello"
	UrlStatus  = "status/dump"
	UrlSetup   = "setup/config"
	UrlSetting = "update/settings"
)

//Sensor driver representation
type Sensor struct {
	ID                         string  `json:"ID,omitempty"`
	IP                         string  `json:"ip"`
	Mac                        string  `json:"mac"`
	Group                      int     `json:"group"`
	Protocol                   string  `json:"protocol"`
	Topic                      string  `json:"topic"`
	SwitchMac                  string  `json:"switchMac"`
	IsConfigured               bool    `json:"isConfigured"`
	SoftwareVersion            float32 `json:"softwareVersion"`
	HardwareVersion            string  `json:"hardwareVersion"`
	IsBleEnabled               bool    `json:"isBleEnabled"`
	Temperature                int     `json:"temperature"`
	Error                      int     `json:"error"`
	ResetNumbers               int     `json:"resetNumbers"`
	InitialSetupDate           float64 `json:"initialSetupDate"`
	LastResetDate              float64 `json:"lastResetDate"`
	Brigthness                 int     `json:"brightness"`
	Presence                   bool    `json:"presence"`
	BrigthnessCorrectionFactor int     `json:"brigthnessCorrectionFactor"`
	ThresoldPresence           int     `json:"thresoldPresence"`
	TemperatureOffset          int     `json:"temperatureOffset"`
	BrigthnessRaw              int     `json:"brigthnessRaw"`
	LastMovment                int     `json:"lastMovement"`
	VoltageInput               int     `json:"voltageInput"`
	TemperatureRaw             int     `json:"temperatureRaw"`
	FriendlyName               string  `json:"friendlyName"`
}

//SensorSetup initial setup send by the server when the driver is authorized
type SensorSetup struct {
	Mac                        string  `json:"mac"`
	Group                      *int    `json:"group"`
	BrigthnessCorrectionFactor *int    `json:"brigthnessCorrectionFactor"`
	ThresoldPresence           *int    `json:"thresoldPresence"`
	TemperatureOffset          *int    `json:"temperatureOffset"`
	IsBleEnabled               *bool   `json:"isBleEnabled"`
	FriendlyName               *string `json:"friendlyName"`
}

//SensorConf customizable configuration by the server
type SensorConf struct {
	Mac                        string  `json:"mac"`
	Group                      *int    `json:"group"`
	BrigthnessCorrectionFactor *int    `json:"brigthnessCorrectionFactor"`
	IsConfigured               *bool   `json:"isConfigured"`
	ThresoldPresence           *int    `json:"thresoldPresence"`
	TemperatureOffset          *int    `json:"temperatureOffset"`
	IsBleEnabled               *bool   `json:"isBleEnabled"`
	FriendlyName               *string `json:"friendlyName"`
}

//ToSensor convert interface to Sensor object
func ToSensor(val interface{}) (*Sensor, error) {
	var cell Sensor
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &cell)
	return &cell, err
}

//ToSensorSetup convert interface to SensorSetup object
func ToSensorSetup(val interface{}) (*SensorSetup, error) {
	var cell SensorSetup
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &cell)
	return &cell, err
}

// ToJSON dump sensor struct
func (sensor Sensor) ToJSON() (string, error) {
	inrec, err := json.Marshal(sensor)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

// ToJSON dump sensor struct
func (sensor SensorSetup) ToJSON() (string, error) {
	inrec, err := json.Marshal(sensor)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

//ToJSON dump struct in json
func (sensor SensorConf) ToJSON() (string, error) {
	inrec, err := json.Marshal(sensor)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}
