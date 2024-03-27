package types

func FlattenMap(m map[string]string) interface{} {
	if m == nil {
		return nil
	}
	flattenedMap := make(map[string]interface{})
	for k, v := range m {
		flattenedMap[k] = v
	}
	return flattenedMap
}

func FlattenMapStringStringPtr(m map[string]*string) interface{} {
	if m == nil {
		return nil
	}
	flattenedMap := make(map[string]interface{})
	for k, v := range m {
		if v != nil {
			flattenedMap[k] = *v
		} else {
			flattenedMap[k] = ""
		}
	}
	return flattenedMap
}

func ExpandMapPtrStringString(data interface{}) *map[string]string {
	if data == nil {
		return nil
	}
	m := make(map[string]string)
	for k, v := range data.(map[string]interface{}) {
		m[k] = v.(string)
	}
	return &m
}

func ExpandMapStringStringPtr(data interface{}) map[string]*string {
	if data == nil {
		return nil
	}
	m := make(map[string]*string)
	for k, v := range data.(map[string]interface{}) {
		m[k] = ExpandStringPtr(v)
	}
	return m
}

func ExpandMapStringString(data any) map[string]string {
	if data == nil {
		return nil
	}
	m := make(map[string]string)
	for k, v := range data.(map[string]interface{}) {
		m[k] = v.(string)
	}
	return m
}
