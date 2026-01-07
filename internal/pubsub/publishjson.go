package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
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
