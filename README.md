# Finnish Anki Card Builder

A Go application that generates Finnish vocabulary cards for Anki using AI (Gemini) and adds them via AnkiConnect.

## Features

- Interactive console input for Finnish words/phrases
- Support for Gemini API
- Generates comprehensive vocabulary cards with:
  - Multiple translations
  - 3-4 example sentences at B1-B2 level
  - Etymology, synonyms, and usage notes
- Automatically adds cards to Anki via AnkiConnect
- Supports custom field structures (Finnish, Translation, Finnish Example, Notes)

## Prerequisites

1. **Anki** with **AnkiConnect** addon installed
2. API key for **Google Gemini** 

## Setup

1. Clone or download this project
2. Copy `config.json.example` to `config.json` and fill in your configuration:
   ```bash
   cp config.json.example config.json
   ```
3. Edit `config.json` with your settings:

## Installation

1. Build the application:
   ```bash
   go build -o anki-builder
   ```

## Usage

1. Make sure Anki is running with AnkiConnect enabled
2. Run the application:
   ```bash
   ./anki-builder
   ```
   Or directly with Go:
   ```bash
   go run .
   ```

3. Enter Finnish words or phrases when prompted
4. Type `quit` to exit

## AnkiConnect Setup

1. Install the AnkiConnect addon in Anki
2. Ensure AnkiConnect is configured to allow connections from localhost
3. The application will automatically detect your deck's field structure

## Card Structure

The application supports two field structures:

### Custom Finnish Model (Recommended)
- **Finnish**: The input word/phrase
- **Translation**: Multiple translations separated by semicolons
- **Finnish Example**: 3-4 example sentences
- **Notes**: Etymology, synonyms, usage notes

### Basic Model (Fallback)
- **Front**: The Finnish word/phrase
- **Back**: Combined translation, examples, and notes

## API Keys

### Google Gemini
1. Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Create a new API key
3. Add it to your `config` file

## Example Usage

```
Finnish Anki Card Builder
Using AI Provider: gemini
Enter Finnish words or phrases (type 'quit' to exit):
> kissa
Processing: kissa
✅ Successfully added card for 'kissa'

> mennä kauppaan
Processing: mennä kauppaan
✅ Successfully added card for 'mennä kauppaan'

> quit
Goodbye!
```

## License

MIT License
