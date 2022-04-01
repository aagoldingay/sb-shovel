package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	servicebus "github.com/Azure/azure-service-bus-go"
)

func getMessageCount(ctx context.Context, sb *servicebus.Namespace, queueName string, isDlq bool) (int, error) {
	qm := sb.NewQueueManager()
	qe, err := qm.Get(ctx, queueName)
	if err != nil {
		return 0, err
	}

	if isDlq {
		return int(*qe.CountDetails.DeadLetterMessageCount), nil
	}
	return int(*qe.CountDetails.ActiveMessageCount), nil
}

func emptyQueue(all, requeue bool, ctx context.Context, source, target *servicebus.Queue, total int) {
	if !all {
		if err := source.ReceiveOne(ctx, servicebus.HandlerFunc(func(c context.Context, m *servicebus.Message) error {
			if requeue {
				err := target.Send(ctx, servicebus.NewMessage(m.Data))
				if err != nil {
					return err
				}
			}
			return m.Complete(ctx)
		})); err != nil {
			fmt.Println(err)
		}
	} else {
		count := 0

		msgChan := make(chan *servicebus.Message, 3)
		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			go func() {
				for m := range msgChan {
					wg.Add(1)
					defer wg.Done()
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
					defer cancel()

					if requeue {
						err := target.Send(ctx, servicebus.NewMessage(m.Data))
						if err != nil {
							fmt.Println(err)
						}
					}

					m.Complete(ctx)
				}
			}()
		}

		innerCtx, cancel := context.WithCancel(ctx)
		if err := source.Receive(innerCtx, servicebus.HandlerFunc(func(c context.Context, m *servicebus.Message) error {
			count++
			if count > 0 && count%50 == 0 {
				fmt.Printf("Completed %d of %d messages\n", count, total)
			}
			if count == total {
				defer cancel()
			}
			msgChan <- m
			return nil
		})); err != nil {
			fmt.Println(err)
			return
		}
		wg.Wait()
		close(msgChan)
	}
}

func setupServiceBus(connectionString string) (*servicebus.Namespace, error) {
	sb, err := servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(connectionString))
	if err != nil {
		return nil, err
	}

	return sb, nil
}

func setupQueue(ctx context.Context, sb *servicebus.Namespace, queueName string, isDlq, isEmpty bool) (*servicebus.Queue, error) {
	if isDlq {
		queueName = fmt.Sprintf("%s/%s", queueName, servicebus.DeadLetterQueueName)
	}

	var client *servicebus.Queue
	var err error

	if isEmpty {
		client, err = sb.NewQueue(queueName, servicebus.QueueWithPrefetchCount(250))
	} else {
		client, err = sb.NewQueue(queueName)
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = client.Close(ctx)
	}()

	return client, nil
}

func readQueue(process chan []string, errChannel chan string, ctx context.Context, client *servicebus.Queue, maxWriteCache int, tmpl *template.Template) {
	opts := []servicebus.PeekOption{servicebus.PeekWithPageSize(100)}
	messageIterator, err := client.Peek(ctx, opts...)
	if err != nil {
		fmt.Println(err)
		return
	}
	messagesToPrint := []string{}
	done := false
	fmt.Println("Reading messages...")
	for !messageIterator.Done() && !done {
		if len(messagesToPrint) == maxWriteCache {
			process <- messagesToPrint
			messagesToPrint = []string{}
		}

		msg, err := messageIterator.Next(ctx)
		if err != nil {
			switch err.(type) {
			case servicebus.ErrNoMessages:
				process <- messagesToPrint
				done = true
				continue
			default:
				if strings.Contains(err.Error(), "401") {
					errChannel <- "ERROR 401: Unauthorised or inaccessible Service Bus. Please check your details."
					return
				}
				if strings.Contains(err.Error(), "404") {
					errChannel <- "ERROR 404: Could not find Service Bus queue."
					return
				}
				fmt.Println(err.Error())
				panic(err)
			}
		}
		var b bytes.Buffer
		if err := tmpl.Execute(&b, msg); err != nil {
			errChannel <- err.Error()
			return
		}
		messagesToPrint = append(messagesToPrint, b.String())
	}
	close(process)
}
