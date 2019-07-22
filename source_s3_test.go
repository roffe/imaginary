package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const s3URL = "s3://testdata/large.jpg"

func TestS3ImageSourceNotAllowedOrigin(t *testing.T) {
	origin, _ := url.Parse("http://foo")
	origins := []*url.URL{origin}
	source := NewS3ImageSource(&SourceConfig{AllowedOrigins: origins})

	fakeHandler := func(w http.ResponseWriter, r *http.Request) {
		if !source.Matches(r) {
			t.Fatal("Cannot match the request")
		}

		_, err := source.GetImage(r)
		if err == nil {
			t.Fatal("Error cannot be empty")
		}

		if err.Error() != "not allowed remote URL origin: bar.com" {
			t.Fatalf("Invalid error message: %s", err)
		}
	}

	r, _ := http.NewRequest(http.MethodGet, "http://foo/bar?url=s3://bar.com", nil)
	w := httptest.NewRecorder()
	fakeHandler(w, r)
}

func TestS3ImageSourceError(t *testing.T) {
	var err error

	s3InvalidURL := "s3://invalid/file"

	source := NewS3ImageSource(&SourceConfig{})
	fakeHandler := func(w http.ResponseWriter, r *http.Request) {
		if !source.Matches(r) {
			t.Fatal("Cannot match the request")
		}

		_, err = source.GetImage(r)
		if err == nil {
			t.Fatalf("Server response should not be valid: %s", err)
		}
	}

	r, _ := http.NewRequest(http.MethodGet, "http://foo/bar?url="+s3InvalidURL, nil)
	w := httptest.NewRecorder()
	fakeHandler(w, r)
}

func TestS3ImageSourceExceedsMaximumAllowedLength(t *testing.T) {
	var body []byte
	var err error

	source := NewS3ImageSource(&SourceConfig{
		MaxAllowedSize: 1023,
	})
	fakeHandler := func(w http.ResponseWriter, r *http.Request) {
		if !source.Matches(r) {
			t.Fatal("Cannot match the request")
		}

		body, err = source.GetImage(r)
		if err == nil {
			t.Fatalf("It should not allow a request to image exceeding maximum allowed size: %s", err)
		}
		w.Write(body)
	}

	r, _ := http.NewRequest(http.MethodGet, "http://foo/bar?url="+s3URL, nil)
	w := httptest.NewRecorder()
	fakeHandler(w, r)
}
