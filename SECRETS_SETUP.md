# Secrets Setup

This project uses a `secrets/` directory to store sensitive information like API keys. This directory is excluded from git.

## Setup Instructions

1. Create the secrets directory:
   ```bash
   mkdir secrets
   ```

2. Create a file `secrets/openai_key` with your OpenAI API key:
   ```bash
   echo "your-api-key-here" > secrets/openai_key
   ```

   Or on Windows (PowerShell):
   ```powershell
   "your-api-key-here" | Out-File -FilePath secrets\openai_key -NoNewline
   ```

3. The file should contain only your API key (no extra whitespace or newlines).

## Priority Order

The application will look for the API key in this order:
1. `--api-key` command-line flag
2. `secrets/openai_key` file
3. `OPENAI_API_KEY` environment variable

## Security Note

- Never commit the `secrets/` directory to git
- The `secrets/` directory is already in `.gitignore`
- Keep your API keys secure and never share them publicly

