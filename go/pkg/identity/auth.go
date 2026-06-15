package identity

// Authenticated reports whether the identity has a platform user id.
func (id Identity) Authenticated() bool {
	return id.UserID != ""
}
