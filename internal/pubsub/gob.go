package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishLog(
	gs *gamelogic.GameState,
	ch *amqp.Channel,
	message string,
	attacker string,
) AckType {
	// make log struct
	log := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     message,
		Username:    gs.GetUsername(),
	}

	// publish Gob of log
	err := PublishGob(
		ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+attacker,
		log,
	)
	if err != nil {
		fmt.Println(err)
		return NackRequeue
	}
	return Ack
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(val); err != nil {
		return fmt.Errorf("Error encoding Gob: %s", err)
	}

	msg := amqp.Publishing{
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	}

	return ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		msg,
	)

}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	// create unmarshaller
	unmarshaller := func(data []byte) (T, error) {
		buf := bytes.NewBuffer(data)
		var target T
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&target)
		return target, err
	}

	// subscribe with gob decoder
	if err := subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		unmarshaller,
	); err != nil {
		return fmt.Errorf("Error subscribing Gob log: %s", err)
	}
	return nil
}
