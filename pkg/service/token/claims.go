package token

// Claims represents extracted claims from a token
type Claims map[string]any

// GetString retrieves a string claim
func (c Claims) GetString(key string) (string, bool) {
	val, ok := c[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// GetInt64 retrieves an int64 claim
func (c Claims) GetInt64(key string) (int64, bool) {
	val, ok := c[key]
	if !ok {
		return 0, false
	}

	switch v := val.(type) {
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case int:
		return int64(v), true
	default:
		return 0, false
	}
}

// GetBool retrieves a bool claim
func (c Claims) GetBool(key string) (bool, bool) {
	val, ok := c[key]
	if !ok {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}

// GetStringSlice retrieves a string slice claim
func (c Claims) GetStringSlice(key string) ([]string, bool) {
	val, ok := c[key]
	if !ok {
		return nil, false
	}

	switch v := val.(type) {
	case []string:
		return v, true
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result, true
	default:
		return nil, false
	}
}

// GetTenantID retrieves the tenant ID from claims
func (c Claims) GetTenantID() (string, bool) {
	return c.GetString("tenant_id")
}

// GetAppID retrieves the app ID from claims
func (c Claims) GetAppID() (string, bool) {
	return c.GetString("app_id")
}

// GetBranchID retrieves the branch ID from claims
func (c Claims) GetBranchID() (string, bool) {
	return c.GetString("branch_id")
}

// GetSubject retrieves the subject (user ID) from claims
func (c Claims) GetSubject() (string, bool) {
	return c.GetString("sub")
}
