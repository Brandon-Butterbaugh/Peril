package pubsub

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	TypeDurable SimpleQueueType = iota
	TypeTransient
)

var typeName = map[SimpleQueueType]string{
	TypeDurable:   "durable",
	TypeTransient: "transient",
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	newCh, err := conn.Channel()
	if err != nil {
		log.Printf("Error creating new channel: %s", err)
		return nil, amqp.Queue{}, err
	}
	defer newCh.Close()

	que, err := newCh.QueueDeclare(
		queueName,
		queueType == TypeDurable,
		queueType == TypeTransient,
		queueType == TypeTransient,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Error creating queue: %s", err)
		return nil, amqp.Queue{}, err
	}

	err = newCh.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Error binding queue: %s", err)
		return nil, amqp.Queue{}, err
	}

	return newCh, que, nil
}
