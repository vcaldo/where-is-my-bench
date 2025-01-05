package handlers

import (
	"bytes"
	"context"
	"log"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func sendMessage(ctx context.Context, b *bot.Bot, chatID int64, text string) error {
	txn := newrelic.FromContext(ctx)
	txn.AddAttribute("chat_id", chatID)
	segment := txn.StartSegment("telegram_api_call.send_message")
	defer segment.End()
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		txn.NoticeError(err)
	}
	return err
}

func sendImage(ctx context.Context, b *bot.Bot, chatID int64, image []byte) error {
	txn := newrelic.FromContext(ctx)
	txn.AddAttribute("chat_id", chatID)
	segment := txn.StartSegment("telegram_api_call.send_photo")
	defer segment.End()

	_, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID: chatID,
		Photo: &models.InputFileUpload{
			Data: bytes.NewReader(image),
		},
	})
	if err != nil {
		txn.NoticeError(err)
	}
	return err
}

func removeImage(ctx context.Context, imgPath string) error {
	txn := newrelic.FromContext(ctx)
	segment := txn.StartSegment("remove_image")
	defer segment.End()

	err := os.Remove(imgPath)
	if err != nil {
		txn.NoticeError(err)
		log.Printf("error removing image: %v", err)
		return err
	}
	return nil
}

func isAdmin(ctx context.Context, adminUserID, userID int64) bool {
	txn := newrelic.FromContext(ctx)
	segment := txn.StartSegment("is_admin")
	defer segment.End()

	if adminUserID == 0 {
		return false
	}
	return adminUserID == userID
}
