package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"anki-builder/aislop/gemini"
	"anki-builder/ankiclient"
	"anki-builder/config"
)

const shutdownTimeout = 5 * time.Second

type FinnishCard struct {
	Finnish        string
	Translation    string
	FinnishExample string
	Notes          string
}

type AIResponse struct {
	Phrase       string   `json:"phrase"`
	Translations []string `json:"translations"`
	Examples     []string `json:"examples"`
	Notes        []string `json:"notes"`
}

//nolint:forbidigo // fmt prints are cute
func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived %v signal. Shutting down...\n", sig)
		cancel()

		time.Sleep(shutdownTimeout)
		fmt.Printf("Slept for %s, force exit triggered", shutdownTimeout)
		os.Exit(1)
	}()

	fmt.Println("Finnish Anki Card Builder")
	fmt.Printf("Using AI Provider: gemini\n")
	fmt.Println("Enter Finnish words or phrases (to exit use Ctrl+C or type 'q', 'quit' or 'exit'):")

	scanner := bufio.NewScanner(os.Stdin)
	dataChan := make(chan string)

loop:
	for {
		go func() {
			fmt.Print("> ")
			if !scanner.Scan() {
				dataChan <- "quit"
			}

			input := strings.TrimSpace(scanner.Text())
			dataChan <- input
		}()

		var input string
		// Check if context was cancelled
		select {
		case <-ctx.Done():
			break loop
		case input = <-dataChan:
		}

		if input == "quit" || input == "q" || input == "exit" {
			break loop
		}

		if input == "" {
			continue
		}

		fmt.Printf("Processing: %s\n", input)

		card, err := generateCard(ctx, cfg, input)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				fmt.Println("Operation cancelled. Shutting down...")
				return
			}
			fmt.Printf("Error generating card: %v\n", err)
			continue
		}

		err = addCardToAnki(ctx, cfg, card)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				fmt.Println("Operation cancelled. Shutting down...")
				break loop
			}
			fmt.Printf("Error adding card to Anki: %v\n", err)
			continue
		}

		fmt.Printf("✅ Successfully added card for '%s'\n\n", input)
	}

	fmt.Println("Goodbye!")
}

func generateCard(ctx context.Context, cfg *config.Config, input string) (*FinnishCard, error) {
	aiResponse, err := generateWithGemini(ctx, cfg, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query AI provider: %w", err)
	}

	var translations string
	for i, tl := range aiResponse.Translations {
		tl = "- " + strings.ToLower(tl)
		if i > 0 {
			translations += "<br>"
		}
		translations += tl
	}

	card := &FinnishCard{
		Finnish:        aiResponse.Phrase,
		Translation:    translations,
		FinnishExample: strings.Join(aiResponse.Examples, "<br>"),
		Notes:          strings.Join(aiResponse.Notes, "<br>"),
	}

	return card, nil
}

func generateWithGemini(ctx context.Context, cfg *config.Config, input string) (*AIResponse, error) {
	client := gemini.NewClient(cfg.GeminiAPIKey)

	prompt := buildPrompt(input)

	responseText, err := client.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content with Gemini: %w", err)
	}

	return parseAIResponse(responseText)
}

func buildPrompt(input string) string {
	return fmt.Sprintf(`You are a Finnish language expert helping to create Anki flashcards for language learners.

For the Finnish word/phrase: "%s"

Please provide a JSON response with the following structure:
{
  "phrase": "the original Finnish word/phrase in dictionary form, lowercase",
  "translations": ["translation1", "translation2", ...],
  "examples": [
    "Example sentence 1 in Finnish",
    "Example sentence 2 in Finnish",
    "Example sentence 3 in Finnish",
    "Example sentence 4 in Finnish"
  ],
  "notes": [
    "synonyms: abc, def...", 
    "etymology: short info on word origin", 
    "extra: grammatical information, usage quirks, and any other useful information for language learners"
  ]
}

Guidelines:
- Provide all relevant translations (most common meanings)
- Create 3-4 example sentences at B1-B2 level, try to include examples for different translations of the word/phrase
- Make examples natural and contextually rich
- Use the word in different grammatical cases/forms when possible
- Include etymology, synonyms, grammatical notes, and usage tips in the notes section, but don't add obvious information - keep it concise.
- Use lowercase for notes
- Make sure to add puhekieli versions of the word to synonyms if applicable
- Try not to skip etymology if available
- Keep the prefixes of the notes consistent (e.g. always use "synonyms:", "etymology:", "extra:")
- Ensure JSON is properly formatted

Respond ONLY with the JSON, no additional text.`, input)
}

func parseAIResponse(responseText string) (*AIResponse, error) {
	// Try to extract JSON from the response
	responseText = strings.TrimSpace(responseText)

	// Find JSON boundaries
	start := strings.Index(responseText, "{")
	end := strings.LastIndex(responseText, "}")

	if start == -1 || end == -1 {
		return nil, errors.New("no JSON found in response")
	}

	jsonStr := responseText[start : end+1]

	var aiResponse AIResponse
	err := json.Unmarshal([]byte(jsonStr), &aiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &aiResponse, nil
}

func addCardToAnki(ctx context.Context, cfg *config.Config, card *FinnishCard) error {
	client := ankiclient.NewAnkiConnectClient(cfg.AnkiConnectURL)

	// Check if AnkiConnect is available
	if !client.IsAvailable(ctx) {
		return fmt.Errorf("AnkiConnect is not available at %s", cfg.AnkiConnectURL)
	}

	// Try to get available models to determine the correct field structure
	models, err := client.GetModelNames(ctx)
	if err != nil {
		return fmt.Errorf("failed to get model names: %w", err)
	}

	// Default to Basic model, but prefer a Finnish model if available
	modelName := "Basic"
	for _, model := range models {
		if strings.Contains(strings.ToLower(model), "finnish") {
			modelName = model
			break
		}
	}

	// Get field names for the selected model
	fields, err := client.GetModelFieldNames(ctx, modelName)
	if err != nil {
		log.Printf("Warning: Could not get field names for model %s: %v", modelName, err)
		// Fall back to Basic model fields
		fields = []string{"Front", "Back"}
	}

	// Create field mapping based on available fields
	fieldMap := make(map[string]string)

	// Check if we have the custom Finnish fields from the image
	hasCustomFields := false
	customFieldNames := []string{"Finnish", "Translation", "Finnish Example", "Notes"}
	for _, customField := range customFieldNames {
		for _, field := range fields {
			if field == customField {
				hasCustomFields = true
				break
			}
		}
		if hasCustomFields {
			break
		}
	}

	if hasCustomFields {
		// Use the custom field structure
		fieldMap["Finnish"] = card.Finnish
		fieldMap["Translation"] = card.Translation
		fieldMap["Finnish Example"] = card.FinnishExample
		fieldMap["Notes"] = card.Notes
	} else {
		// Fall back to Basic model structure
		fieldMap["Front"] = card.Finnish
		fieldMap["Back"] = fmt.Sprintf("**Translation:** %s\n\n**Examples:**\n%s\n\n**Notes:**\n%s",
			card.Translation, card.FinnishExample, card.Notes)
	}

	err = client.AddNote(ctx, cfg.AnkiDeck, modelName, fieldMap, []string{"auto-generated", "finnish"})
	if err != nil {
		return fmt.Errorf("failed to add note: %w", err)
	}

	return nil
}
