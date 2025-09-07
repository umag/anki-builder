package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	AIProvider     string `json:"aiProvider"` // "gemini" or "openai"
	GeminiAPIKey   string `json:"geminiApiKey"`
	OpenAIAPIKey   string `json:"openaiApiKey"`
	AnkiDeck       string `json:"ankiDeck"`
	AnkiConnectURL string `json:"ankiConnectUrl"`
}

type FinnishCard struct {
	Finnish        string
	Translation    string
	FinnishExample string
	Notes          string
}

type AIResponse struct {
	Translations []string `json:"translations"`
	Examples     []string `json:"examples"`
	Notes        string   `json:"notes"`
}

type AnkiConnectRequest struct {
	Action  string                 `json:"action"`
	Version int                    `json:"version"`
	Params  map[string]interface{} `json:"params"`
}

type AnkiConnectResponse struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
}

//nolint:forbidigo // fmt prints are cute
func main() {
	config := loadConfig()

	fmt.Println("Finnish Anki Card Builder")
	fmt.Printf("Using AI Provider: %s\n", config.AIProvider)
	fmt.Println("Enter Finnish words or phrases (type 'quit' to exit):")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "" {
			continue
		}

		fmt.Printf("Processing: %s\n", input)

		// Generate card data using AI
		card, err := generateCard(config, input)
		if err != nil {
			fmt.Printf("Error generating card: %v\n", err)
			continue
		}

		// Add card to Anki
		err = addCardToAnki(config, card)
		if err != nil {
			fmt.Printf("Error adding card to Anki: %v\n", err)
			continue
		}

		fmt.Printf("✅ Successfully added card for '%s'\n\n", input)
	}
}

func loadConfig() *Config {
	var cfg Config
	cfgBytes, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Failed to read config.json: %v", err)
	}

	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse config.json: %v", err)
	}

	// Validate configuration
	if cfg.AIProvider == "gemini" && cfg.GeminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required when using Gemini")
	}
	if cfg.AIProvider == "openai" && cfg.OpenAIAPIKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required when using OpenAI")
	}

	return &cfg
}

func generateCard(config *Config, input string) (*FinnishCard, error) {
	var aiResponse *AIResponse
	var err error

	switch config.AIProvider {
	case "gemini":
		aiResponse, err = generateWithGemini(config, input)
	case "openai":
		aiResponse, err = generateWithOpenAI(config, input)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", config.AIProvider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query AI provider: %w", err)
	}

	// Create card from AI response
	card := &FinnishCard{
		Finnish:        input,
		Translation:    strings.Join(aiResponse.Translations, "; "),
		FinnishExample: strings.Join(aiResponse.Examples, "\n\n"),
		Notes:          aiResponse.Notes,
	}

	return card, nil
}

func generateWithGemini(config *Config, input string) (*AIResponse, error) {
	ctx := context.Background()
	client := NewGeminiClient(config.GeminiAPIKey)

	prompt := buildPrompt(input)

	responseText, err := client.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content with Gemini: %w", err)
	}

	return parseAIResponse(responseText)
}

func generateWithOpenAI(config *Config, input string) (*AIResponse, error) {
	ctx := context.Background()
	client := NewOpenAIClient(config.OpenAIAPIKey)

	prompt := buildPrompt(input)

	responseText, err := client.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content with OpenAI: %w", err)
	}

	return parseAIResponse(responseText)
}

func buildPrompt(input string) string {
	return fmt.Sprintf(`You are a Finnish language expert helping to create Anki flashcards for language learners.

For the Finnish word/phrase: "%s"

Please provide a JSON response with the following structure:
{
  "translations": ["translation1", "translation2", ...],
  "examples": [
    "Example sentence 1 in Finnish",
    "Example sentence 2 in Finnish",
    "Example sentence 3 in Finnish",
    "Example sentence 4 in Finnish"
  ],
  "notes": "Word origin, synonyms, grammatical information, usage quirks, and any other useful information for language learners"
}

Guidelines:
- Provide 1-3 translations (most common meanings)
- Create 3-4 example sentences at B1-B2 level
- Use the word in different grammatical cases/forms when possible
- Include etymology, synonyms, grammatical notes, and usage tips in the notes section
- Make examples natural and contextually rich
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

func addCardToAnki(config *Config, card *FinnishCard) error {
	ctx := context.Background()
	client := NewAnkiConnectClient(config.AnkiConnectURL)

	// Check if AnkiConnect is available
	if !client.IsAvailable(ctx) {
		return fmt.Errorf("AnkiConnect is not available at %s", config.AnkiConnectURL)
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

	// Add the card to Anki
	err = client.AddNote(ctx, config.AnkiDeck, modelName, fieldMap, []string{"auto-generated", "finnish"})
	if err != nil {
		return fmt.Errorf("failed to add note: %w", err)
	}

	return nil
}
