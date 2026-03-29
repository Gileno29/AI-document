import os
import logging
from fastapi import FastAPI
from routes.ingest import router as ingest_router
from routes.retrieve import router as retrieve_router

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)


app = FastAPI(title="docqa-embed")

app.include_router(ingest_router)
app.include_router(retrieve_router)

@app.get("/health")
def health():
    return {"status": "ok"}