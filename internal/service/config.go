package service

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/romana/rlog"
)

func (s *CoreService) toJSON() (string, error) {
	inrec, err := json.Marshal(*s)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

func (s *CoreService) loadPersistentData() error {
	jsonFile, err := os.Open(s.persistentDataPath)
	if err != nil {
		rlog.Error("Cannot OPen configuration" + s.persistentDataPath + " err: " + err.Error())
		return err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, s)
	if err != nil {
		rlog.Error("Cannot Read config to " + s.persistentDataPath + " err: " + err.Error())
	}
	return err
}

func (s *CoreService) writePersistentData() error {
	dump, err := s.toJSON()
	if err != nil {
		rlog.Error("Cannot dump", err.Error())
		return err
	}
	err = ioutil.WriteFile(s.persistentDataPath, []byte(dump), 0644)
	if err != nil {
		rlog.Error("Cannot write config to " + s.persistentDataPath + " err: " + err.Error())
	}
	return err
}
