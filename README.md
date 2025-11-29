# April Fools Joke Detector

A Go tool that analyzes articles to determine if they are April Fools jokes or pranks using Large Language Model (LLM) analysis.

## Features

- Fetches article content from any URL
- Extracts clean text from web pages
- Uses OpenAI's GPT-4 to analyze content for April Fools indicators
- Provides detailed analysis with confidence levels and reasoning
- Single binary executable - no runtime dependencies

## Requirements

- Go 1.21 or higher
- OpenAI API key

## Installation

1. Clone this repository:
```bash
git clone <your-repo-url>
cd poisson
```

2. Install dependencies:
```bash
go mod download
```

3. Build the binary:
```bash
go build -o detector
```

Or install directly:
```bash
go install
```

4. Set up your OpenAI API key:
```bash
# On Windows (PowerShell)
$env:OPENAI_API_KEY="your-api-key-here"

# On Linux/Mac
export OPENAI_API_KEY="your-api-key-here"
```

Or create a `.env` file (not included in repo for security):
```
OPENAI_API_KEY=your-api-key-here
```

## Usage

### Basic Usage

```bash
./detector <article-url>
```

Or if installed via `go install`:
```bash
detector <article-url>
```

### With API Key as Argument

```bash
./detector -api-key your-api-key-here <article-url>
```

### Verbose Output

```bash
./detector -verbose <article-url>
```

### Example

```bash
./detector https://example.com/article
```

## How It Works

1. **Content Fetching**: The tool fetches the article from the provided URL and extracts the main text content, removing scripts, styles, and other non-content elements.

2. **LLM Analysis**: The extracted content is sent to OpenAI's GPT-4 model with a carefully crafted prompt that asks it to analyze:
   - Unrealistic or absurd claims
   - Publication date (April 1st indicators)
   - Tone and style
   - References to April Fools
   - Context clues

3. **Results**: The LLM provides:
   - A yes/no/uncertain verdict
   - Confidence level (0-100)
   - Detailed reasoning
   - Key indicators

## Configuration

The tool uses GPT-4 by default. You can modify the model in `main.go` by changing the `Model` field in the `OpenAIRequest` struct within the `analyzeWithLLM` function.

## Limitations

- Requires an active internet connection
- Requires a valid OpenAI API key (usage incurs costs)
- Content is truncated to 8000 characters for LLM context limits
- Some websites may block automated requests
- Analysis quality depends on the LLM's understanding of context

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See LICENSE file for details.

## Disclaimer

This tool is for entertainment and educational purposes. The analysis is based on AI interpretation and may not always be accurate. Always verify information from reliable sources.

