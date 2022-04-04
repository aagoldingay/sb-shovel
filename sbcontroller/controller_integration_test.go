package sbcontroller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

type Config struct {
	ConnectionString string
	Queue            string
}

func ReadIntegrationConfig() (Config, error) {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		return Config{}, err
	}

	for _, f := range files {
		fmt.Println(f.Name())
	}
	c, err := ioutil.ReadFile("../test_files/integration.json")
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(c, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

func Test_ServiceBusController_NewController_FakeResource(t *testing.T) {
	sb, err := NewController("Endpoint=sb://fake.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=NoTaReAlAcCeSsKeY=")

	if err != nil {
		t.Error(err)
	}

	err = sb.SetupSourceQueue("queue", false, false)
	if err != nil {
		t.Error(err)
	}

	c, err := sb.GetSourceQueueCount()
	if c != 0 {
		t.Errorf("Source queue returned an unexpected count: %d", c)
	}
	if !strings.Contains(err.Error(), "no such host") {
		t.Error(err)
	}
}

func Test_ServiceBusController_NewController_Success(t *testing.T) {
	skipCI(t)

	config, err := ReadIntegrationConfig()
	if err != nil {
		t.Errorf("Test setup failed %v", err)
	}

	sb, err := NewController(config.ConnectionString)

	if err != nil {
		t.Error(err)
	}

	err = sb.SetupSourceQueue(config.Queue, false, false)
	if err != nil {
		t.Error(err)
	}

	_, err = sb.GetSourceQueueCount()
	if err != nil {
		t.Error(err)
	}

	err = sb.DisconnectSource()
	if err != nil {
		t.Error(err)
	}
}

func Test_ServiceBusController_DisconnectQueues(t *testing.T) {
	skipCI(t)

	config, err := ReadIntegrationConfig()
	if err != nil {
		t.Errorf("Test setup failed %v", err)
	}

	sb, err := NewController(config.ConnectionString)

	if err != nil {
		t.Error(err)
	}

	err = sb.SetupSourceQueue(config.Queue, true, false)
	if err != nil {
		t.Error(err)
	}

	err = sb.SetupTargetQueue(config.Queue, false, false)
	if err != nil {
		t.Error(err)
	}

	_, err = sb.GetSourceQueueCount()
	if err != nil {
		t.Error(err)
	}
	_, err = sb.GetTargetQueueCount()
	if err != nil {
		t.Error(err)
	}

	err = sb.DisconnectQueues()
	if err != nil {
		t.Error(err)
	}
}

func Test_ServiceBusController_GetSourceQueueCount_DeadLetterQueue(t *testing.T) {
	skipCI(t)

	config, err := ReadIntegrationConfig()
	if err != nil {
		t.Errorf("Test setup failed %v", err)
	}

	sb, err := NewController(config.ConnectionString)

	if err != nil {
		t.Error(err)
	}

	err = sb.SetupSourceQueue(config.Queue, true, false)
	if err != nil {
		t.Error(err)
	}

	c, err := sb.GetSourceQueueCount()
	if c != 0 {
		t.Errorf("Source queue returned an unexpected count: %d", c)
	}
	if err != nil {
		t.Error(err)
	}

	err = sb.DisconnectSource()
	if err != nil {
		t.Error(err)
	}
}

func Test_ServiceBusController_Send_And_Delete_One(t *testing.T) {
	skipCI(t)

	config, err := ReadIntegrationConfig()
	if err != nil {
		t.Errorf("Test setup failed %v", err)
	}

	sb, err := NewController(config.ConnectionString)
	if err != nil {
		t.Error(err)
	}

	err = sb.SetupSourceQueue(config.Queue, false, false)
	if err != nil {
		t.Error(err)
	}

	err = sb.SendJsonMessage(false, []byte("hello world"))
	if err != nil {
		t.Error(err)
	}

	c, err := sb.GetSourceQueueCount()
	if c != 1 {
		t.Errorf("Unexpected queue count: %d", c)
	}
	if err != nil {
		t.Error()
	}

	// cleanup
	err = sb.DeleteOneMessage(false)
	if err != nil {
		t.Error(err)
	}
}

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing for CI pipeline")
	}
}
