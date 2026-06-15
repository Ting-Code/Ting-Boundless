package identity

// HasRole reports whether id includes the given role name.
func HasRole(id Identity, role string) bool {
	if role == "" {
		return false
	}
	for _, r := range id.Roles {
		if r == role {
			return true
		}
	}
	return false
}
