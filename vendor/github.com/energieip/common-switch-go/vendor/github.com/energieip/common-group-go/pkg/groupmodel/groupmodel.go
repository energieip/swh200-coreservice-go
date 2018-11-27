package groupmodel

import (
	"encoding/json"

	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/common-sensor-go/pkg/driversensor"
)

const (
	// Sensor rules when there are several sensors in the same group
	SensorAverage = "average"
	SensorMin     = "min"
	SensorMax     = "max"

	DbStatusName    = "status"
	TableStatusName = "groups"

	GroupConfigUrl = "update/settings"
)

//GroupBase
type GroupBase struct {
	Group              int     `json:"group"` //groupID
	SensorRule         *string `json:"sensorRule"`
	Auto               *bool   `json:"auto"`
	SlopeStart         *int    `json:"slopeStart"`
	SlopeStop          *int    `json:"slopeStop"`
	Watchdog           *int    `json:"watchdog"`
	CorrectionInterval *int    `json:"correctionInterval"`
	GroupRules         *Rule   `json:"groupRules"`
}

//GroupConfig representation
type GroupConfig struct {
	GroupBase
	Leds    []string `json:"leds"` //Mac address list
	Sensors []string `json:"sensors"`
}

// Rule when the group is in automatic mode
type Rule struct {
	Brightness *int `json:"brightness"`
	Presence   *int `json:"presence"`
}

type Setpoint struct {
	SpLeds *int `json:"spLeds"`
}

//GroupRuntime runtime execution
type GroupRuntime struct {
	GroupBase
	Setpoints *Setpoint             `json:"setpoints"`
	Leds      []driverled.Led       `json:"leds"`
	Sensors   []driversensor.Sensor `json:"sensors"`
}

//GroupStatus status dump to the server
type GroupStatus struct {
	ID                 string   `json:"ID,omitempty"` //database id
	Group              int      `json:"group"`        //groupID
	SensorRule         string   `json:"sensorRule"`
	Auto               bool     `json:"auto"`
	SlopeStart         int      `json:"slopeStart"`
	SlopeStop          int      `json:"slopeStop"`
	Watchdog           int      `json:"watchdog"`
	CorrectionInterval int      `json:"correctionInterval"`
	GroupRules         Rule     `json:"groupRules"`
	Error              int      `json:"error"`
	TimeToAuto         int      `json:"timeToAuto"`
	SetpointLeds       int      `json:"setpoint_leds"`
	Presence           bool     `json:"presence"`
	TimeToLeave        int      `json:"timeToLeave"`
	Leds               []string `json:"leds"` //Mac address list
	Sensors            []string `json:"sensors"`
}

// ToMapInterface convert group struct in Map[string] interface{}
func (group GroupConfig) ToMapInterface() map[string]interface{} {
	var inInterface map[string]interface{}
	inrec, _ := json.Marshal(group)
	json.Unmarshal(inrec, &inInterface)
	return inInterface
}

// ToJSON dump group struct
func (group GroupConfig) ToJSON() (string, error) {
	inrec, err := json.Marshal(group)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

//ToGroupConfig convert interface to group config object
func ToGroupConfig(val interface{}) (*GroupStatus, error) {
	var group GroupStatus
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &group)
	return &group, err
}

// ToMapInterface convert group struct in Map[string] interface{}
func (group GroupRuntime) ToMapInterface() map[string]interface{} {
	var inInterface map[string]interface{}
	inrec, _ := json.Marshal(group)
	json.Unmarshal(inrec, &inInterface)
	return inInterface
}

// ToJSON dump group struct
func (group GroupRuntime) ToJSON() (string, error) {
	inrec, err := json.Marshal(group)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

//ToGroupRuntime convert interface to group runtime object
func ToGroupRuntime(val interface{}) (*GroupRuntime, error) {
	var group GroupRuntime
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &group)
	return &group, err
}

// ToMapInterface convert GroupStatus struct in Map[string] interface{}
func (group GroupStatus) ToMapInterface() map[string]interface{} {
	var inInterface map[string]interface{}
	inrec, _ := json.Marshal(group)
	json.Unmarshal(inrec, &inInterface)
	return inInterface
}

// ToJSON dump group GroupStatus
func (group GroupStatus) ToJSON() (string, error) {
	inrec, err := json.Marshal(group)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

//ToGroupStatus convert interface to status object
func ToGroupStatus(val interface{}) (*GroupStatus, error) {
	var group GroupStatus
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &group)
	return &group, err
}
