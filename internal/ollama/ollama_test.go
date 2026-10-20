package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChatPostsMessagesAndParses(t *testing.T) {
	var gotModel string
	var gotMsgs []message
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			var req chatRequest
			json.NewDecoder(r.Body).Decode(&req)
			gotModel, gotMsgs = req.Model, req.Messages
			if req.Stream {
				t.Error("stream must be false")
			}
			w.Write([]byte(`{"message":{"role":"assistant","content":"- Did work [a1b2c3d]"}}`))
			return
		}
		w.Write([]byte(`{"version":"test"}`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	if err := c.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	out, err := c.Chat(context.Background(), "m", "sys", "hello prompt")
	if err != nil {
		t.Fatal(err)
	}
	if out != "- Did work [a1b2c3d]" {
		t.Fatalf("bad response: %q", out)
	}
	if gotModel != "m" || len(gotMsgs) != 2 || gotMsgs[0].Role != "system" || gotMsgs[1].Content != "hello prompt" {
		t.Fatalf("request not passed through: %q %+v", gotModel, gotMsgs)
	}
}

func TestChatServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"model 'nope' not found"}`))
	}))
	defer srv.Close()
	_, err := New(srv.URL).Chat(context.Background(), "nope", "s", "x")
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("want 404 error, got %v", err)
	}
}
