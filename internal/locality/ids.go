package locality

// ExpandID returns the id whether it is a localizedID or a raw ID.
func ExpandID(id interface{}) string {
	_, ID, err := ParseLocalizedID(id.(string))
	if err != nil {
		return id.(string)
	}
	return ID
}
