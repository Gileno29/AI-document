import pytest
from pydantic import ValidationError
from schemas import IngestRequest, IngestResponse, RetrieveRequest, RetrieveResponse, TOP_K


class TestIngestRequest:
    def test_valid(self):
        req = IngestRequest(doc_id="doc-1", chunks=["chunk a", "chunk b"])
        assert req.doc_id == "doc-1"
        assert len(req.chunks) == 2

    def test_missing_doc_id(self):
        with pytest.raises(ValidationError):
            IngestRequest(chunks=["chunk a"])

    def test_missing_chunks(self):
        with pytest.raises(ValidationError):
            IngestRequest(doc_id="doc-1")

    def test_chunks_must_be_list(self):
        with pytest.raises(ValidationError):
            IngestRequest(doc_id="doc-1", chunks="not a list")

    def test_empty_chunks_list(self):
        req = IngestRequest(doc_id="doc-1", chunks=[])
        assert req.chunks == []


class TestIngestResponse:
    def test_valid(self):
        resp = IngestResponse(doc_id="doc-1", ingested=5)
        assert resp.ingested == 5

    def test_ingested_must_be_int(self):
        with pytest.raises(ValidationError):
            IngestResponse(doc_id="doc-1", ingested="five")


class TestRetrieveRequest:
    def test_valid(self):
        req = RetrieveRequest(doc_id="doc-1", query="what is this?")
        assert req.top_k == TOP_K

    def test_custom_top_k(self):
        req = RetrieveRequest(doc_id="doc-1", query="what is this?", top_k=3)
        assert req.top_k == 3

    def test_missing_query(self):
        with pytest.raises(ValidationError):
            RetrieveRequest(doc_id="doc-1")

    def test_top_k_must_be_int(self):
        with pytest.raises(ValidationError):
            RetrieveRequest(doc_id="doc-1", query="test", top_k="five")


class TestRetrieveResponse:
    def test_valid(self):
        resp = RetrieveResponse(doc_id="doc-1", chunks=["a", "b"])
        assert len(resp.chunks) == 2

    def test_empty_chunks(self):
        resp = RetrieveResponse(doc_id="doc-1", chunks=[])
        assert resp.chunks == []