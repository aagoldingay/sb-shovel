package sbcontroller

import (
	"context"
	"errors"
	"fmt"
	"strings"

	servicebus "github.com/Azure/azure-service-bus-go"
)

const (
	ERR_NOQUEUEOBJECT string = "no queue to close"
)

type Controller interface {
	DeleteOneMessage(requeue bool) error
	DisconnectQueues() error
	DisconnectSource() error
	DisconnectTarget() error
	GetSourceQueueCount() (int, error)
	GetTargetQueueCount() (int, error)
	SendJsonMessage(q bool, data []byte) error
	SetupSourceQueue(name string, dlq, purge bool) error
	SetupTargetQueue(name string, dlq, purge bool) error

	getQueueCount(q *servicebus.Queue, dlq bool) (int, error)
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

func (sb *ServiceBusController) DeleteOneMessage(requeue bool) error {
	if err := sb.source.ReceiveOne(sb.ctx, servicebus.HandlerFunc(func(c context.Context, m *servicebus.Message) error {
		if requeue {
			err := sb.target.Send(sb.ctx, servicebus.NewMessage(m.Data))
			if err != nil {
				return err
			}
		}
		return m.Complete(sb.ctx)
	})); err != nil {
		return err
	}
	return nil
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

func (sb *ServiceBusController) SendJsonMessage(q bool, data []byte) error {
	if !q {
		return sb.sendMessage(sb.source, data)
	}
	return sb.sendMessage(sb.target, data)
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
