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

// GetMapValue returns the value for a key from a map.
// returns zero value if key does not exist in map.
func GetMapValue[T any]( //nolint:ireturn
	m map[string]any,
	key string,
) T {
	var val T
	valI, exists := m[key]
	if exists {
		val = valI.(T)
	}
	return val
}
