package httpx

import "testing"

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"/v1/business/items", "/v1/business/items"},
		{"/v1/business/items/6y5waiahu40n921lfb3zt", "/v1/business/items/:id"},
		{"/v1/users/550e8400-e29b-41d4-a716-446655440000", "/v1/users/:id"},
		{"/admin/items/42", "/admin/items/:id"},
	}
	for _, tc := range tests {
		if got := NormalizePath(tc.in); got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
}
