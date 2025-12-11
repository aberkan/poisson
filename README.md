# Joke News Article Detector

A Go tool that analyzes articles to determine if they are jokes or pranks using Large Language Model (LLM) analysis.  It takes a URL of an article or RSS feed and outputs the analysis.

## Features

- Fetches article content from any URL or RSS URL
- Extracts clean text from web pages
- Uses OpenAI's GPT-4 to analyze content for April Fools indicators
- Provides detailed analysis with confidence levels and reasoning
- Single binary executable - no runtime dependencies

## Requirements

- Go 1.21 or higher
- OpenAI API key
- Access to the datastore

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

## License

See LICENSE file for details.

## Disclaimer

This tool is for entertainment and educational purposes. The analysis is based on AI interpretation and may not always be accurate. Always verify information from reliable sources.

