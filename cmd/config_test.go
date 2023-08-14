package cmd

import (
	"testing"

	cc "github.com/aagoldingay/sb-shovel/config"
	// sbio "github.com/aagoldingay/sb-shovel/io"
)

func Test_Config_Update_Existing(t *testing.T) {
	cfg, err := cc.NewConfigController("sb-shovel")
	if err != nil {
		t.Error(err)
	}

	cfg.UpdateConfig("TEST_CONFIG_UPDATE_EXISTING", "old_value")
	cfg.SaveConfig()

	config(cfg, []string{"update", "TEST_CONFIG_UPDATE_EXISTING", "new_value"})

	v, err := cfg.GetConfigValue("TEST_CONFIG_UPDATE_EXISTING")

	if err != nil {
		t.Error(err)
	}

	if v != "new_value" {
		t.Errorf("value for TEST_CONFIG_UPDATE_EXISTING was not as expected")
	}
}

func Test_Config_Remove(t *testing.T) {
	cfg, err := cc.NewConfigController("sb-shovel")
	if err != nil {
		t.Error(err)
	}

	cfg.UpdateConfig("TEST_CONFIG_REMOVE", "TEST_VALUE")
	cfg.SaveConfig()

	config(cfg, []string{"remove", "TEST_CONFIG_REMOVE"})

	v, err := cfg.GetConfigValue("TEST_CONFIG_REMOVE")

	if err.Error() != cc.ERR_CONFIGEMPTY {
		t.Error(err)
	}

	if v != "" {
		t.Errorf("config returned a value: %s", v)
	}
}

func Test_Config_List(t *testing.T) {
	cfg, err := cc.NewConfigController("sb-shovel")
	if err != nil {
		t.Error(err)
	}

	if v := cfg.ListConfig(); v != cc.ERR_CONFIGEMPTY {
		t.Errorf("config wasn't empty: %s", v)
	}
}
