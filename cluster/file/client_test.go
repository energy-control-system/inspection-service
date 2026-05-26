package file

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sunshineOfficial/golib/goctx"
	"github.com/sunshineOfficial/golib/gohttp"
	"github.com/sunshineOfficial/golib/pagination"
)

func TestGetByIDsPassesForwardedHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(ForwardedHostHeader) != "api.example.test" {
			t.Fatalf("%s = %q, want %q", ForwardedHostHeader, r.Header.Get(ForwardedHostHeader), "api.example.test")
		}
		if r.Header.Get(ForwardedProtoHeader) != "https" {
			t.Fatalf("%s = %q, want %q", ForwardedProtoHeader, r.Header.Get(ForwardedProtoHeader), "https")
		}

		if got := r.URL.Query()["id"]; len(got) != 2 || got[0] != "10" || got[1] != "20" {
			t.Fatalf("id query = %+v, want [10 20]", got)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]File{{ID: 10, URL: "https://api.example.test/storage/a.jpg"}}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(gohttp.NewClient(), server.URL)

	files, err := client.GetByIDs(goctx.Wrap(t.Context()), []int{10, 20}, pagination.Pagination{}, ForwardedHeaders{
		Host:  "api.example.test",
		Proto: "https",
	})
	if err != nil {
		t.Fatalf("GetByIDs returned error: %v", err)
	}

	if len(files) != 1 || files[0].URL != "https://api.example.test/storage/a.jpg" {
		t.Fatalf("files = %+v, want URL %q", files, "https://api.example.test/storage/a.jpg")
	}
}

func TestUploadPassesForwardedHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(ForwardedHostHeader) != "api.example.test" {
			t.Fatalf("%s = %q, want %q", ForwardedHostHeader, r.Header.Get(ForwardedHostHeader), "api.example.test")
		}
		if r.Header.Get(ForwardedProtoHeader) != "https" {
			t.Fatalf("%s = %q, want %q", ForwardedProtoHeader, r.Header.Get(ForwardedProtoHeader), "https")
		}

		if err := r.ParseMultipartForm(1024); err != nil {
			t.Fatalf("ParseMultipartForm returned error: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(File{ID: 10, URL: "https://api.example.test/storage/a.jpg"}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(gohttp.NewClient(), server.URL)

	file, err := client.Upload(goctx.Wrap(t.Context()), "a.jpg", bytes.NewReader([]byte("image")), ForwardedHeaders{
		Host:  "api.example.test",
		Proto: "https",
	})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	if file.URL != "https://api.example.test/storage/a.jpg" {
		t.Fatalf("file.URL = %q, want %q", file.URL, "https://api.example.test/storage/a.jpg")
	}
}
