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
	raw, ok := data.([]any)
	if !ok || data == nil {
		return []string{}
	}

	expandedIDs := make([]string, 0, len(raw))
	for _, s := range raw {
		if s == nil {
			s = ""
		}

		expandedIDs = append(expandedIDs, ExpandID(s.(string)))
	}

	return expandedIDs
}
