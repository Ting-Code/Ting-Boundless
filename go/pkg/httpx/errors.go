package httpx

import (
	"net/http"

	"github.com/ting-boundless/boundless/pkg/contracts"
	"github.com/ting-boundless/boundless/pkg/errs"
)

// WriteError serializes a unified error envelope aligned with platform-contracts.
func WriteError(w http.ResponseWriter, requestID string, err error) {
	e, ok := err.(*errs.Error)
	if !ok {
		e = errs.Internal("internal", "internal server error")
	}

	resp := contracts.ErrorToProto(e, requestID)
	if resp == nil || resp.GetError() == nil {
		errs.Write(w, requestID, e)
		return
	}

	inner := resp.GetError()
	body := map[string]any{
		"error": map[string]any{
			"code":    inner.GetCode(),
			"message": inner.GetMessage(),
		},
	}
	if rid := inner.GetRequestId(); rid != "" {
		body["error"].(map[string]any)["request_id"] = rid
	}
	JSON(w, e.Status, body)
}
