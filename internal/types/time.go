package types

import "time"

func FlattenDuration(duration *time.Duration) interface{} {
	if duration != nil {
		return duration.String()
	}
	return ""
}

func ExpandDuration(data interface{}) (*time.Duration, error) {
	if data == nil || data == "" {
		return nil, nil
	}
	d, err := time.ParseDuration(data.(string))
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func FlattenTime(date *time.Time) interface{} {
	if date != nil {
		return date.Format(time.RFC3339)
	}
	return ""
}
