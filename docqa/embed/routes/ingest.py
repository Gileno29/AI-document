from fastapi import APIRouter, HTTPException
from schemas import IngestRequest, IngestResponse
from services.embedder import EmbedderService

router = APIRouter()
embedder = EmbedderService()

@router.post("/ingest", response_model=IngestResponse)
def ingest(req: IngestRequest):
    if not req.chunks:
        raise HTTPException(status_code=422, detail="chunks must not be empty")
    ingested = embedder.ingest(req.doc_id, req.chunks)
    return IngestResponse(doc_id=req.doc_id, ingested=ingested)