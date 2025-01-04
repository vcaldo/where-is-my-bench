package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
)

type Client struct {
	bot *bot.Bot
}

func NewClient(token string, options ...bot.Option) (*Client, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	b, err := bot.New(token, options...)
	if err != nil {
		return nil, err
	}

	return &Client{
		bot: b,
	}, nil
}

func (c *Client) Start(ctx context.Context) {
	c.bot.Start(ctx)
}

func (c *Client) Close(ctx context.Context) {
	c.bot.Close(ctx)
}
