package access

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/file-service/internal/store"
)

type listStub struct {
	rows []store.File
}

func (s *listStub) GetByID(context.Context, string) (store.File, error) {
	return store.File{}, nil
}

func (s *listStub) ListByOwner(context.Context, string, string, int) ([]store.File, error) {
	return s.rows, nil
}

func (s *listStub) DeleteByID(context.Context, string, string, string) error {
	return nil
}

func TestListHandler_ReturnsFiles(t *testing.T) {
	created := time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC)
	h := NewList(Config{
		Files: &listStub{
			rows: []store.File{{
				ID: "f1", OwnerID: "u1", TenantID: "t1",
				ObjectKey: "k", ContentType: "text/plain", SizeBytes: 3, CreatedAt: created,
			}},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/files/", nil)
	req = req.WithContext(identity.NewContext(req.Context(), identity.Identity{
		UserID: "u1", TenantID: "t1", RequestID: "r1",
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var body struct {
		Files []map[string]any `json:"files"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Files) != 1 || body.Files[0]["file_id"] != "f1" {
		t.Fatalf("body=%+v", body)
	}
}
