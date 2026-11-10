package qai

// Wire contract for the RAG collection routes, pinned to the gateway handlers
// in routes_rag_collections.go. All tests run against a mock gateway.

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

// searchBody is a response in the exact shape handleRAGCollectionsSearch
// writes: results plus query, collections_searched and request_id.
const searchBody = `{
	"results": [
		{"content":"Goroutines are cheap.","score":0.91,"collection":"go-notes","collection_id":"col_1",
		 "document_id":"doc_7","filename":"concurrency.md","is_shared":false},
		{"content":"Channels synchronise.","score":0.42,"collection":"shared-docs","collection_id":"col_s",
		 "is_shared":true}
	],
	"query": "goroutines",
	"collections_searched": 2,
	"request_id": "req_cs"
}`

func TestCollectionsSearchRequest(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, nil, searchBody, &cap)

	_, err := client.CollectionsSearch(t.Context(), &CollectionSearchRequest{
		Query:         "goroutines",
		CollectionIDs: []string{"col_1", "col_s"},
		MaxChunks:     5,
	})
	if err != nil {
		t.Fatal(err)
	}

	if cap.Method != "POST" || cap.Path != "/qai/v1/rag/collections/search" {
		t.Fatalf("request = %s %s", cap.Method, cap.Path)
	}
	body := cap.bodyJSON(t)
	if got := strings.Join(keys(body), ","); got != "collection_ids,max_chunks,query" {
		t.Fatalf("body keys = %s, want collection_ids,max_chunks,query", got)
	}
	if body["query"] != "goroutines" || body["max_chunks"] != float64(5) {
		t.Fatalf("body = %v", body)
	}
	ids, ok := body["collection_ids"].([]any)
	if !ok || len(ids) != 2 || ids[0] != "col_1" || ids[1] != "col_s" {
		t.Fatalf("collection_ids = %v", body["collection_ids"])
	}
}

// An empty CollectionIDs means "every collection I can read" and a zero
// MaxChunks means "gateway default", so neither may go on the wire.
func TestCollectionsSearchOmitsUnsetFields(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, nil, searchBody, &cap)

	if _, err := client.CollectionsSearch(t.Context(), &CollectionSearchRequest{Query: "goroutines"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(keys(cap.bodyJSON(t)), ","); got != "query" {
		t.Fatalf("body keys = %s, want query", got)
	}
}

func TestCollectionsSearchDecodesResults(t *testing.T) {
	client := newTestClient(t, http.StatusOK, nil, searchBody, nil)

	results, err := client.CollectionsSearch(t.Context(), &CollectionSearchRequest{Query: "goroutines"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	want := CollectionSearchResult{
		Content: "Goroutines are cheap.", Score: 0.91, Collection: "go-notes", CollectionID: "col_1",
		DocumentID: "doc_7", Filename: "concurrency.md", IsShared: false,
	}
	if results[0] != want {
		t.Fatalf("results[0] = %+v, want %+v", results[0], want)
	}
	if !results[1].IsShared || results[1].Collection != "shared-docs" || results[1].DocumentID != "" {
		t.Fatalf("results[1] = %+v", results[1])
	}
}

func TestCollectionsGetDecodesCollectionAndDocuments(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, nil, `{
		"collection": {"id":"col_1","owner":"user_1","provider":"xai","name":"go-notes",
			"provider_collection_id":"xai_col_9","document_count":1,"created_at":"2026-11-01T09:00:00Z"},
		"documents": [{"id":"doc_7","collection_id":"col_1","file_id":"file_3","filename":"concurrency.md",
			"status":"indexed","chunks":0,"uploaded_at":"2026-11-01T09:05:00Z"}]
	}`, &cap)

	got, err := client.CollectionsGet(t.Context(), "col_1")
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "GET" || cap.Path != "/qai/v1/rag/collections/col_1" {
		t.Fatalf("request = %s %s", cap.Method, cap.Path)
	}
	if got.Collection.ID != "col_1" || got.Collection.Name != "go-notes" || got.Collection.Owner != "user_1" ||
		got.Collection.DocumentCount != 1 || got.Collection.CreatedAt != "2026-11-01T09:00:00Z" {
		t.Fatalf("collection = %+v", got.Collection)
	}
	wantDoc := CollectionDocument{
		ID: "doc_7", CollectionID: "col_1", FileID: "file_3", Filename: "concurrency.md",
		Status: "indexed", Chunks: 0, UploadedAt: "2026-11-01T09:05:00Z",
	}
	if len(got.Documents) != 1 || got.Documents[0] != wantDoc {
		t.Fatalf("documents = %+v, want [%+v]", got.Documents, wantDoc)
	}
}

func TestCollectionsDeleteRequest(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, nil, `{"deleted":true,"id":"col_1"}`, &cap)

	if err := client.CollectionsDelete(t.Context(), "col_1"); err != nil {
		t.Fatal(err)
	}
	if cap.Method != "DELETE" || cap.Path != "/qai/v1/rag/collections/col_1" {
		t.Fatalf("request = %s %s", cap.Method, cap.Path)
	}
}

func TestCollectionsUploadSendsFileAndDecodesDocument(t *testing.T) {
	var cap capturedRequest
	client := newTestClient(t, http.StatusOK, map[string]string{"X-QAI-Cost-Ticks": "100"}, `{
		"id":"doc_8","collection_id":"col_1","file_id":"file_4","filename":"notes.txt",
		"status":"indexed","chunks":0,"uploaded_at":"2026-11-10T12:00:00Z"
	}`, &cap)

	doc, err := client.CollectionsUpload(t.Context(), "col_1", "notes.txt", []byte("hello collections"))
	if err != nil {
		t.Fatal(err)
	}

	if cap.Method != "POST" || cap.Path != "/qai/v1/rag/collections/col_1/upload" {
		t.Fatalf("request = %s %s", cap.Method, cap.Path)
	}
	mediaType, params, err := mime.ParseMediaType(cap.ContentType)
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("content type = %q (%v)", cap.ContentType, err)
	}
	part, err := multipart.NewReader(bytes.NewReader(cap.Body), params["boundary"]).NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if part.FormName() != "file" || part.FileName() != "notes.txt" {
		t.Fatalf("part = field %q file %q, want field file", part.FormName(), part.FileName())
	}
	if content, _ := io.ReadAll(part); string(content) != "hello collections" {
		t.Fatalf("part content = %q", content)
	}

	want := CollectionUploadResult{
		ID: "doc_8", CollectionID: "col_1", FileID: "file_4", Filename: "notes.txt",
		Status: "indexed", UploadedAt: "2026-11-10T12:00:00Z",
	}
	if *doc != want {
		t.Fatalf("doc = %+v, want %+v", *doc, want)
	}
}
