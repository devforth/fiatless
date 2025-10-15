package processor

import (
	"context"
	"fiatless/internal/blockchain/handler"
	"fiatless/internal/ijson"
	"log"
	"sync"
)

// CommandProcessor processes blockchain commands
type CommandProcessor struct {
	client   *ijson.IJSONClient
	handlers map[string]handler.CommandHandler
	wg       sync.WaitGroup
}

// NewCommandProcessor creates a new command processor
func NewCommandProcessor(client *ijson.IJSONClient) *CommandProcessor {
	p := &CommandProcessor{
		client:   client,
		handlers: make(map[string]handler.CommandHandler),
	}

	return p
}

// RegisterHandler registers a command handler
func (p *CommandProcessor) RegisterHandler(handler handler.CommandHandler) {
	p.handlers[handler.CommandPath()] = handler
}

// RegisterHandlers registers multiple handlers at once
func (p *CommandProcessor) RegisterHandlers(handlers []handler.CommandHandler) {
	for _, h := range handlers {
		p.RegisterHandler(h)
	}
}

// Start begins processing commands
func (p *CommandProcessor) Start() {
	log.Println("Starting command processor...")

	for path, handler := range p.handlers {
		p.wg.Add(1)
		go p.processCommandPath(path, handler)
	}

	p.wg.Wait() // Wait for all handlers to finish (they won't in practice)
}

// processCommandPath processes commands for a specific path
func (p *CommandProcessor) processCommandPath(path string, handler handler.CommandHandler) {
	defer p.wg.Done()

	log.Printf("Starting command listener for path: %s", path)

	for {
		command, err := p.client.GetCommandLongPolling(path)
		if err != nil {
			log.Printf("Error getting command for path %s: %v", path, err)
			continue
		}

		if command != nil {
			log.Printf("Received command for path %s: %v", path, command)
			ctx := context.Background()
			go func(cmd map[string]any) {
				if err := handler.Handle(ctx, p.client, cmd); err != nil {
					log.Printf("Error processing command %s: %v", path, err)
				}
			}(command)
		}
	}
}
