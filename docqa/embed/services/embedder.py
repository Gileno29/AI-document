import os
from sentence_transformers import SentenceTransformer
import chromadb

class EmbedderService:
    def __init__(self):
        model_name = os.getenv("MODEL_NAME", "all-MiniLM-L6-v2")
        self.model = SentenceTransformer(model_name)
        self.chroma = chromadb.HttpClient(
            host=os.getenv("CHROMA_HOST", "chroma"),
            port=int(os.getenv("CHROMA_PORT", "8000")),
        )

    
    def ingest(self, doc_id: str, chunks: list[str]) -> int:
        embeddings = self.model.encode(chunks).tolist()
        ids = [f"{doc_id}_{i}" for i in range(len(chunks))]
        col = self._collection(doc_id)
        col.upsert(embeddings=embeddings, documents=chunks, ids=ids)
        return len(chunks)

    def retrieve(self, doc_id: str, query: str, top_k: int) -> list[str]:
        query_emb = self.model.encode([query]).tolist()
        col = self._collection(doc_id)
        results = col.query(
            query_embeddings=query_emb,
            n_results=min(top_k, col.count() or 1),
        )
        return results["documents"][0] if results["documents"] else []

    def _collection(self, doc_id: str):
        safe = doc_id.replace("-", "_")
        return self.chroma.get_or_create_collection(f"doc_{safe}")