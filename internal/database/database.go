package database

import (
	"github.com/energieip/common-database-go/pkg/database"
	gm "github.com/energieip/common-group-go/pkg/groupmodel"
	led "github.com/energieip/common-led-go/pkg/driverled"
	sensor "github.com/energieip/common-sensor-go/pkg/driversensor"
	"github.com/romana/rlog"
)

type Database = database.DatabaseInterface

//ConnectDatabase plug datbase
func ConnectDatabase(ip, port string) (*Database, error) {
	db, err := database.NewDatabase(database.RETHINKDB)
	if err != nil {
		rlog.Error("database err " + err.Error())
		return nil, err
	}

	confDb := database.DatabaseConfig{
		IP:   ip,
		Port: port,
	}
	err = db.Initialize(confDb)
	if err != nil {
		rlog.Error("Cannot connect to database " + err.Error())
		return nil, err
	}
	return &db, nil
}

//GetSwitchLeds return the switch leds
func GetSwitchLeds(db Database, switchMac string) map[string]led.Led {
	leds := make(map[string]led.Led)

	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = switchMac
	ledsStored, err := db.GetRecords(led.DbStatus, led.TableName, criteria)

	if err == nil && ledsStored != nil {
		for _, v := range ledsStored {
			light, err := led.ToLed(v)
			if err != nil {
				continue
			}
			leds[light.Mac] = *light
		}
	}
	return leds
}

//GetSwitchSensors return the switch sensors
func GetSwitchSensors(db Database, switchMac string) map[string]sensor.Sensor {
	sensors := make(map[string]sensor.Sensor)

	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = switchMac
	sensorsStored, err := db.GetRecords(sensor.DbStatus, sensor.TableName, criteria)

	if err == nil && sensorsStored != nil {
		for _, v := range sensorsStored {
			cell, err := sensor.ToSensor(v)
			if err != nil {
				continue
			}
			sensors[cell.Mac] = *cell
		}
	}
	return sensors
}

//GetSensor return the sensor
func GetSensor(db Database, mac string) *sensor.Sensor {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	sensorStored, err := db.GetRecord(sensor.DbStatus, sensor.TableName, criteria)
	if err != nil || sensorStored == nil {
		return nil
	}
	cell, err := sensor.ToSensor(sensorStored)
	if err != nil {
		return nil
	}
	return cell
}

//GetLed return the led
func GetLed(db Database, mac string) *led.Led {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	ledStored, err := db.GetRecord(led.DbStatus, led.TableName, criteria)
	if err != nil || ledStored == nil {
		return nil
	}
	light, err := led.ToLed(ledStored)
	if err != nil {
		return nil
	}
	return light
}

//GetStatusGroup return the switch groups
func GetStatusGroup(db Database, runGroup map[int]bool) map[int]gm.GroupStatus {
	groups := make(map[int]gm.GroupStatus)
	groupsStored, err := db.FetchAllRecords(gm.DbStatusName, gm.TableStatusName)

	if err == nil && groupsStored != nil {
		for _, v := range groupsStored {
			group, err := gm.ToGroupStatus(v)
			if err != nil {
				continue
			}
			if _, ok := runGroup[group.Group]; ok {
				groups[group.Group] = *group
			}
		}
	}
	return groups
}
