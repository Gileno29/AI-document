import pytest
from fastapi.testclient import TestClient
from unittest.mock import MagicMock, patch


@pytest.fixture
def client():
    with patch("routes.ingest.EmbedderService") as mock_ingest_svc, \
         patch("routes.retrieve.EmbedderService") as mock_retrieve_svc:

        ingest_instance = MagicMock()
        ingest_instance.ingest.return_value = 3
        mock_ingest_svc.return_value = ingest_instance

        retrieve_instance = MagicMock()
        retrieve_instance.retrieve.return_value = ["chunk a", "chunk b"]
        mock_retrieve_svc.return_value = retrieve_instance

        from main import app
        yield TestClient(app)


class TestHealth:
    def test_returns_200(self, client):
        resp = client.get("/health")
        assert resp.status_code == 200
        assert resp.json()["status"] == "ok"


class TestIngestRoute:
    def test_valid_request(self, client):
        resp = client.post("/ingest", json={
            "doc_id": "doc-1",
            "chunks": ["chunk a", "chunk b", "chunk c"]
        })
        assert resp.status_code == 200
        assert resp.json()["ingested"] == 3
        assert resp.json()["doc_id"] == "doc-1"

    def test_empty_chunks_returns_422(self, client):
        resp = client.post("/ingest", json={
            "doc_id": "doc-1",
            "chunks": []
        })
        assert resp.status_code == 422

    def test_missing_doc_id_returns_422(self, client):
        resp = client.post("/ingest", json={"chunks": ["a"]})
        assert resp.status_code == 422

    def test_missing_chunks_returns_422(self, client):
        resp = client.post("/ingest", json={"doc_id": "doc-1"})
        assert resp.status_code == 422


class TestRetrieveRoute:
    def test_valid_request(self, client):
        resp = client.post("/retrieve", json={
            "doc_id": "doc-1",
            "query": "what is this about?"
        })
        assert resp.status_code == 200
        assert resp.json()["chunks"] == ["chunk a", "chunk b"]
        assert resp.json()["doc_id"] == "doc-1"

    def test_missing_query_returns_422(self, client):
        resp = client.post("/retrieve", json={"doc_id": "doc-1"})
        assert resp.status_code == 422

    def test_custom_top_k(self, client):
        resp = client.post("/retrieve", json={
            "doc_id": "doc-1",
            "query": "test",
            "top_k": 3
        })
        assert resp.status_code == 200