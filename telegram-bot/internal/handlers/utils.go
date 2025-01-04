package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func sendMessage(ctx context.Context, b *bot.Bot, chatID int64, text string) error {
	txn := newrelic.FromContext(ctx)
	txn.AddAttribute("chat_id", chatID)
	segment := txn.StartSegment("telegram_api_call")
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
