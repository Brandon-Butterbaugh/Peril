package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	handlerFunc := func(rout routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(rout)
		return pubsub.Ack
	}
	return handlerFunc
}
