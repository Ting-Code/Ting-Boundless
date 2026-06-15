package revocation

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func TestStore_SubjectAndSessionRevocation(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	store := NewStore(rdb)
	ctx := context.Background()

	revoked, err := store.IsSubjectRevoked(ctx, "user-1")
	if err != nil || revoked {
		t.Fatalf("subject before revoke: revoked=%v err=%v", revoked, err)
	}

	if err := store.RevokeSubject(ctx, "user-1"); err != nil {
		t.Fatal(err)
	}
	revoked, err = store.IsSubjectRevoked(ctx, "user-1")
	if err != nil || !revoked {
		t.Fatalf("subject after revoke: revoked=%v err=%v", revoked, err)
	}

	if err := store.RevokeSession(ctx, "sess-abc"); err != nil {
		t.Fatal(err)
	}
	revoked, err = store.IsSessionRevoked(ctx, "sess-abc")
	if err != nil || !revoked {
		t.Fatalf("session after revoke: revoked=%v err=%v", revoked, err)
	}
}

func TestStore_DisabledWhenNoRedis(t *testing.T) {
	store := NewStore(nil)
	ctx := context.Background()
	revoked, err := store.IsSubjectRevoked(ctx, "x")
	if err != nil || revoked {
		t.Fatalf("disabled store: revoked=%v err=%v", revoked, err)
	}
}
