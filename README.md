# Nederlands Anki Card Builder

A Go application that generates Nederlands vocabulary cards for Anki using AI
(Gemini) and adds them via AnkiConnect.

## Features

- Interactive console input for Nederlands words/phrases
- Support for Gemini API
- Generates comprehensive vocabulary cards with:
  - Multiple translations
  - 3-4 example sentences at B1-B2 level
  - Etymology, synonyms, and usage notes
- Automatically adds cards to Anki via AnkiConnect
- Supports custom field structures (Nederlands, Translation, Nederlands Example,
  Notes)

## Prerequisites

1. **Anki** with **AnkiConnect** addon installed
2. API key for **Google Gemini**
3. **Nix** with flakes enabled (for development environment)
4. **direnv** (recommended for automatic environment loading)

## Development Environment Setup (Recommended)

This project uses Nix flakes and direnv for a reproducible development
environment.

### Option 1: Using Nix Flakes + direnv (Recommended)

1. Install [Nix](https://nixos.org/download.html) with flakes enabled
2. Install [direnv](https://direnv.net/docs/installation.html)
3. Clone this project and navigate to the directory
4. Allow direnv to load the environment:
   ```bash
   direnv allow
   ```
5. The development environment will be automatically loaded with Go and all
   necessary tools

### Option 2: Using Nix Flakes directly

1. Install [Nix](https://nixos.org/download.html) with flakes enabled
2. Clone this project and navigate to the directory
3. Enter the development shell:
   ```bash
   nix develop
   ```

### Option 3: Traditional Go setup

1. Install [Go 1.23+](https://golang.org/dl/)
2. Ensure Go is in your PATH

## Setup

1. Clone or download this project
2. Copy `config.json.example` to `config.json` and fill in your configuration:
   ```bash
   cp config.json.example config.json
   ```
3. Edit `config.json` with your settings
4. (Optional) Copy `.env.example` to `.env` for additional environment variables

## Installation

### With Nix (in development shell)

```bash
nix build
```

### With Go directly

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

3. Enter Nederlands words or phrases when prompted
4. Type `quit` to exit

## AnkiConnect Setup

1. Install the AnkiConnect addon in Anki
2. Ensure AnkiConnect is configured to allow connections from localhost
3. The application will automatically detect your deck's field structure

## Card Structure

The application supports two field structures:

### Custom Nederlands Model (Recommended)

- **Nederlands**: The input word/phrase
- **Translation**: Multiple translations separated by semicolons
- **Nederlands Example**: 3-4 example sentences
- **Notes**: Etymology, synonyms, usage notes

### Basic Model (Fallback)

- **Front**: The Nederlands word/phrase
- **Back**: Combined translation, examples, and notes

## API Keys

### Google Gemini

1. Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Create a new API key
3. Add it to your `config` file

## Example Usage

```
Nederlands Anki Card Builder
Using AI Provider: gemini
Enter Nederlands words or phrases (type 'quit' to exit):
> huis
Processing: huis
✅ Successfully added card for 'huis'

> naar de winkel gaan
Processing: naar de winkel gaan
✅ Successfully added card for 'naar de winkel gaan'

> quit
Goodbye!
```
