package core

import (
	"encoding/json"
	"io/ioutil"
	"os"

	group "github.com/energieip/common-group-go/pkg/groupmodel"
	led "github.com/energieip/common-led-go/pkg/driverled"
	sensor "github.com/energieip/common-sensor-go/pkg/driversensor"
)

// Switch description
type Switch struct {
	Mac                   string `json:"mac"`
	Protocol              string `json:"protocol"`
	IP                    string `json:"ip"`
	Topic                 string `json:"topic"`
	LastSystemUpgradeDate string `json:"lastUpgradeDate"`
	IsConfigured          *bool  `json:"isConfigured"`
}

//SwitchConfig content
type SwitchConfig struct {
	Switch
	Services      map[string]Service            `json:"services"`
	Groups        map[int]group.GroupConfig     `json:"groups"`
	LedsSetup     map[string]led.LedSetup       `json:"ledsSetup"`
	LedsConfig    map[string]led.LedConf        `json:"ledsConfig"`
	SensorsSetup  map[string]sensor.SensorSetup `json:"sensorsSetup"`
	SensorsConfig map[string]sensor.SensorConf  `json:"sensorsConfig"`
}

//SwitchStatus description
type SwitchStatus struct {
	Switch
	Status    string                    `json:"status"` //enable/running/disable etc.
	ErrorCode *int                      `json:"errorCode"`
	Services  map[string]ServiceStatus  `json:"services"`
	Leds      map[string]led.Led        `json:"leds"`
	Sensors   map[string]sensor.Sensor  `json:"sensors"`
	Groups    map[int]group.GroupStatus `json:"groups"`
}

// ToJSON dump status struct
func (status SwitchStatus) ToJSON() (string, error) {
	inrec, err := json.Marshal(status)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

// ToJSON dump switch struct
func (status Switch) ToJSON() (string, error) {
	inrec, err := json.Marshal(status)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

// LoadSwitchConfig load in memory the json config
func LoadSwitchConfig(path string) (*SwitchConfig, error) {
	var switchConf SwitchConfig
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(byteValue, &switchConf)
	return &switchConf, err
}
