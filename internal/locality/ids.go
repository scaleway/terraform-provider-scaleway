package locality

// ExpandID returns the id whether it is a localizedID or a raw ID.
func ExpandID(id interface{}) string {
	_, ID, err := ParseLocalizedID(id.(string))
	if err != nil {
		return id.(string)
	}

	return ID
}

func ExpandIDs(data interface{}) []string {
	expandedIDs := make([]string, 0, len(data.([]interface{})))

	for _, s := range data.([]interface{}) {
		if s == nil {
			s = ""
		}

		expandedID := ExpandID(s.(string))
		expandedIDs = append(expandedIDs, expandedID)
	}

	return expandedIDs
}
