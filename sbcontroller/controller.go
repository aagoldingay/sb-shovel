// Package sbcontroller creates a wrapper around azure-service-bus-go to control how sb-shovel should interact with a Service Bus queue.
//
// As azure-service-bus-go does not expose it's own interfaces, the Controller interface was created to support Dependency Injection, allowing dependencies to mock and test logic without interacting with a queue directly.
package sbcontroller

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	servicebus "github.com/Azure/azure-service-bus-go"
)

const (
	ERR_DELETESTATUS     string = "\r[status] completed %d of %d messages"
	ERR_FOUNDPATTERN     string = "[status] identified %s in message"
	ERR_NOMESSAGESTOSEND string = "no messages to send"
	ERR_NOQUEUEOBJECT    string = "no queue to close"
	ERR_NOTFOUND         string = "could not find service bus queue - 404"
	ERR_QUEUEEMPTY       string = "no messages to pull"
	ERR_UNAUTHORISED     string = "unauthorised or inaccessible service bus. please confirm details - 401"
)

// Controller is a generic wrapper to control interactions with a Service Bus client.
type Controller interface {
	DeleteOneMessage() error
	DeleteManyMessages(errChan chan error, total int, delay bool)
	DisconnectQueues() error
	DisconnectSource() error
	DisconnectTarget() error
	GetSourceQueueCount() (int, error)
	GetTargetQueueCount() (int, error)
	ReadSourceQueue(outChan chan []string, errChan chan error, maxWrite int)
	RequeueOneMessage() error
	RequeueManyMessages(total int) error
	SendJsonMessage(q bool, data []byte) error
	SendManyJsonMessages(q bool, data [][]byte) error
	SetupSourceQueue(name string, dlq, purge bool) error
	SetupTargetQueue(name string, dlq, purge bool) error
	TidyMessages(errChan chan error, rex *regexp.Regexp, execute bool, total int)

	closeQueue(q *servicebus.Queue) error
	getQueueCount(q *servicebus.Queue, dlq bool) (int, error)
	sendMessage(q *servicebus.Queue, data []byte) error
	setupQueue(name string, dlq, purge bool) (*servicebus.Queue, error)
}

// ServiceBusController is the concrete implementation for the azure-service-bus-go package.
type ServiceBusController struct {
	Controller
	client                   *servicebus.Namespace
	ctx                      context.Context
	isSourceDlq, isTargetDlq bool
	source, target           *servicebus.Queue
}

// NewServiceBusController builds and returns a ServiceBusController, initialising the azure-service-bus-go package client using a supplied connection string.
func NewServiceBusController(conn string) (Controller, error) {
	ns, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(conn))
	if err != nil {
		return nil, err
	}
	return &ServiceBusController{
		client: ns,
		ctx:    context.Background(),
		source: nil,
		target: nil}, nil
}

// DeleteOneMessage receives then completes exactly ONE message from the queue. An error is returned if a problem was encountered.
func (sb *ServiceBusController) DeleteOneMessage() error {
	if err := sb.source.ReceiveOne(sb.ctx, servicebus.HandlerFunc(func(c context.Context, m *servicebus.Message) error {
		return m.Complete(sb.ctx)
	})); err != nil {
		return err
	}
	return nil
}

// DeleteManyMessages concurrently receives and completes many messages.
//
// Given a total number (e.g. the current value on the queue), this process will run until that many messages have been deleted.
//
// Choosing to action a delay will slow down the operation per 50 messages.
//
// Errors are returned via a channel.
func (sb *ServiceBusController) DeleteManyMessages(errChan chan error, total int, delay bool) {
	count := 0
	var wg sync.WaitGroup

	processMessage := func(m *servicebus.Message) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()
		defer wg.Done()
		m.Complete(ctx)
	}

	innerCtx, cancel := context.WithCancel(sb.ctx)
	if err := sb.source.Receive(innerCtx, servicebus.HandlerFunc(func(c context.Context, m *servicebus.Message) error {
		count++
		if count > 0 && count%50 == 0 {
			errChan <- fmt.Errorf(ERR_DELETESTATUS, count, total)
			if delay {
				time.Sleep(250 * time.Millisecond)
			}
		}
		wg.Add(1)
		go processMessage(m)
		if count == total {
			wg.Wait()
			cancel()
		}
		return nil
	})); err != nil {
		errChan <- err
		return
	}
}

