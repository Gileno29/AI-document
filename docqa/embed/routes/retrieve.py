from fastapi import APIRouter
from schemas import RetrieveRequest, RetrieveResponse
from services.embedder import EmbedderService

router = APIRouter()
embedder = EmbedderService()

@router.post("/retrieve", response_model=RetrieveResponse)
def retrieve(req: RetrieveRequest):
    chunks = embedder.retrieve(req.doc_id, req.query, req.top_k)
    return RetrieveResponse(doc_id=req.doc_id, chunks=chunks)