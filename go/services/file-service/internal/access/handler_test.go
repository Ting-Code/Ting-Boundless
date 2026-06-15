package access

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/file-service/internal/store"
)

func TestMetaHandler_ForbiddenForOtherOwner(t *testing.T) {
	h := NewMeta(Config{
		Files: &filesStub{
			row: store.File{ID: "f1", OwnerID: "owner-a", TenantID: "t1"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/files/f1", nil)
	req.SetPathValue("id", "f1")
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "owner-b", TenantID: "t1", RequestID: "r1",
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestMetaHandler_ReturnsMetadata(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	h := NewMeta(Config{
		Files: &filesStub{
			row: store.File{
				ID: "f1", OwnerID: "u1", TenantID: "t1",
				Bucket: "b", ObjectKey: "t1/u1/f1/readme.md",
				ContentType: "text/plain", SizeBytes: 12, CreatedAt: created,
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/files/f1", nil)
	req.SetPathValue("id", "f1")
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", TenantID: "t1", RequestID: "r1",
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestMetaHandler_NotFound(t *testing.T) {
	h := NewMeta(Config{Files: &filesStub{err: pgx.ErrNoRows}})

	req := httptest.NewRequest(http.MethodGet, "/v1/files/missing", nil)
	req.SetPathValue("id", "missing")
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", RequestID: "r1",
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestCanAccess(t *testing.T) {
	row := store.File{OwnerID: "u1", TenantID: "t1"}
	if !canAccess(identity.Identity{UserID: "u1", TenantID: "t1"}, row) {
		t.Fatal("owner in tenant should access")
	}
	if canAccess(identity.Identity{UserID: "u2", TenantID: "t1"}, row) {
		t.Fatal("other user denied")
	}
	if canAccess(identity.Identity{UserID: "u1", TenantID: "t2"}, row) {
		t.Fatal("other tenant denied")
	}
}

type filesStub struct {
	row store.File
	err error
}

func (f *filesStub) GetByID(context.Context, string) (store.File, error) {
	if f.err != nil {
		return store.File{}, f.err
	}
	return f.row, nil
}

func (f *filesStub) ListByOwner(context.Context, string, string, int) ([]store.File, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.row.ID != "" {
		return []store.File{f.row}, nil
	}
	return nil, nil
}

func (f *filesStub) DeleteByID(context.Context, string, string, string) error {
	if f.err != nil {
		return f.err
	}
	return nil
}

type s3GetStub struct {
	enabled bool
	body    string
}

func (s *s3GetStub) Enabled() bool { return s.enabled }

func (s *s3GetStub) GetObject(context.Context, string) (*http.Response, error) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusOK)
	_, _ = rec.WriteString(s.body)
	return rec.Result(), nil
}

func (s *s3GetStub) PresignGetURL(string, time.Duration) (string, time.Time, error) {
	return "http://signed.example/obj", time.Now().UTC().Add(time.Hour), nil
}

func (s *s3GetStub) DeleteObject(context.Context, string) error {
	return nil
}

func TestDownloadHandler_StreamsBody(t *testing.T) {
	h := NewDownload(Config{
		Files: &filesStub{
			row: store.File{
				ID: "f1", OwnerID: "u1", ObjectKey: "t/u1/f1/a.txt",
				ContentType: "text/plain", SizeBytes: 5,
			},
		},
		S3: &s3GetStub{enabled: true, body: "hello"},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/files/f1/download", nil)
	req.SetPathValue("id", "f1")
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", RequestID: "r1",
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK || rr.Body.String() != "hello" {
		t.Fatalf("status=%d body=%q", rr.Code, rr.Body.String())
	}
}

func TestURLHandler_ReturnsPresignedURL(t *testing.T) {
	h := NewURL(Config{
		Files: &filesStub{
			row: store.File{ID: "f1", OwnerID: "u1", ObjectKey: "t/u1/f1/a.txt"},
		},
		S3:            &s3GetStub{enabled: true},
		DefaultExpiry: time.Hour,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/files/f1/url", nil)
	req.SetPathValue("id", "f1")
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", RequestID: "r1",
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "signed.example") {
		t.Fatalf("body=%s", rr.Body.String())
	}
}

func TestDeleteHandler_RemovesFile(t *testing.T) {
	h := NewDelete(Config{
		Files: &filesStub{
			row: store.File{ID: "f1", OwnerID: "u1", TenantID: "t1", ObjectKey: "t/u1/f1/a.txt"},
		},
		S3: &s3GetStub{enabled: true},
	})

	req := httptest.NewRequest(http.MethodDelete, "/v1/files/f1", nil)
	req.SetPathValue("id", "f1")
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", TenantID: "t1", RequestID: "r1",
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestDeleteHandler_ForbiddenForOtherOwner(t *testing.T) {
	h := NewDelete(Config{
		Files: &filesStub{
			row: store.File{ID: "f1", OwnerID: "owner-a", ObjectKey: "k"},
		},
		S3: &s3GetStub{enabled: true},
	})

	req := httptest.NewRequest(http.MethodDelete, "/v1/files/f1", nil)
	req.SetPathValue("id", "f1")
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "owner-b", RequestID: "r1",
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rr.Code)
	}
}
