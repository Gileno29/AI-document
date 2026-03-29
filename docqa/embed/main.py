import os
import logging
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
import chromadb

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)

app = FastAPI(title="docqa-embed")

CHROMA_HOST = os.getenv("CHROMA_HOST", "localhost")
CHROMA_PORT = int(os.getenv("CHROMA_PORT", "8001"))
MODEL_NAME  = os.getenv("MODEL_NAME", "all-MiniLM-L6-v2")
TOP_K       = int(os.getenv("TOP_K", "5"))

log.info("loading model %s …", MODEL_NAME)
model = SentenceTransformer(MODEL_NAME)
log.info("model loaded")

chroma = chromadb.HttpClient(host=CHROMA_HOST, port=CHROMA_PORT)


def get_collection(doc_id: str):
    safe = doc_id.replace("-", "_")
    return chroma.get_or_create_collection(name=f"doc_{safe}")


# ── request/response schemas ─────────────────────────────────────────────────

class IngestRequest(BaseModel):
    doc_id: str
    chunks: list[str]

class IngestResponse(BaseModel):
    doc_id: str
    ingested: int

class RetrieveRequest(BaseModel):
    doc_id: str
    query: str
    top_k: int = TOP_K

class RetrieveResponse(BaseModel):
    doc_id: str
    chunks: list[str]


# ── endpoints ────────────────────────────────────────────────────────────────

@app.get("/health")
def health():
    return {"status": "ok", "model": MODEL_NAME}


@app.post("/ingest", response_model=IngestResponse)
def ingest(req: IngestRequest):
    if not req.chunks:
        raise HTTPException(status_code=422, detail="chunks must not be empty")

    log.info("ingesting doc_id=%s chunks=%d", req.doc_id, len(req.chunks))
    embeddings = model.encode(req.chunks, show_progress_bar=False).tolist()
    ids = [f"{req.doc_id}_{i}" for i in range(len(req.chunks))]

    col = get_collection(req.doc_id)
    col.upsert(embeddings=embeddings, documents=req.chunks, ids=ids)

    log.info("ingested doc_id=%s", req.doc_id)
    return IngestResponse(doc_id=req.doc_id, ingested=len(req.chunks))


@app.post("/retrieve", response_model=RetrieveResponse)
def retrieve(req: RetrieveRequest):
    log.info("retrieve doc_id=%s query='%s'", req.doc_id, req.query[:60])
    query_emb = model.encode([req.query], show_progress_bar=False).tolist()

    col = get_collection(req.doc_id)
    results = col.query(
        query_embeddings=query_emb,
        n_results=min(req.top_k, col.count() or 1),
    )
    chunks = results["documents"][0] if results["documents"] else []
    return RetrieveResponse(doc_id=req.doc_id, chunks=chunks)
