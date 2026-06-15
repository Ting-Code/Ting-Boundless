package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
)

func TestWriteError_MatchesContractShape(t *testing.T) {
	rr := httptest.NewRecorder()
	httpx.WriteError(rr, "req-1", errs.NotFound("file_not_found", "file not found"))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rr.Code)
	}
	var body struct {
		Error struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "file_not_found" || body.Error.RequestID != "req-1" {
		t.Fatalf("body=%+v", body)
	}
}
