package handler

import (
	"encoding/json"
	"fiatless/internal/ijson"
	"fmt"
	"log"
)

type BaseHandler struct{}

func (h *BaseHandler) SendErrorResult(client *ijson.IJSONClient, requestID any, message string) error {
	return client.SendResult(map[string]any{
		"id":    requestID,
		"error": message,
	})
}

// ParseParams parses command parameters into the specified struct
func (h *BaseHandler) ParseParams(command map[string]any, params any) (any, error) {
	requestID, ok := command["id"]
	if !ok {
		return nil, fmt.Errorf("missing request ID")
	}

	paramsMap, ok := command["params"].(map[string]any)
	if !ok {
		log.Printf("Invalid params format: %v", command["params"])
		return requestID, fmt.Errorf("invalid params format")
	}

	paramsJSON, err := json.Marshal(paramsMap)
	if err != nil {
		log.Printf("Failed to marshal params: %v", err)
		return requestID, fmt.Errorf("invalid params format")
	}

	if err := json.Unmarshal(paramsJSON, params); err != nil {
		log.Printf("Failed to unmarshal params: %v", err)
		return requestID, fmt.Errorf("invalid params format")
	}

	return requestID, nil
}
