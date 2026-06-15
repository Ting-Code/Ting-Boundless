package auth

import (
	"context"
	"log/slog"

	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/services/gateway/internal/identityresolve"
)

// mapPlatformUserID resolves Logto OIDC subjects to platform user_ids via auth-service.
// Tokens that already carry a distinct user_id claim (e.g. auth-service issuer) are unchanged.
func mapPlatformUserID(
	ctx context.Context,
	resolver *identityresolve.Client,
	id identity.Identity,
	requestID string,
) identity.Identity {
	if resolver == nil || id.Subject == "" {
		return id
	}
	if id.UserID != "" && id.UserID != id.Subject {
		return id
	}

	userID, err := resolver.ResolveLogto(ctx, id.Subject, requestID)
	if err != nil {
		slog.Default().Warn("logto identity resolve failed; using subject as user_id",
			slog.String("subject", id.Subject),
			slog.Any("error", err),
		)
		return id
	}
	id.UserID = userID
	return id
}
