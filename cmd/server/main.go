package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// connect to rabbitMQ
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	// create peril_topic exchange
	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		"game_logs",
		"game_logs.*",
		0,
	)
	if err != nil {
		log.Fatalf("could not create exchange: %v", err)
	}

	// subscribe server to log queue
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.TypeDurable,
		handlerLog(),
	)
	if err != nil {
		log.Fatalf("error subscribing client to war queue: %v\n", err)
	}

	// server loop
	gamelogic.PrintServerHelp()
ServerLoop:
	for {
		inputs := gamelogic.GetInput()
		if inputs == nil {
			continue
		}
		switch inputs[0] {
		case "pause":
			fmt.Println("Sending pause message...")

			// pause the server
			pubConn, err := conn.Channel()
			if err != nil {
				log.Fatalf("could not create channel: %v", err)
			}
			defer pubConn.Close()
			err = pubsub.PublishJSON(
				pubConn,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)
			if err != nil {
				log.Fatalf("could not publish pause queue json: %v", err)
			}

		case "resume":
			fmt.Println("Sending resume message...")

			// resume the server
			pubConn, err := conn.Channel()
			if err != nil {
				log.Fatalf("could not create channel: %v", err)
			}
			defer pubConn.Close()
			err = pubsub.PublishJSON(
				pubConn,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			)
			if err != nil {
				log.Fatalf("could not publish json: %v", err)
			}

		case "quit":
			fmt.Println("Exiting server...")
			break ServerLoop

		default:
			fmt.Println("Command not understood")
		}
	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")
}
