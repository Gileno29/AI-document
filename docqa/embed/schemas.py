from pydantic import BaseModel
import os

TOP_K       = int(os.getenv("TOP_K", "5"))

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