package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType string

const (
	Ack         AckType = "ack"
	NackRequeue AckType = "nack_requeue"
	NackDiscard AckType = "nack_discard"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	dat, err := json.Marshal(val)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return err
	}

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        dat,
	}

	ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		msg,
	)

	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	// make sure the queue exists
	ch, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		return fmt.Errorf("declareAndBind: %w", err)
	}

	// get channel of Delivery structs
	newch, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	// handles messages
	go func() {
		defer ch.Close()
		for data := range newch {
			var msg T
			err := json.Unmarshal(data.Body, &msg)
			if err != nil {
				fmt.Println(err)
				continue
			}
			ack := handler(msg)
			switch ack {
			case Ack:
				log.Println("Ack message")
				data.Ack(false)
			case NackRequeue:
				log.Println("NackRequeue message")
				data.Nack(false, true)
			case NackDiscard:
				log.Println("NackDiscard message")
				data.Nack(false, false)
			}
		}
	}()

	return nil
}
