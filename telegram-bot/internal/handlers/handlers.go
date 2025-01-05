package handlers

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/internal/config"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/internal/downloader"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/internal/storage/redis"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/pkg/bench"
	"github.com/vcaldo/where-is-my-bench/telegram-bot/pkg/maps"
)

const searchRadius float64 = 250

func Handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("error loading config: %v", err)
		return
	}

	switch {
	case update.Message != nil && update.Message.Text == "/start":
		startHandler(ctx, b, update)
	case update.Message != nil && update.Message.Location != nil:
		locationHandler(ctx, cfg, b, update)
	case update.Message != nil && update.Message.Text == "/update_benches":
		if !isAdmin(ctx, cfg.AdminUserID, update.Message.From.ID) {
			log.Printf("unauthorized admin command received: %s\n %d not equal %d", update.Message.Text, cfg.AdminUserID, update.Message.From.ID)
			err := sendMessage(ctx, b, update.Message.Chat.ID, "You are not authorized to perform this action.")
			if err != nil {
				log.Printf("error sending message: %v", err)
			}
			return
		}
		log.Printf("authorized admin command received: %s", update.Message.Text)
		updateBenchesHandler(ctx, cfg, b, update)
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

func locationHandler(ctx context.Context, cfg *config.Config, b *bot.Bot, update *models.Update) {
	txn := newrelic.FromContext(ctx)
	segment := txn.StartSegment("command.location")
	defer segment.End()

	rdb := redis.NewBenchStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

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

func updateBenchesHandler(ctx context.Context, cfg *config.Config, b *bot.Bot, update *models.Update) {
	txn := newrelic.FromContext(ctx)
	segment := txn.StartSegment("command.location")
	defer segment.End()

	d := downloader.NewDownloader(cfg.BenchesDatasetURL)
	data, err := d.DownloadJSON(ctx)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error downloading JSON: %v", err)
		return
	}

	benches, err := bench.LoadBenches(ctx, data)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error loading benches: %v", err)
		return
	}

	if len(benches) == 0 {
		log.Printf("no benches found in the dataset %s, skipping update", cfg.BenchesDatasetURL)

		msg := "No benches found in the dataset, skipping update."
		err = sendMessage(ctx, b, update.Message.Chat.ID, msg)
		if err != nil {
			txn.NoticeError(err)
			log.Printf("error sending message: %v", err)
		}
		return
	}

	log.Printf("Found %d benches, updating redis", len(benches))

	rdb := redis.NewBenchStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	err = rdb.DeleteAllBenches(ctx)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error deleting all benches: %v", err)
		return
	}

	err = rdb.StoreBenches(ctx, benches)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error storing benches: %v", err)
		return
	}

	msg := fmt.Sprintf("Successfully updated %d benches 🪑", len(benches))
	err = sendMessage(ctx, b, update.Message.Chat.ID, msg)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error sending message: %v", err)
		return
	}
}
