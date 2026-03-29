# upload a document
curl -F "file=@mydoc.pdf" http://localhost:8080/upload

# ask a question (use the doc_id returned above)
curl -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d '{"doc_id":"mydoc-1","question":"What is the main topic?"}'
