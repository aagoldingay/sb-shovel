package sbcontroller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	servicebus "github.com/Azure/azure-service-bus-go"
)

const (
	ERR_DELETESTATUS     string = "[status] completed %d of %d messages"
	ERR_NOMESSAGESTOSEND string = "no messages to send"
	ERR_NOQUEUEOBJECT    string = "no queue to close"
	ERR_NOTFOUND         string = "could not find service bus queue - 404"
	ERR_QUEUEEMPTY       string = "no messages to pull"
	ERR_UNAUTHORISED     string = "unauthorised or inaccessible service bus. please confirm details - 401"
)

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

	closeQueue(q *servicebus.Queue) error
	getQueueCount(q *servicebus.Queue, dlq bool) (int, error)
	sendMessage(q *servicebus.Queue, data []byte) error
	setupQueue(name string, dlq, purge bool) (*servicebus.Queue, error)
}

type ServiceBusController struct {
	Controller
	client                   *servicebus.Namespace
	ctx                      context.Context
	isSourceDlq, isTargetDlq bool
	source, target           *servicebus.Queue
}

func NewController(conn string) (Controller, error) {
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

func (sb *ServiceBusController) DeleteOneMessage() error {
	if err := sb.source.ReceiveOne(sb.ctx, servicebus.HandlerFunc(func(c context.Context, m *servicebus.Message) error {
		return m.Complete(sb.ctx)
	})); err != nil {
		return err
	}
	return nil
}

func (sb *ServiceBusController) DeleteManyMessages(errChan chan error, total int, delay bool) {
	count := 0
	var wg sync.WaitGroup

	processMessage := func(m *servicebus.Message) {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()
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
		if count == total {
			wg.Add(1)
			go processMessage(m)
			cancel()
			return nil
		}
		wg.Add(1)
		go processMessage(m)
		return nil
	})); err != nil {
		errChan <- err
		return
	}
	wg.Wait()
}

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

func (sb *ServiceBusController) DisconnectSource() error {
	if sb.source != nil {
		return sb.closeQueue(sb.source)
	}
	return errors.New(ERR_NOQUEUEOBJECT)
}

func (sb *ServiceBusController) DisconnectTarget() error {
	if sb.target != nil {
		return sb.closeQueue(sb.target)
	}
	return errors.New(ERR_NOQUEUEOBJECT)
}

func (sb *ServiceBusController) GetSourceQueueCount() (int, error) {
	return sb.getQueueCount(sb.source, sb.isSourceDlq)
}

func (sb *ServiceBusController) GetTargetQueueCount() (int, error) {
	return sb.getQueueCount(sb.target, sb.isTargetDlq)
}

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
			return fmt.Errorf(ERR_DELETESTATUS, count, total)
		}
		if count == total {
			err := processMessage(m)
			if err != nil {
				cancel()
				return err
			}
			return nil
		}
		err := processMessage(m)
		if err != nil {
			cancel()
			return err
		}
		return nil
	})); err != nil {
		return err
	}
	return nil
}

func (sb *ServiceBusController) SendJsonMessage(q bool, data []byte) error {
	if !q {
		return sb.sendMessage(sb.source, data)
	}
	return sb.sendMessage(sb.target, data)
}

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

func (sb *ServiceBusController) SetupSourceQueue(name string, dlq, purge bool) error {
	var err error
	sb.source, err = sb.setupQueue(name, dlq, purge)
	sb.isSourceDlq = dlq
	return err
}

func (sb *ServiceBusController) SetupTargetQueue(name string, dlq, purge bool) error {
	var err error
	sb.target, err = sb.setupQueue(name, dlq, purge)
	sb.isTargetDlq = dlq
	return err
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
