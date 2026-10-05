package tags

import (
	"testing"
)

func TestMergeTags(t *testing.T) {
	tests := []struct {
		name         string
		defaultTags  []string
		resourceTags []string
		want         []string
	}{
		{
			name:         "nil defaults",
			defaultTags:  nil,
			resourceTags: []string{"a", "b"},
			want:         []string{"a", "b"},
		},
		{
			name:         "empty defaults",
			defaultTags:  []string{},
			resourceTags: []string{"a", "b"},
			want:         []string{"a", "b"},
		},
		{
			name:         "nil resource tags",
			defaultTags:  []string{"x", "y"},
			resourceTags: nil,
			want:         []string{"x", "y"},
		},
		{
			name:         "simple merge",
			defaultTags:  []string{"dev"},
			resourceTags: []string{"web"},
			want:         []string{"dev", "web"},
		},
		{
			name:         "duplicate tag removed",
			defaultTags:  []string{"dev", "env:prod"},
			resourceTags: []string{"dev", "web"},
			want:         []string{"dev", "env:prod", "web"},
		},
		{
			name:         "all duplicates",
			defaultTags:  []string{"dev", "web"},
			resourceTags: []string{"dev", "web"},
			want:         []string{"dev", "web"},
		},
		{
			name:         "both empty",
			defaultTags:  []string{},
			resourceTags: []string{},
			want:         []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeTags(tt.defaultTags, tt.resourceTags)
			if !TagsEqual(got, tt.want) {
				t.Errorf("MergeTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveDefaultTags(t *testing.T) {
	tests := []struct {
		name        string
		allTags     []string
		defaultTags []string
		want        []string
	}{
		{
			name:        "nil defaults",
			allTags:     []string{"a", "b"},
			defaultTags: nil,
			want:        []string{"a", "b"},
		},
		{
			name:        "empty defaults",
			allTags:     []string{"a", "b"},
			defaultTags: []string{},
			want:        []string{"a", "b"},
		},
		{
			name:        "remove one default",
			allTags:     []string{"dev", "web"},
			defaultTags: []string{"dev"},
			want:        []string{"web"},
		},
		{
			name:        "remove multiple defaults",
			allTags:     []string{"dev", "env:prod", "web", "api"},
			defaultTags: []string{"dev", "env:prod"},
			want:        []string{"web", "api"},
		},
		{
			name:        "all tags are defaults",
			allTags:     []string{"dev", "web"},
			defaultTags: []string{"dev", "web"},
			want:        []string{},
		},
		{
			name:        "no tags are defaults",
			allTags:     []string{"a", "b"},
			defaultTags: []string{"c", "d"},
			want:        []string{"a", "b"},
		},
		{
			name:        "duplicate in all tags",
			allTags:     []string{"dev", "dev", "web"},
			defaultTags: []string{"dev"},
			want:        []string{"web"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveDefaultTags(tt.allTags, tt.defaultTags)
			if !TagsEqual(got, tt.want) {
				t.Errorf("RemoveDefaultTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagsEqual(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{"both nil", nil, nil, true},
		{"both empty", []string{}, []string{}, true},
		{"nil and empty", nil, []string{}, true},
		{"same order", []string{"a", "b"}, []string{"a", "b"}, true},
		{"different order", []string{"a", "b"}, []string{"b", "a"}, true},
		{"different length", []string{"a", "b"}, []string{"a"}, false},
		{"different values", []string{"a", "b"}, []string{"a", "c"}, false},
		{"duplicates", []string{"a", "a", "b"}, []string{"a", "b", "b"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TagsEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("TagsEqual(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDefaultConfig_MergeTags(t *testing.T) {
	dc := &DefaultConfig{Tags: []string{"dev"}}

	tests := []struct {
		name         string
		dc           *DefaultConfig
		resourceTags []string
		want         []string
	}{
		{"nil config", nil, []string{"a"}, []string{"a"}},
		{"with config", dc, []string{"web"}, []string{"dev", "web"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dc.MergeTags(tt.resourceTags)
			if !TagsEqual(got, tt.want) {
				t.Errorf("MergeTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig_RemoveDefaultTags(t *testing.T) {
	dc := &DefaultConfig{Tags: []string{"dev"}}

	tests := []struct {
		name    string
		dc      *DefaultConfig
		allTags []string
		want    []string
	}{
		{"nil config", nil, []string{"a"}, []string{"a"}},
		{"with config", dc, []string{"dev", "web"}, []string{"web"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dc.RemoveDefaultTags(tt.allTags)
			if !TagsEqual(got, tt.want) {
				t.Errorf("RemoveDefaultTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