// DisconnectQueues performs both DisconnectSource and DisconnectTarget.
func (sb *ServiceBusController) DisconnectQueues() error {
	err := sb.DisconnectSource()
	if err != nil && err.Error() != ERR_NOQUEUEOBJECT {
		return err
	}
	err = sb.DisconnectTarget()
	if err != nil && err.Error() != ERR_NOQUEUEOBJECT {
		return err
	}
	return nil
}

// DisconnectSource breaks the connection for the queue assigned to the internal source queue attribute on the Controller.
func (sb *ServiceBusController) DisconnectSource() error {
	if sb.source != nil {
		return sb.closeQueue(sb.source)
	}
	return errors.New(ERR_NOQUEUEOBJECT)
}

// DisconnectSource breaks the connection for the queue assigned to the internal target queue attribute on the Controller.
func (sb *ServiceBusController) DisconnectTarget() error {
	if sb.target != nil {
		return sb.closeQueue(sb.target)
	}
	return errors.New(ERR_NOQUEUEOBJECT)
}

// GetSourceQueueCount retrieves the count of messages on the configured source queue.
func (sb *ServiceBusController) GetSourceQueueCount() (int, error) {
	return sb.getQueueCount(sb.source, sb.isSourceDlq)
}

// GetTargetQueueCount retrieves teh count of messages on the configured target queue.
func (sb *ServiceBusController) GetTargetQueueCount() (int, error) {
	return sb.getQueueCount(sb.target, sb.isTargetDlq)
}

// ReadSourceQueue peeks messages on the configured source queue, returning a batch of messages, controlled by the maxWrite variable, to a channel.
//
// Errors are returned on a separate channel.
func (sb *ServiceBusController) ReadSourceQueue(outChan chan []string, errChan chan error, maxWrite int) {
	opts := []servicebus.PeekOption{servicebus.PeekWithPageSize(100)}
	messageIterator, err := sb.source.Peek(sb.ctx, opts...)
	if err != nil {
		errChan <- err
		return
	}

	messagesOutput := []string{}

	done := false
	for !messageIterator.Done() && !done {
		if len(messagesOutput) == maxWrite {
			outChan <- messagesOutput
			messagesOutput = []string{}
		}

		msg, err := messageIterator.Next(sb.ctx)
		if err != nil {
			switch err.(type) {
			case servicebus.ErrNoMessages:
				if len(messagesOutput) > 0 {
					outChan <- messagesOutput
				}
				done = true
				errChan <- errors.New(ERR_QUEUEEMPTY)
				return
			default:
				if strings.Contains(err.Error(), "401") {
					errChan <- errors.New(ERR_UNAUTHORISED)
					return
				}
				if strings.Contains(err.Error(), "404") {
					errChan <- errors.New(ERR_NOTFOUND)
					return
				}
				errChan <- err
				return
			}
		}
		messagesOutput = append(messagesOutput, string(msg.Data))
	}
}

// RequeueOneMessage receives exactly ONE message from the source queue, resends the Data property of a message to the target queue,
// then completes from the source queue.
//
// An error is returned if a problem was encountered.
func (sb *ServiceBusController) RequeueOneMessage() error {
	if err := sb.source.ReceiveOne(sb.ctx, servicebus.HandlerFunc(func(c context.Context, m *servicebus.Message) error {
		err := sb.sendMessage(sb.target, m.Data)
		if err != nil {
			return err
		}
		return m.Complete(sb.ctx)
	})); err != nil {
		return err
	}
	return nil
}

// RequeueManyMessages receives from a source queue, sends as new to the target queue, then completes from the source queue. This is performed
// on many messages, controlled by the total parameter.
func (sb *ServiceBusController) RequeueManyMessages(total int) error {
	count := 0
	processMessage := func(m *servicebus.Message) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()
		err := sb.sendMessage(sb.target, m.Data)
		if err != nil {
			return err
		}
		m.Complete(ctx)
		return nil
	}

	innerCtx, cancel := context.WithCancel(sb.ctx)
	if err := sb.source.Receive(innerCtx, servicebus.HandlerFunc(func(c context.Context, m *servicebus.Message) error {
		count++
		if count > 0 && count%50 == 0 {
			fmt.Printf(ERR_DELETESTATUS, count, total)
		}
		err := processMessage(m)
		if err != nil {
			cancel()
			return err
		}
		if count == total {
			cancel()
		}
		return nil
	})); err != nil {
		return err
	}
	return nil
}

