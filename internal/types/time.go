package types

import "time"

func FlattenDuration(duration *time.Duration) any {
	if duration != nil {
		return duration.String()
	}

	return ""
}

func ExpandDuration(data any) (*time.Duration, error) {
	if data == nil || data == "" {
		return nil, nil
	}

	d, err := time.ParseDuration(data.(string))
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func FlattenTime(date *time.Time) any {
	if date != nil {
		return date.Format(time.RFC3339)
	}

	return ""
}

// ExpandTimePtr returns a time pointer for an RFC3339 time.
// It returns nil if time is not valid, you should use validateDate to validate field.
func ExpandTimePtr(i any) *time.Time {
	rawTime := ExpandStringPtr(i)
	if rawTime == nil {
		return nil
	}

	parsedTime, err := time.Parse(time.RFC3339, *rawTime)
	if err != nil {
		return nil
	}

	return &parsedTime
}
