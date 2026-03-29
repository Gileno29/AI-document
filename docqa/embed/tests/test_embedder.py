import pytest
from unittest.mock import MagicMock, patch
import numpy as np


@pytest.fixture
def mock_embedder():
    with patch("services.embedder.SentenceTransformer") as mock_model_cls, \
         patch("services.embedder.chromadb.HttpClient") as mock_chroma_cls:

        mock_model = MagicMock()
        mock_model.encode.return_value = np.array([[0.1, 0.2, 0.3], [0.4, 0.5, 0.6]])
        mock_model_cls.return_value = mock_model

        mock_chroma = MagicMock()
        mock_chroma_cls.return_value = mock_chroma

        from services.embedder import EmbedderService
        service = EmbedderService()
        service._chroma = mock_chroma
        service._model = mock_model

        yield service, mock_model, mock_chroma


class TestIngest:
    def test_ingest_returns_chunk_count(self, mock_embedder):
        service, mock_model, mock_chroma = mock_embedder

        mock_col = MagicMock()
        mock_chroma.get_or_create_collection.return_value = mock_col

        result = service.ingest("doc-1", ["chunk a", "chunk b"])

        assert result == 2
        mock_col.upsert.assert_called_once()

    def test_ingest_calls_encode(self, mock_embedder):
        service, mock_model, mock_chroma = mock_embedder
        mock_chroma.get_or_create_collection.return_value = MagicMock()

        service.ingest("doc-1", ["hello", "world"])

        mock_model.encode.assert_called_once_with(["hello", "world"])

    def test_ingest_generates_correct_ids(self, mock_embedder):
        service, mock_model, mock_chroma = mock_embedder

        mock_col = MagicMock()
        mock_chroma.get_or_create_collection.return_value = mock_col

        service.ingest("doc-1", ["a", "b"])

        call_kwargs = mock_col.upsert.call_args[1]
        assert call_kwargs["ids"] == ["doc-1_0", "doc-1_1"]

    def test_ingest_sanitizes_doc_id_for_collection(self, mock_embedder):
        service, _, mock_chroma = mock_embedder
        mock_chroma.get_or_create_collection.return_value = MagicMock()

        service.ingest("my-doc-123", ["chunk"])

        mock_chroma.get_or_create_collection.assert_called_with("doc_my_doc_123")


class TestRetrieve:
    def test_retrieve_returns_chunks(self, mock_embedder):
        service, mock_model, mock_chroma = mock_embedder

        mock_col = MagicMock()
        mock_col.count.return_value = 10
        mock_col.query.return_value = {"documents": [["chunk one", "chunk two"]]}
        mock_chroma.get_or_create_collection.return_value = mock_col

        mock_model.encode.return_value = np.array([[0.1, 0.2, 0.3]])

        result = service.retrieve("doc-1", "what is this?", top_k=2)

        assert result == ["chunk one", "chunk two"]

    def test_retrieve_empty_collection_returns_empty(self, mock_embedder):
        service, mock_model, mock_chroma = mock_embedder

        mock_col = MagicMock()
        mock_col.count.return_value = 0
        mock_col.query.return_value = {"documents": []}
        mock_chroma.get_or_create_collection.return_value = mock_col

        mock_model.encode.return_value = np.array([[0.1, 0.2, 0.3]])

        result = service.retrieve("doc-1", "what is this?", top_k=5)
        assert result == []

    def test_retrieve_respects_top_k(self, mock_embedder):
        service, mock_model, mock_chroma = mock_embedder

        mock_col = MagicMock()
        mock_col.count.return_value = 10
        mock_col.query.return_value = {"documents": [["a", "b", "c"]]}
        mock_chroma.get_or_create_collection.return_value = mock_col

        mock_model.encode.return_value = np.array([[0.1, 0.2]])

        service.retrieve("doc-1", "query", top_k=3)

        call_kwargs = mock_col.query.call_args[1]
        assert call_kwargs["n_results"] == 3