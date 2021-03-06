package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Script represents a single, user script which performs the actual scan
// of an address (IP/network or FQDN).
type Script struct {
	Name      string
	LocalPath string
	RepoURL   string
	Manifest  *Manifest
}

var execCommand = exec.Command

// NewScript creates a new instance of Script given as fileName and performs some basic
// validation of a file associated with it (e.g., is it executable).
// Script should be located in "scripts" subdir of cfgDir. When cfgDir is given as an
// empty string, then "~/.ralph-cli/scripts" will be searched.
func NewScript(fileName, cfgDir string) (Script, error) {
	path := filepath.Join(cfgDir, "scripts", fileName)
	finfo, err := os.Stat(path)
	if err != nil {
		return Script{}, err
	}
	exec := finfo.Mode() & 0100
	if exec == 0 {
		return Script{}, fmt.Errorf("file %s is not executable for the owner", path)
	}
	return Script{
		Name:      fileName,
		LocalPath: path,
		RepoURL:   "",
		Manifest:  nil,
	}, nil
}

// Run launches a scan Script on a given address (at this moment, only IPs are fully
// supported).
func (s Script) Run(addr Addr) (*ScanResult, error) {
	var res ScanResult
	var err error
	cmd := execCommand(s.LocalPath, string(addr))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &res, fmt.Errorf("error running script %s: %s\nstderr: %s",
			s.LocalPath, err, cmd.Stderr)
	}
	err = json.Unmarshal(output, &res)
	return &res, err
}

// ScanResult holds parsed output of a scan script.
type ScanResult struct {
	// TODO(xor-xor): Consider adding here a field holding an ADDR being scanned.
	MACAddresses []MACAddress `json:"mac_addresses"`
	Disks        []Disk
	Memory       []Memory
	// TODO(xor-xor): Consider using Model type instead of string here.
	Model      string `json:"model_name"`
	Processors []Processor
	SN         string `json:"serial_number"`
}

func (sr ScanResult) String() string {
	return fmt.Sprintf("MACAddresses: %s\nDisks: %s\nMemory: %s\nModel: %s\nProcessors: %s\nSerial Number: %s\n",
		sr.MACAddresses, sr.Disks, sr.Memory, sr.Model, sr.Processors, sr.SN)
}
