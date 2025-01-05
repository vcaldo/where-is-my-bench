package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/internal/config"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/internal/storage/redis"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/pkg/bench"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/pkg/maps"
)

const searchRadius float64 = 250

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

	msg := "Hello! I'm a bot that can help you find your bench in Barcelona.\nJust send me your location and I'll do the rest. "
	msgFmt := fmt.Sprintf("%s\n\n🏃‍♂️‍➡️🪑", msg)

	txn.AddAttribute("message_type", "welcome")
	err := sendMessage(ctx, b, update.Message.Chat.ID, msgFmt)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error sending message: %v", err)
	}
}

func locationHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	txn := newrelic.FromContext(ctx)
	segment := txn.StartSegment("command.location")
	defer segment.End()

	cfg, err := config.LoadConfig()
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error loading config: %v", err)
		return
	}

	redisDB, _ := strconv.Atoi(cfg.RedisDB)
	rdb := redis.NewBenchStore(cfg.RedisAddr, cfg.RedisPassword, redisDB)

	benches, err := rdb.FindNearby(ctx, update.Message.Location.Latitude, update.Message.Location.Longitude, searchRadius)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error finding benches: %v", err)
		return
	}

	benchesNearby := make([]bench.Bench, len(benches))
	for i, b := range benches {
		bench, err := rdb.GetBenchByID(ctx, b.GisID)
		if err != nil {
			log.Fatalf("error getting bench by id: %v", err)
		}
		benchesNearby[i] = *bench
	}

	mg := maps.NewMapGenerator
	imgPath, err := mg().GenerateMap(ctx, update.Message.Location.Latitude, update.Message.Location.Longitude, searchRadius, benchesNearby)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error generating map: %v", err)
		return
	}

	img, err := os.ReadFile(imgPath)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error reading image file: %v", err)
		return
	}

	msg := fmt.Sprintf("I found %d benches 🪑 in a %.0f m radius near you:", len(benchesNearby), searchRadius)
	err = sendMessage(ctx, b, update.Message.Chat.ID, msg)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error sending message: %v", err)
		return
	}

	err = sendImage(ctx, b, update.Message.Chat.ID, img)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error sending image: %v", err)
		return
	}

	err = removeImage(ctx, imgPath)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error removing image: %v", err)
	}
}
