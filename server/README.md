# GraphQL Server

A GraphQL server for the Poisson project that provides access to article analysis and crawled page data.

## Features

- GraphQL API for querying analysis results and crawled pages
- Mutation endpoint for analyzing new URLs
- Health check endpoint
- Ready for Google Cloud Run deployment

## Endpoints

- `POST /graphql` - GraphQL endpoint
- `GET /health` - Health check endpoint

## GraphQL Schema

### Queries

- `health: String!` - Health check
- `analysis(url: String!, mode: String): AnalysisResult` - Get analysis result for a URL
- `crawledPage(url: String!): CrawledPage` - Get crawled page for a URL

### Mutations

- `analyze(url: String!, mode: String): AnalysisResult!` - Analyze a URL and return the result

## Environment Variables

- `PORT` - Server port (default: 8080, Cloud Run sets this automatically)
- `GOOGLE_CLOUD_PROJECT` - Google Cloud project ID (default: "poisson-berkan")
- `OPENAI_API_KEY` - OpenAI API key for analysis

## Local Development

```bash
# Build and run
go run server/cmd/main.go

# Or build first
go build -o server ./server/cmd
./server
```

## Testing with curl

### Health Check
```bash
curl http://localhost:8080/health
```

### GraphQL Health Query
```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "query { health }"}'
```

### Query Analysis Result
```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { analysis(url: \"https://example.com/article\", mode: \"joke\") { mode jokePercentage jokeReasoning promptFingerprint } }"
  }'
```

### Query Crawled Page
```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { crawledPage(url: \"https://example.com/article\") { url title content datetime } }"
  }'
```

### Analyze URL (Mutation)
```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { analyze(url: \"https://example.com/article\", mode: \"joke\") { mode jokePercentage jokeReasoning promptFingerprint } }"
  }'
```

## Docker Build

```bash
# Build Docker image
docker build -f server/Dockerfile -t poisson-server .

# Run locally
docker run -p 8080:8080 -e OPENAI_API_KEY=your_key poisson-server
```

## Cloud Run Deployment

The server is configured to run on Google Cloud Run. Use the provided Dockerfile to build and deploy:

```bash
# Build and push to Google Container Registry
gcloud builds submit --tag gcr.io/poisson-berkan/poisson-server --file server/Dockerfile .

# Deploy to Cloud Run
gcloud run deploy poisson-server \
  --image gcr.io/poisson-berkan/poisson-server \
  --platform managed \
  --region us-central1 \
  --set-env-vars OPENAI_API_KEY=your_key
```

## Example Queries

### Health Check
```graphql
query {
  health
}
```

### Get Analysis
```graphql
query {
  analysis(url: "https://example.com/article", mode: "joke") {
    mode
    jokePercentage
    jokeReasoning
    promptFingerprint
  }
}
```

### Analyze URL
```graphql
mutation {
  analyze(url: "https://example.com/article", mode: "joke") {
    mode
    jokePercentage
    jokeReasoning
    promptFingerprint
  }
}
```


