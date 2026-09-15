package tem_test

import (
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/tem"
	"github.com/stretchr/testify/assert"
)

func TestFlattenMXRecordValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedPriority int
		name             string
		value            string
		expectedExchange string
	}{
		{
			name:             "priority and exchange",
			value:            "10 blackhole.tem.scaleway.com.",
			expectedPriority: 10,
			expectedExchange: "blackhole.tem.scaleway.com.",
		},
		{
			name:             "exchange only",
			value:            "blackhole.tem.scaleway.com.",
			expectedPriority: 0,
			expectedExchange: "blackhole.tem.scaleway.com.",
		},
		{
			name:             "empty",
			value:            "",
			expectedPriority: 0,
			expectedExchange: "",
		},
		{
			name:             "whitespace",
			value:            "  10 blackhole.tem.scaleway.com.  ",
			expectedPriority: 10,
			expectedExchange: "blackhole.tem.scaleway.com.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			priority, exchange := tem.FlattenMXRecordValue(tt.value)
			assert.Equal(t, tt.expectedPriority, priority)
			assert.Equal(t, tt.expectedExchange, exchange)
		})
	}
}
