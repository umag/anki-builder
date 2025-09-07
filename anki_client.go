package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AnkiConnectClient handles AnkiConnect API interactions
type AnkiConnectClient struct {
	BaseURL string
}

// NewAnkiConnectClient creates a new AnkiConnect client
func NewAnkiConnectClient(baseURL string) *AnkiConnectClient {
	return &AnkiConnectClient{
		BaseURL: baseURL,
	}
}

// IsAvailable checks if AnkiConnect is available
func (a *AnkiConnectClient) IsAvailable(ctx context.Context) bool {
	request := AnkiConnectRequest{
		Action:  "version",
		Version: 6,
		Params:  map[string]interface{}{},
	}

	_, err := a.SendRequest(ctx, request)
	return err == nil
}

// AddNote adds a note to Anki
func (a *AnkiConnectClient) AddNote(ctx context.Context, deckName, modelName string, fields map[string]string, tags []string) error {
	note := map[string]interface{}{
		"deckName":  deckName,
		"modelName": modelName,
		"fields":    fields,
		"tags":      tags,
	}

	request := AnkiConnectRequest{
		Action:  "addNote",
		Version: 6,
		Params: map[string]interface{}{
			"note": note,
		},
	}

	_, err := a.SendRequest(ctx, request)
	return err
}

// GetDeckNames returns available deck names
func (a *AnkiConnectClient) GetDeckNames(ctx context.Context) ([]string, error) {
	request := AnkiConnectRequest{
		Action:  "deckNames",
		Version: 6,
		Params:  map[string]interface{}{},
	}

	response, err := a.SendRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	var decks []string
	if result, ok := response.Result.([]interface{}); ok {
		for _, deck := range result {
			if deckName, ok := deck.(string); ok {
				decks = append(decks, deckName)
			}
		}
	}

	return decks, nil
}

// GetModelNames returns available model names
func (a *AnkiConnectClient) GetModelNames(ctx context.Context) ([]string, error) {
	request := AnkiConnectRequest{
		Action:  "modelNames",
		Version: 6,
		Params:  map[string]interface{}{},
	}

	response, err := a.SendRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	var models []string
	if result, ok := response.Result.([]interface{}); ok {
		for _, model := range result {
			if modelName, ok := model.(string); ok {
				models = append(models, modelName)
			}
		}
	}

	return models, nil
}

// GetModelFieldNames returns field names for a specific model
func (a *AnkiConnectClient) GetModelFieldNames(ctx context.Context, modelName string) ([]string, error) {
	request := AnkiConnectRequest{
		Action:  "modelFieldNames",
		Version: 6,
		Params: map[string]interface{}{
			"modelName": modelName,
		},
	}

	response, err := a.SendRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	var fields []string
	if result, ok := response.Result.([]interface{}); ok {
		for _, field := range result {
			if fieldName, ok := field.(string); ok {
				fields = append(fields, fieldName)
			}
		}
	}

	return fields, nil
}

// SendRequest sends a request to AnkiConnect and returns the response
func (a *AnkiConnectClient) SendRequest(ctx context.Context, request AnkiConnectRequest) (*AnkiConnectResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AnkiConnect request failed with status: %d", resp.StatusCode)
	}

	var response AnkiConnectResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("AnkiConnect error: %v", response.Error)
	}

	return &response, nil
}
