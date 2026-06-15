package contracts

import (
	"github.com/ting-boundless/boundless/pkg/errs"
	commonv1 "github.com/ting-boundless/boundless/gen/go/ting/common/v1"
)

// ErrorToProto maps pkg/errs to the wire proto envelope (HTTP status is not in proto).
func ErrorToProto(e *errs.Error, requestID string) *commonv1.ErrorResponse {
	if e == nil {
		return nil
	}
	return &commonv1.ErrorResponse{
		Error: &commonv1.Error{
			Code:      e.Code,
			Message:   e.Message,
			RequestId: requestID,
		},
	}
}

// ErrorFromProto extracts pkg/errs from proto. Status defaults to 500 when unknown.
func ErrorFromProto(m *commonv1.ErrorResponse, status int) *errs.Error {
	if m == nil || m.GetError() == nil {
		return nil
	}
	inner := m.GetError()
	if status <= 0 {
		status = 500
	}
	return errs.New(status, inner.GetCode(), inner.GetMessage())
}
