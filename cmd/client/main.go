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
	fmt.Println("Peril game client connected to RabbitMQ!")

	// create channel for publishing
	pubConn, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer pubConn.Close()

	// create username
	fmt.Println("Starting Peril client...")
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not create username: %v", err)
	}

	gameState := gamelogic.NewGameState(username)
	// subscrube client to pause queue
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		1,
		handlerPause(gameState),
	)
	if err != nil {
		fmt.Printf("error subscribing client to pause queue: %v\n", err)
	}

	// subscribe client to army_moves
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		routing.ArmyMovesPrefix+".*",
		1,
		handlerMove(gameState, pubConn),
	)
	if err != nil {
		fmt.Printf("error subscribing client to army_moves: %v\n", err)
	}

	// subscribe client to war
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		"war",
		routing.ArmyMovesPrefix+".*",
		0,
		handlerWar(gameState),
	)
	if err != nil {
		fmt.Printf("error subscribing client to war queue: %v\n", err)
	}

GameLoop:
	for {
		inputs := gamelogic.GetInput()
		if inputs == nil {
			continue
		}
		switch inputs[0] {
		case "spawn":
			err = gameState.CommandSpawn(inputs)
			if err != nil {
				fmt.Println(err)
			}

		case "move":
			// make move locally
			move, err := gameState.CommandMove(inputs)
			if err != nil {
				fmt.Println(err)
			}

			// publish move
			err = pubsub.PublishJSON(
				pubConn,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
				move,
			)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Move published successfully")

		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break GameLoop
		default:
			fmt.Println("Unkown Command")
		}
	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")
}
