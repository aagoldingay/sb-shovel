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

type ConfigController struct {
	ConfigManager
	file string
	args map[string]string
	// loaded  bool
	updated bool
}

func NewConfigController() (ConfigManager, error) {
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	f := filepath.Dir(ex) + "/.sb-shovel"
	return &ConfigController{
		file:    f,
		args:    make(map[string]string),
		updated: false}, nil
}

func (cc *ConfigController) UpdateConfig(k string, v string) {
	cc.args[k] = v
	cc.updated = true
}

func (cc *ConfigController) DeleteConfigValue(k string) error {
	if len(cc.args) == 0 {
		return errors.New(ERR_CONFIGEMPTY)
	}
	delete(cc.args, k)
	cc.updated = true
	return nil
}

func (cc *ConfigController) GetConfigValue(k string) (string, error) {
	if len(cc.args) == 0 {
		return "", errors.New(ERR_CONFIGEMPTY)
	}
	return cc.args[k], nil
}

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

func (cc *ConfigController) NewConfigFile() error {
	if cc.file == "" {
		return errors.New(ERR_PATHEMPTY)
	}
	if len(cc.args) == 0 {
		return errors.New(ERR_NOCHANGES)
	}
	return nil
}

// CALL AT END OF MAIN FUNC
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
