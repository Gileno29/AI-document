# AIDocument — AI Document Q&A API

A production-ready REST API that lets you upload documents and ask questions about them in natural language. Built with **Go** (API gateway) and **Python** (embedding + retrieval), using a full RAG pipeline under the hood.

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go)
![Python](https://img.shields.io/badge/Python-3.11-3776AB?style=flat-square&logo=python)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker)
![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)

---

## How it works

```
POST /upload  →  extract text  →  chunk  →  embed  →  store in ChromaDB
POST /ask     →  embed query   →  top-k retrieval  →  LLM  →  answer
```

1. You upload a `.pdf` or `.txt` file
2. The Go API extracts the text and splits it into overlapping chunks
3. The Python service generates vector embeddings with `sentence-transformers`
4. Embeddings are stored in ChromaDB
5. On a question, the query is embedded and matched against stored chunks
6. The top matching chunks are sent as context to an LLM, which generates the answer

---

## Architecture

```
┌─────────────┐     POST /upload      ┌──────────────────┐
│   Client    │ ───────────────────▶  │  Go API Gateway  │
│  (curl etc) │ ◀───────────────────  │   :8080          │
└─────────────┘     POST /ask         └────────┬─────────┘
                                               │
                          ┌────────────────────┼────────────────────┐
                          ▼                    ▼                    ▼
                  ┌──────────────┐   ┌──────────────────┐  ┌──────────────┐
                  │  Object      │   │  Python Embed    │  │  Ollama LLM  │
                  │  Storage     │   │  Service  :8000  │  │  :11434      │
                  │  (local)     │   │  FastAPI         │  │  llama3.2    │
                  └──────────────┘   └────────┬─────────┘  └──────────────┘
                                              │
                                     ┌────────▼─────────┐
                                     │    ChromaDB      │
                                     │  Vector Store    │
                                     │    :8001         │
                                     └──────────────────┘
```

---

## Stack

| Layer | Technology |
|---|---|
| API gateway | Go 1.22 + chi router |
| Embedding & retrieval | Python 3.11 + FastAPI + sentence-transformers |
| Vector database | ChromaDB |
| LLM (default) | Ollama + llama3.2 — runs locally, free |
| LLM (alternative) | OpenAI `gpt-4o-mini` or Anthropic `claude-haiku` |
| Infrastructure | Docker Compose |

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose v2
- 4GB+ free disk space (for the llama3.2 model)

---

## Quickstart

```bash
# 1. clone the repo
git clone https://github.com/your-username/docqa
cd docqa

# 2. configure environment
cp .env.example .env

# 3. start all services (first run downloads llama3.2 ~2GB)
docker compose up --build

# 4. wait for all services to be healthy (~2-3 min on first run)
docker compose ps
```

All five services will start in the correct order:

```
docqa-chroma     → healthy
docqa-embed      → healthy  (downloads sentence-transformers model on first run)
docqa-ollama     → healthy
docqa-ollama-pull → exits 0 (pulls llama3.2 once, then done)
docqa-api        → healthy
```

---

## API Reference

### `GET /health`

Check that all services are running.

```bash
curl http://localhost:8080/health
```

```json
{ "status": "ok" }
```

---

### `POST /upload`

Upload a document. Accepts `.pdf` or `.txt`.

```bash
curl -F "file=@document.pdf" http://localhost:8080/upload
```

```json
{
  "doc_id": "document-1",
  "chunks": 24
}
```

---

### `POST /ask`

Ask a question about an uploaded document.

```bash
curl -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d '{"doc_id": "document-1", "question": "What is the main topic?"}'
```

```json
{
  "doc_id": "document-1",
  "question": "What is the main topic?",
  "answer": "The document covers...",
  "sources": 5
}
```

---

## Try it with the example file

An `example.txt` about artificial intelligence and RAG is included in the repo.

```bash
# upload
curl -F "file=@example.txt" http://localhost:8080/upload

# use the returned doc_id in the questions below
curl -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d '{"doc_id":"example-1","question":"What is RAG and how does it work?"}'

curl -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d '{"doc_id":"example-1","question":"What are the three types of machine learning?"}'

curl -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d '{"doc_id":"example-1","question":"What are the ethical concerns about AI?"}'
```

---

## LLM providers

The LLM backend is configurable via the `.env` file. No code changes needed.

### Ollama — free, local (default)

```env
LLM_PROVIDER=ollama
OLLAMA_MODEL=llama3.2
```

The model is pulled automatically on first `docker compose up`. Subsequent starts use the cached volume.

To use a different model:

```bash
# in .env
OLLAMA_MODEL=mistral

# pull it
docker exec -it docqa-ollama ollama pull mistral
```

### OpenAI

```env
LLM_PROVIDER=openai
OPENAI_API_KEY=sk-...
```

Uses `gpt-4o-mini` by default (~$0.15 per 1M input tokens).

### Anthropic

```env
LLM_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-ant-...
```

Uses `claude-haiku-4-5` by default.

---

## Project structure

```
docqa/
├── docker-compose.yml
├── .env.example
├── example.txt
├── api/                          # Go service
│   ├── Dockerfile
│   ├── go.mod
│   └── cmd/
│       └── main.go               # server entry, graceful shutdown
│   └── internal/
│       ├── handlers/
│       │   └── handlers.go       # /health  /upload  /ask
│       ├── python/
│       │   └── client.go         # HTTP client → embed service
│       ├── llm/
│       │   └── client.go         # Ollama + OpenAI + Anthropic
│       └── text/
│           └── text.go           # PDF extraction + chunker
└── embed/                        # Python service
    ├── Dockerfile
    ├── requirements.txt
    └── main.py                   # FastAPI /ingest /retrieve /health
```

---

## Useful commands

```bash
# view logs for a specific service
docker logs -f docqa-api
docker logs -f docqa-embed
docker logs -f docqa-ollama-pull

# check all services status
docker compose ps

# stop everything
docker compose down

# stop and delete all data (vectors, uploads, model cache)
docker compose down -v

# rebuild a single service after a code change
docker compose up --build api
```

---

## License

MIT