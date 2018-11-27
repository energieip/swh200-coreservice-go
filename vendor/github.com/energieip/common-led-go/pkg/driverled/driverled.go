package driverled

import (
	"encoding/json"
)

const (
	DbName    = "status"
	TableName = "leds"

	UrlHello   = "setup/hello"
	UrlStatus  = "status/dump"
	UrlSetup   = "setup/config"
	UrlSetting = "update/settings"
)

//Led led driver representation
type Led struct {
	ID                string  `json:"ID,omitempty"`
	Mac               string  `json:"mac"`
	IP                string  `json:"ip"`
	Group             int     `json:"group"`
	Protocol          string  `json:"protocol"`
	Topic             string  `json:"topic"`
	SwitchMac         string  `json:"switchMac"`
	IsConfigured      bool    `json:"isConfigured"`
	SoftwareVersion   float32 `json:"softwareVersion"`
	HardwareVersion   string  `json:"hardwareVersion"`
	IsBleEnabled      bool    `json:"isBleEnabled"`
	Temperature       int     `json:"temperature"`
	Error             int     `json:"error"`
	ResetNumbers      int     `json:"resetNumbers"`
	InitialSetupDate  float64 `json:"initialSetupDate"`
	LastResetDate     float64 `json:"lastResetDate"`
	IMax              int     `json:"iMax"`
	SlopeStart        int     `json:"slopeStart"`
	SlopeStop         int     `json:"slopeStop"`
	Duration          float64 `json:"duration"`
	Setpoint          int     `json:"setpoint"`
	ThresoldLow       int     `json:"thresoldLow"`
	ThresoldHigh      int     `json:"thresoldHigh"`
	DaisyChainEnabled bool    `json:"daisyChainEnabled"`
	DaisyChainPos     int     `json:"daisyChainPos"`
	DevicePower       int     `json:"devicePower"`
	Energy            float64 `json:"energy"`
	VoltageLed        int     `json:"voltageLed"`
	VoltageInput      int     `json:"voltageInput"`
	LinePower         int     `json:"linePower"`
	TimeToAuto        int     `json:"timeToAuto"`
	Auto              bool    `json:"auto"`
	Watchdog          int     `json:"watchdog"`
	FriendlyName      string  `json:"friendlyName"`
}

//LedSetup initial setup send by the server when the driver is authorized
type LedSetup struct {
	Mac          string  `json:"mac"`
	IMax         int     `json:"iMax"`
	Group        *int    `json:"group"`
	Auto         *bool   `json:"auto"`
	Watchdog     *int    `json:"watchdog"`
	IsBleEnabled *bool   `json:"isBleEnabled"`
	ThresoldHigh *int    `json:"thresoldHigh"`
	ThresoldLow  *int    `json:"thresoldLow"`
	FriendlyName *string `json:"friendlyName"`
}

//LedConf customizable configuration by the server
type LedConf struct {
	Mac          string  `json:"mac"`
	Group        *int    `json:"group"`
	Setpoint     *int    `json:"setpoint"`
	Auto         *bool   `json:"auto"`
	Watchdog     *int    `json:"watchdog"`
	IsConfigured *bool   `json:"isConfigured"`
	IsBleEnabled *bool   `json:"isBleEnabled"`
	ThresoldHigh *int    `json:"thresoldHigh"`
	ThresoldLow  *int    `json:"thresoldLow"`
	FriendlyName *string `json:"friendlyName"`
}

//ToLed convert map interface to Led object
func ToLed(val interface{}) (*Led, error) {
	var light Led
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &light)
	return &light, err
}

//ToLedSetup convert map interface to Led object
func ToLedSetup(val interface{}) (*LedSetup, error) {
	var light LedSetup
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &light)
	return &light, err
}

// ToJSON dump led struct
func (led Led) ToJSON() (string, error) {
	inrec, err := json.Marshal(led)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

// ToJSON dump led setup struct
func (led LedSetup) ToJSON() (string, error) {
	inrec, err := json.Marshal(led)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

//ToJSON dump struct in json
func (led LedConf) ToJSON() (string, error) {
	inrec, err := json.Marshal(led)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}
