package core

import (
	"github.com/energieip/common-group-go/pkg/groupmodel"
	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/common-sensor-go/pkg/driversensor"
	"github.com/energieip/swh200-coreservice-go/internal/database"
)

//CreateGroup return new runtime group, if the group is local
func CreateGroup(db database.Database, mac string, gr groupmodel.GroupConfig) *groupmodel.GroupRuntime {
	group := groupmodel.GroupRuntime{}
	group.Auto = gr.Auto
	group.CorrectionInterval = gr.CorrectionInterval
	if group.CorrectionInterval == nil {
		interval := 1
		group.CorrectionInterval = &interval
	}
	group.Group = gr.Group
	group.GroupRules = gr.GroupRules
	group.SensorRule = gr.SensorRule
	if group.SensorRule == nil {
		rule := "average"
		group.SensorRule = &rule
	}
	group.SlopeStart = gr.SlopeStart
	if group.SlopeStart == nil {
		start := 10
		group.SlopeStart = &start
	}

	group.SlopeStop = gr.SlopeStop
	if group.SlopeStop == nil {
		stop := 10
		group.SlopeStop = &stop
	}
	group.Watchdog = gr.Watchdog
	if group.Watchdog == nil {
		watchdog := 86400 //1day
		group.Watchdog = &watchdog
	}

	var sensors []driversensor.Sensor
	for _, mac := range gr.Sensors {
		sensor := database.GetSensor(db, mac)
		if sensor == nil {
			continue
		}
		sensors = append(sensors, *sensor)
	}
	group.Sensors = sensors

	var leds []driverled.Led
	for _, mac := range gr.Leds {
		led := database.GetLed(db, mac)
		if led == nil {
			continue
		}
		leds = append(leds, *led)
	}
	group.Leds = leds
	group.SetpointLeds = gr.SetpointLeds
	return &group
}

//UpdateGroup Update existing group. If the group is no longer local: remove it
func UpdateGroup(db database.Database, mac string, gr groupmodel.GroupRuntime, gc groupmodel.GroupConfig) *groupmodel.GroupRuntime {
	if gc.Auto != nil {
		gr.Auto = gc.Auto
	}
	if gc.CorrectionInterval != nil {
		gr.CorrectionInterval = gc.CorrectionInterval
	}
	if gc.GroupRules != nil {
		if gr.GroupRules == nil {
			rule := groupmodel.Rule{}
			gr.GroupRules = &rule
		}
		if gc.GroupRules.Brightness != nil {
			gr.GroupRules.Brightness = gc.GroupRules.Brightness
		}
		if gc.GroupRules.Presence != nil {
			gr.GroupRules.Presence = gc.GroupRules.Presence
		}
	}
	if gc.SlopeStart != nil {
		gr.SlopeStart = gc.SlopeStart
	}
	if gc.SlopeStop != nil {
		gr.SlopeStop = gc.SlopeStop
	}
	if gc.Watchdog != nil {
		gr.Watchdog = gc.Watchdog
	}

	var sensors []driversensor.Sensor
	for _, mac := range gc.Sensors {
		sensor := database.GetSensor(db, mac)
		if sensor == nil {
			continue
		}
		sensors = append(sensors, *sensor)
	}
	gr.Sensors = sensors

	var leds []driverled.Led
	for _, mac := range gc.Leds {
		led := database.GetLed(db, mac)
		if led == nil {
			continue
		}
		leds = append(leds, *led)
	}
	gr.Leds = leds

	if gc.SetpointLeds != nil {
		gr.SetpointLeds = gc.SetpointLeds
	}
	return &gr
}