// SendJsonMessage sends to either the source or target queue, passing in solely the message content.
//
// If q is true, the message is sent to target.
// If q is false, the message is sent to the source.
//
// The message is sent in JSON format.
func (sb *ServiceBusController) SendJsonMessage(q bool, data []byte) error {
	if !q {
		return sb.sendMessage(sb.source, data)
	}
	return sb.sendMessage(sb.target, data)
}

// SendManyJsonMessages sends many messages, from an array, to either the source or target.
//
// If q is true, the message is sent to target.
// If q is false, the message is sent to source.
//
// The message is sent in JSON format.
func (sb *ServiceBusController) SendManyJsonMessages(q bool, data [][]byte) error {
	if len(data) == 0 {
		return errors.New(ERR_NOMESSAGESTOSEND)
	}
	for i := 0; i < len(data); i++ {
		err := sb.SendJsonMessage(q, data[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// SetupSourceQueue configures the queue connection, by name and whether the dead letter queue should be treated as the queue.
//
// Specifying purge as true will increase the prefetch count for faster processing of many messages.
func (sb *ServiceBusController) SetupSourceQueue(name string, dlq, purge bool) error {
	var err error
	sb.source, err = sb.setupQueue(name, dlq, purge)
	sb.isSourceDlq = dlq
	return err
}

// SetupTargetQueue configures the queue connection, by name and whether the dead letter queue should be treated as the queue.
//
// Specifying purge as true will increase the prefetch count for faster processing of many messages.
func (sb *ServiceBusController) SetupTargetQueue(name string, dlq, purge bool) error {
	var err error
	sb.target, err = sb.setupQueue(name, dlq, purge)
	sb.isTargetDlq = dlq
	return err
}

// TidyMessages concurrently receives and identifies messages to be deleted based on a supplied regex pattern.
//
// WARNING: This operation will not delete messages by default. Provide execute as true to trigger deletion.
//
// When running without executing, matched messages are output through the error channel, and abandoned, which may result in them getting sent to the dead-letter queue.
func (sb *ServiceBusController) TidyMessages(errChan chan error, rex *regexp.Regexp, execute bool, total int) {
	count := 0
	var wg sync.WaitGroup

	processMessage := func(m *servicebus.Message) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()
		defer wg.Done()

		result := rex.Find(m.Data)

		if string(result) == "" {
			m.Abandon(ctx)
			return
		}

		errChan <- fmt.Errorf(ERR_FOUNDPATTERN, string(result))

		if execute {
			m.Complete(ctx)
		} else {
			m.Abandon(ctx)
		}
	}

	innerCtx, cancel := context.WithCancel(sb.ctx)
	if err := sb.source.Receive(innerCtx, servicebus.HandlerFunc(func(c context.Context, m *servicebus.Message) error {
		count++
		wg.Add(1)
		go processMessage(m)
		if count == total {
			wg.Wait()
			cancel()
		}
		return nil
	})); err != nil {
		errChan <- err
		return
	}
}

func (sb *ServiceBusController) closeQueue(q *servicebus.Queue) error {
	return q.Close(sb.ctx)
}

func (sb *ServiceBusController) getQueueCount(q *servicebus.Queue, dlq bool) (int, error) {
	qm := sb.client.NewQueueManager()

	qe, err := qm.Get(sb.ctx, strings.Split(q.Name, "/")[0])
	if err != nil {
		return 0, err
	}

	if dlq {
		return int(*qe.CountDetails.DeadLetterMessageCount), nil
	}
	return int(*qe.CountDetails.ActiveMessageCount), nil
}

func (sb *ServiceBusController) sendMessage(q *servicebus.Queue, data []byte) error {
	return q.Send(sb.ctx, &servicebus.Message{
		Data:        data,
		ContentType: "application/json",
	})
}

func (sb *ServiceBusController) setupQueue(name string, dlq, purge bool) (*servicebus.Queue, error) {
	if dlq {
		name = fmt.Sprintf("%s/%s", name, servicebus.DeadLetterQueueName)
	}

	var q *servicebus.Queue
	var err error

	if purge {
		q, err = sb.client.NewQueue(name, servicebus.QueueWithPrefetchCount(250))
	} else {
		q, err = sb.client.NewQueue(name)
	}
	if err != nil {
		return nil, err
	}

	return q, nil
}
