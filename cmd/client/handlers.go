package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	handlerFunc := func(rout routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(rout)
		return pubsub.Ack
	}
	return handlerFunc
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	handlerFunc := func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.Ack
		case gamelogic.MoveOutcomeSafe:
			return pubsub.Ack

		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Println(err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		default:
			fmt.Println("error: unknown move outcome")
			return pubsub.NackDiscard
		}
	}
	return handlerFunc
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(war gamelogic.RecognitionOfWar) pubsub.AckType {
	handlerFunc := func(war gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(war)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.PublishLog(
				gs,
				ch,
				fmt.Sprintf("%s won a war against %s", winner, loser),
				war.Attacker.Username,
			)

		case gamelogic.WarOutcomeYouWon:
			return pubsub.PublishLog(
				gs,
				ch,
				fmt.Sprintf("%s won a war against %s", winner, loser),
				war.Attacker.Username,
			)

		case gamelogic.WarOutcomeDraw:
			return pubsub.PublishLog(
				gs,
				ch,
				fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser),
				war.Attacker.Username,
			)

		default:
			fmt.Println("Error handling war")
			return pubsub.NackDiscard
		}
	}
	return handlerFunc
}
