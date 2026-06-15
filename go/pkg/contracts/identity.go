package contracts

import (
	"github.com/ting-boundless/boundless/pkg/identity"
	commonv1 "github.com/ting-boundless/boundless/gen/go/ting/common/v1"
)

// IdentityToProto maps pkg/identity to the generated proto message.
func IdentityToProto(id identity.Identity) *commonv1.IdentityContext {
	return &commonv1.IdentityContext{
		RequestId: id.RequestID,
		UserId:    id.UserID,
		TenantId:  id.TenantID,
		Roles:     append([]string(nil), id.Roles...),
		Scopes:    append([]string(nil), id.Scopes...),
		Subject:   id.Subject,
	}
}

// IdentityFromProto maps the generated proto message to pkg/identity.
func IdentityFromProto(m *commonv1.IdentityContext) identity.Identity {
	if m == nil {
		return identity.Identity{}
	}
	return identity.Identity{
		RequestID: m.GetRequestId(),
		UserID:    m.GetUserId(),
		TenantID:  m.GetTenantId(),
		Roles:     append([]string(nil), m.GetRoles()...),
		Scopes:    append([]string(nil), m.GetScopes()...),
		Subject:   m.GetSubject(),
	}
}
