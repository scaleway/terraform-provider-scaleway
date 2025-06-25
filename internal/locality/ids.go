package locality

// ExpandID returns the id whether it is a localizedID or a raw ID.
func ExpandID(id any) string {
	_, ID, err := ParseLocalizedID(id.(string))
	if err != nil {
		return id.(string)
	}

	return ID
}

func ExpandIDs(data any) []string {
	expandedIDs := make([]string, 0, len(data.([]any)))

	for _, s := range data.([]any) {
		if s == nil {
			s = ""
		}

		expandedID := ExpandID(s.(string))
		expandedIDs = append(expandedIDs, expandedID)
	}

	return expandedIDs
}
