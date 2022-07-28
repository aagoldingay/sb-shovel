// Package config is used to persist key/value pairs to a hidden file in the same location as an executable.
//
// Usage:
// - Initiate a ConfigController using NewConfigController("myApp") to produce a config file .myApp
// - Load an existing config using LoadConfig()
// - List present key/value pairs (staged or loaded from config) using ListConfig()
// - Retrieve config settings using GetConfigValue(...)
// - Stage key/value pairs to the ConfigController using UpdateConfig(...), DeleteConfigValue(...)
// - Save staged changes to the config file using SaveConfig()
package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	ERR_PATHEMPTY   string = "no config path provided"
	ERR_NOCHANGES   string = "no changes to save"
	ERR_NOCONFIG    string = "config file not configured"
	ERR_CONFIGEMPTY string = "config file is empty"
	CHAR_NEWLINE    string = "\n"
	CHAR_DELIMITER  string = "|"
)

// ConfigManager is a generic wrapper for controlling config interactions.
type ConfigManager interface {
	UpdateConfig(k string, v string)
	DeleteConfigValue(k string) error
	GetConfigValue(k string) (string, error)
	LoadConfig() error
	ListConfig() string
	NewConfigFile() error
	SaveConfig() error

	formatMap() string
}

// ConfigController is the concrete implementation of ConfigManager
//
// This contains a map of key/value pairs from an existing config file, or any staged changes made during a session, that have not yet been saved.
//
// Any changes which have not been saved by the time the process exits, will be lost.
//
// Call SaveConfig() before the end of the process to update the persisted config file.
type ConfigController struct {
	ConfigManager
	file string
	args map[string]string
	// loaded  bool
	updated bool
}

// NewConfigController creates a new ConfigController.
//
// name: This parameter configures the filename and, for best practice, should be the name of the executable only.
//
// Example for `myApp.exe`: NewConfigController("myApp") creates .myApp config file.
func NewConfigController(name string) (ConfigManager, error) {
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	f := filepath.Dir(ex) + fmt.Sprintf("/.%s", name)
	return &ConfigController{
		file:    f,
		args:    make(map[string]string),
		updated: false}, nil
}

// UpdateConfig adds a key/value pair to the ConfigController struct.
//
// WARNING: This does not automatically update the config file. Use SaveConfig() to persist to storage.
func (cc *ConfigController) UpdateConfig(k string, v string) {
	cc.args[k] = v
	cc.updated = true
}

// DeleteConfigValue deletes a config setting by looking for the key.
//
// WARNING: This does not automatically update the config file. Use SaveConfig() to persist to storage.
func (cc *ConfigController) DeleteConfigValue(k string) error {
	if len(cc.args) == 0 {
		return errors.New(ERR_CONFIGEMPTY)
	}
	delete(cc.args, k)
	cc.updated = true
	return nil
}

// GetConfigValue retrieves a value for a provided key, from ConfigController.
//
// NOTE: This will be empty in a new session unless LoadConfig() is called, first.
// NOTE: Staged config updates are accessible.
func (cc *ConfigController) GetConfigValue(k string) (string, error) {
	if len(cc.args) == 0 {
		return "", errors.New(ERR_CONFIGEMPTY)
	}
	return cc.args[k], nil
}

// LoadConfig reads the file of the same name passed in to NewConfigController.
func (cc *ConfigController) LoadConfig() error {
	cfg, err := ioutil.ReadFile(cc.file)
	if err != nil {
		if strings.Contains(err.Error(), "The system cannot find the file specified.") {
			return errors.New(ERR_NOCONFIG)
		}
		return err
	}
	cc.args = cc.splitInput(strings.Split(string(cfg), CHAR_NEWLINE), CHAR_DELIMITER)
	return nil
}

// ListConfig outputs existing key/value pairs.
func (cc *ConfigController) ListConfig() string {
	if len(cc.args) == 0 {
		return ERR_CONFIGEMPTY
	}
	s := ""
	for k, v := range cc.args {
		s += fmt.Sprintf("%s=%s\n", k, v)
	}
	return s
}

// NewConfigFile creates a new file with the name passed in to NewConfigController.
func (cc *ConfigController) NewConfigFile() error {
	if cc.file == "" {
		return errors.New(ERR_PATHEMPTY)
	}
	if len(cc.args) == 0 {
		return errors.New(ERR_NOCHANGES)
	}
	return nil
}

// SaveConfig overwrites the config file with the key/value pairs staged and/or loaded from config.
//
// This only runs if a change has been staged by UpdateConfig() or DeleteConfigValue()
func (cc *ConfigController) SaveConfig() error {
	if !cc.updated {
		return errors.New(ERR_NOCHANGES)
	}
	s := cc.formatMap()

	err := ioutil.WriteFile(cc.file, []byte(s), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (cc *ConfigController) formatMap() string {
	s := ""
	for k, v := range cc.args {
		s += fmt.Sprintf("%s|%s\n", k, v)
	}
	return s
}

func (cc *ConfigController) splitInput(arr []string, del string) map[string]string {
	m := make(map[string]string)
	for i, s := range arr {
		if i == len(arr)-2 { // -2 due to index starting at 0 + final line in file being 0
			kv := strings.Split(s, del)
			m[kv[0]] = kv[1]
			break
		}
		kv := strings.Split(s, del)
		m[kv[0]] = kv[1]
	}
	return m
}
