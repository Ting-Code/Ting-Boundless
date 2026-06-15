package upload

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/file-service/internal/store"
)

func TestHandler_RequiresFileField(t *testing.T) {
	h := New(Config{
		Files: storeStub{},
		S3:    &s3Stub{enabled: true},
		Log:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/files/", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{UserID: "u1", RequestID: "r1"}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandler_RequiresIdentity(t *testing.T) {
	h := New(Config{S3: &s3Stub{enabled: true}})

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/files/", nil))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestSanitizeFilename(t *testing.T) {
	if sanitizeFilename("../../etc/passwd") != "passwd" {
		t.Fatal("expected basename only")
	}
	if sanitizeFilename("") != "upload" {
		t.Fatal("empty -> upload")
	}
}

type storeStub struct{}

func (storeStub) Insert(context.Context, store.File) (store.File, error) {
	return store.File{}, nil
}

type s3Stub struct {
	enabled bool
}

func (s *s3Stub) Enabled() bool { return s.enabled }
func (s *s3Stub) Bucket() string {
	return "test-bucket"
}
func (s *s3Stub) PutObject(context.Context, string, io.Reader, int64, string) error {
	return nil
}
