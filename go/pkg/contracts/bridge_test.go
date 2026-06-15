package contracts_test

import (
	"testing"
	"time"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/contracts"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/identity"
	commonv1 "github.com/ting-boundless/boundless/gen/go/ting/common/v1"
)

func TestIdentityProtoRoundTrip(t *testing.T) {
	in := identity.Identity{
		RequestID: "r1",
		UserID:    "u1",
		TenantID:  "t1",
		Roles:     []string{"admin", "user"},
		Scopes:    []string{"read"},
		Subject:   "sub-abc",
	}
	out := contracts.IdentityFromProto(contracts.IdentityToProto(in))
	if out.RequestID != in.RequestID || out.UserID != in.UserID || out.TenantID != in.TenantID ||
		out.Subject != in.Subject || !stringSliceEqual(out.Roles, in.Roles) || !stringSliceEqual(out.Scopes, in.Scopes) {
		t.Fatalf("round trip mismatch: %+v vs %+v", out, in)
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestErrorProtoRoundTrip(t *testing.T) {
	in := errs.NotFound("user.not_found", "missing user")
	out := contracts.ErrorFromProto(contracts.ErrorToProto(in, "req-1"), in.Status)
	if out.Code != in.Code || out.Message != in.Message || out.Status != in.Status {
		t.Fatalf("round trip mismatch: %+v vs %+v", out, in)
	}
}

func TestAuditProtoRoundTrip(t *testing.T) {
	in := audit.Event{
		ID:          "ev-1",
		Source:      "gateway",
		Type:        "api.access.denied",
		Subject:     "user:u1",
		Time:        time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC),
		TenantID:    "t1",
		ActorUserID: "u1",
		Data:        map[string]any{"path": "/v1/users/me"},
	}
	pb, err := contracts.AuditToProto(in)
	if err != nil {
		t.Fatal(err)
	}
	out := contracts.AuditFromProto(pb)
	if out.ID != in.ID || out.Source != in.Source || out.Type != in.Type {
		t.Fatalf("mismatch: %+v vs %+v", out, in)
	}
	if !out.Time.Equal(in.Time) {
		t.Fatalf("time mismatch: %v vs %v", out.Time, in.Time)
	}
	if out.Data["path"] != "/v1/users/me" {
		t.Fatalf("data mismatch: %+v", out.Data)
	}
}

func TestProtoIdentityGenerated(t *testing.T) {
	m := &commonv1.IdentityContext{UserId: "u1", RequestId: "r1"}
	if m.GetUserId() != "u1" || m.GetRequestId() != "r1" {
		t.Fatalf("unexpected identity proto: %+v", m)
	}
}

func TestProtoErrorGenerated(t *testing.T) {
	m := &commonv1.ErrorResponse{
		Error: &commonv1.Error{Code: "not_found"},
	}
	if m.GetError().GetCode() != "not_found" {
		t.Fatal()
	}
}

func TestProtoAuditGenerated(t *testing.T) {
	m := &commonv1.AuditEvent{Id: "a1", Type: "test.event"}
	if m.GetId() != "a1" || m.GetType() != "test.event" {
		t.Fatal()
	}
}
