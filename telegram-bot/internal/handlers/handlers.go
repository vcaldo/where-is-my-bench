package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func Handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	switch {
	case update.Message != nil && update.Message.Text == "/start":
		startHandler(ctx, b, update)
	case
		update.Message != nil && update.Message.Location != nil:
		locationHandler(ctx, b, update)
	}
}

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	txn := newrelic.FromContext(ctx)
	segment := txn.StartSegment("command.start")
	defer segment.End()

	msgCat := "Hola! Sóc un bot que et pot ajudar a trobar el teu banc a Barcelona.\nEnvia'm la teva ubicació i jo faré la resta."
	magEs := "¡Hola! Soy un bot que puede ayudarte a encontrar tu banco en Barcelona.\nEnvíame tu ubicación y haré el resto."
	msgEn := "Hello! I'm a bot that can help you find your bench in Barcelona.\nJust send me your location and I'll do the rest. "
	msg := fmt.Sprintf("%s\n\n%s\n\n%s\n\n🏃‍♂️‍➡️🪑", msgCat, magEs, msgEn)

	txn.AddAttribute("message_type", "welcome")
	err := sendMessage(ctx, b, update.Message.Chat.ID, msg)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error sending message: %v", err)
	}

}

func locationHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	txn := newrelic.FromContext(ctx)
	segment := txn.StartSegment("command.location")
	defer segment.End()

	lat := update.Message.Location.Latitude
	lon := update.Message.Location.Longitude

	txn.AddAttribute("location_lat", lat)
	txn.AddAttribute("location_lon", lon)

	sendMessage(ctx, b, update.Message.Chat.ID, fmt.Sprintf("lat: %f, lon: %f", lat, lon))
}
