package handler

import (
	"context"
	"fiatless/internal/ijson"
)

type CommandHandler interface {
	CommandPath() string

	Handle(ctx context.Context, client *ijson.IJSONClient, command map[string]any) error
}
