package core

import (
	"bytes"
	"os/exec"
	"strings"
)

const (
	ServiceRunning = "running"
	ServiceFailed  = "failed"
	ServiceMissing = "missing"
	ServiceStop    = "stopped"
)

//Service description
type Service struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	PackageName string `json:"packageName"` //DebianPackageName
}

//ServiceStatus description
type ServiceStatus struct {
	Service
	Status *string `json:"status"` //enable/running/disable etc.
}

// GetServiceStatus return service status
func (s Service) GetServiceStatus() string {
	outputActive := &bytes.Buffer{}
	cmd := exec.Command("systemctl", "is-active", s.Name)
	cmd.Stdout = outputActive
	cmd.Run()
	output := strings.TrimSpace(string(outputActive.Bytes()))
	switch output {
	case "failed":
		return ServiceFailed
	case "active":
		return ServiceRunning
	default:
		outputEnable := &bytes.Buffer{}
		cmd = exec.Command("systemctl", "is-enabled", s.Name)
		cmd.Stdout = outputEnable
		cmd.Run()
		output = strings.TrimSpace(string(outputEnable.Bytes()))
		if output == "disabled" {
			return ServiceStop
		}
		return ServiceMissing
	}
}

// Install install a given service
func (s Service) Install() (string, error) {
	cmd := exec.Command("apt-get", "install", "-y", s.PackageName)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Remove a given service
func (s Service) Remove() (string, error) {
	cmd := exec.Command("apt-get", "remove", "-y", s.PackageName)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Start a given service
func (s Service) Start() (string, error) {
	cmd := exec.Command("systemctl", "start", s.Name)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Stop a given service
func (s Service) Stop() (string, error) {
	cmd := exec.Command("systemctl", "stop", s.Name)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

//InstallPackages start all given services
func InstallPackages(services map[string]Service) {
	for _, service := range services {
		service.Install()
	}
}

//StartServices start all given services
func StartServices(services map[string]Service) {
	for _, service := range services {
		service.Start()
	}
}

//RemoveServices remove all given services
func RemoveServices(services map[string]Service) {
	for _, service := range services {
		service.Stop()
		service.Remove()
	}
}

//GetPackageVersion return package version
func GetPackageVersion(service string) *string {
	cmd := exec.Command("apt", "show", service)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err := cmd.Run()
	if err != nil {
		return nil
	}
	output := string(cmdOutput.Bytes())
	for _, line := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {
		if !strings.HasPrefix(line, "Version:") {
			continue
		}
		lineSplit := strings.Split(line, " ")
		if len(lineSplit) > 1 {
			version := lineSplit[1]
			return &version
		}
	}

	return nil
}
