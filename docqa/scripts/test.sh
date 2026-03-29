# upload a document
curl -F "file=@mydoc.pdf" http://localhost:8080/upload

# ask a question (use the doc_id returned above)
curl -X POST http://localhost:8080/ask \
  -H "Content-Type: application/json" \
  -d '{"doc_id":"mydoc-1","question":"What is the main topic?"}'



# upload
curl -F "file=@example.txt" http://localhost:8080/upload
{"question": "What is RAG and how does it work?"}
{"question": "What are the three types of machine learning?"}
{"question": "What are the ethical concerns about AI?"}