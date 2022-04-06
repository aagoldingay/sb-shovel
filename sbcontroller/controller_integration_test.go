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

var config Config

func ReadIntegrationConfig() (Config, error) {
	if config != (Config{}) {
		return config, nil
	}
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

func Test_ServiceBusController_ReadSourceQueue_Empty(t *testing.T) {
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

	returnedMsgs := make(chan []string)
	eChan := make(chan error)

	go sb.ReadSourceQueue(returnedMsgs, eChan, 5)

	done := false
	for !done {
		select {
		case msgs := <-returnedMsgs:
			if len(msgs) > 0 {
				t.Errorf("Queue unexpectedly had %d messages", len(msgs))
			}
		case e := <-eChan:
			if e.Error() != ERR_QUEUEEMPTY {
				t.Error(err)
			}
			done = true
		}
	}

	close(returnedMsgs)
	close(eChan)

	err = sb.DisconnectSource()
	if err != nil {
		t.Error(err)
	}
}

func Test_ServiceBusController_ReadSourceQueue_MultipleMessages_OneBatch(t *testing.T) {
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

	msgBody := `{"message":"hello, world %d"}`
	for i := 0; i < 5; i++ {
		sb.SendJsonMessage(false, []byte(fmt.Sprintf(msgBody, i)))
	}

	returnedMsgs := make(chan []string)
	eChan := make(chan error)

	go sb.ReadSourceQueue(returnedMsgs, eChan, 5)

	batches := 0
	done := false
	for !done {
		select {
		case msgs := <-returnedMsgs:
			if len(msgs) != 5 {
				t.Errorf("Unexpected number of messages returned: %d", len(msgs))
			}

			for i := 0; i < len(msgs); i++ {
				if msgs[i] != fmt.Sprintf(msgBody, i) {
					t.Errorf("Unexpected message body: %s", msgs[i])
				}
			}
			batches++
		case e := <-eChan:
			if e.Error() != ERR_QUEUEEMPTY {
				t.Error(e)
			}
			done = true
		}
	}
	close(returnedMsgs)
	close(eChan)

	if batches != 1 {
		t.Errorf("Did not run 1 times, ran: %d times", batches)
	}

	for i := 0; i < 5; i++ {
		err = sb.DeleteOneMessage(false)
		if err != nil {
			t.Error(err)
		}
	}
	err = sb.DisconnectSource()
	if err != nil {
		t.Error(err)
	}
}

func Test_ServiceBusController_ReadSourceQueue_MultipleMessages_MultipleBatches(t *testing.T) {
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

	msgBody := `{"message":"hello, world %d"}`
	for i := 0; i < 10; i++ {
		sb.SendJsonMessage(false, []byte(fmt.Sprintf(msgBody, i)))
	}

	returnedMsgs := make(chan []string)
	eChan := make(chan error)

	go sb.ReadSourceQueue(returnedMsgs, eChan, 5)

	batches := 0
	i := 0
	done := false
	for !done {
		select {
		case msgs := <-returnedMsgs:
			if len(msgs) != 5 {
				t.Errorf("Unexpected number of messages returned: %d", len(msgs))
			}

			for j := 0; j < len(msgs); j++ {
				if msgs[j] != fmt.Sprintf(msgBody, i) {
					t.Errorf("Unexpected message body: %s", msgs[j])
				}
				i++
			}
			batches++
		case e := <-eChan:
			if e.Error() != ERR_QUEUEEMPTY {
				t.Error(e)
			}
			done = true
		}
	}
	close(returnedMsgs)
	close(eChan)

	if batches != 2 {
		t.Errorf("Did not run 2 times, ran: %d times", batches)
	}

	for i := 0; i < 10; i++ {
		err = sb.DeleteOneMessage(false)
		if err != nil {
			t.Error(err)
		}
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

	c, err := sb.GetSourceQueueCount()
	if c != 0 {
		t.Errorf("Unexpected queue count: %d", c)
	}
	if err != nil {
		t.Error()
	}

	err = sb.SendJsonMessage(false, []byte("hello world"))
	if err != nil {
		t.Error(err)
	}

	c, err = sb.GetSourceQueueCount()
	if c != 1 {
		t.Errorf("Unexpected queue count: %d", c)
	}
	if err != nil {
		t.Error()
	}

	err = sb.DeleteOneMessage(false)
	if err != nil {
		t.Error(err)
	}

	c, err = sb.GetSourceQueueCount()
	if c != 0 {
		t.Errorf("Unexpected queue count: %d", c)
	}
	if err != nil {
		t.Error()
	}
}

func Test_ServiceBusController_Send_And_Delete_Many(t *testing.T) {
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

	c, err := sb.GetSourceQueueCount()
	if c != 0 {
		t.Errorf("Unexpected queue count: %d", c)
	}
	if err != nil {
		t.Error()
	}

	sendMessages := [][]byte{}

	for i := 0; i < 5; i++ {
		sendMessages = append(sendMessages, []byte(`{"hello":"world"}`))
	}

	err = sb.SendManyJsonMessages(false, sendMessages)
	if err != nil {
		t.Error()
	}

	c, err = sb.GetSourceQueueCount()
	if c != 5 {
		t.Errorf("Unexpected queue count: %d", c)
	}
	if err != nil {
		t.Error()
	}

	eChan := make(chan error)
	go sb.DeleteManyMessages(eChan, false, c)

	done := false
	for !done {
		e := <-eChan
		if e.Error() != "context canceled" {
			t.Error(e)
		} else {
			done = true
		}
	}

	close(eChan)

	c, err = sb.GetSourceQueueCount()
	if c != 0 {
		t.Errorf("Unexpected queue count: %d", c)
	}
	if err != nil {
		t.Error()
	}
}

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing for CI pipeline")
	}
}
