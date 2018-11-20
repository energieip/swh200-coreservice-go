package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

//Configuration configuration structure
type Configuration struct {
	ServerIP     string        `json:"server_ip"`
	ServerPort   string        `json:"server_port"`
	DriversIP    string        `json:"drivers_ip"`
	DriversPort  string        `json:"drivers_port"`
	DatabaseIP   string        `json:"database_ip"`
	DatabasePort string        `json:"database_port"`
	LogLevel     *string       `json:"log_level"`
	TimerDump    time.Duration `json:"timer_dump"`
}

//ReadConfig parse the configuration file
func ReadConfig(path string) (*Configuration, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config Configuration

	json.Unmarshal(byteValue, &config)
	if config.LogLevel == nil {
		level := "INFO"
		config.LogLevel = &level
	}
	return &config, nil
}
