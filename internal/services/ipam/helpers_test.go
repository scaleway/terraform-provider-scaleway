package ipam

import (
	"testing"
)

func TestDiffSuppress_IPAMIP(t *testing.T) {
	tests := []struct {
		name     string
		oldValue string
		newValue string
		want     bool
	}{
		{
			name:     "IP == IP",
			oldValue: "172.16.32.7",
			newValue: "172.16.32.7",
			want:     true,
		},
		{
			name:     "IP == IP/CIDR same host",
			oldValue: "172.16.32.7/22",
			newValue: "172.16.32.7",
			want:     true,
		},
		{
			name:     "IP/CIDR == IP same host (reversed)",
			oldValue: "172.16.32.7",
			newValue: "172.16.32.7/22",
			want:     true,
		},
		{
			name:     "Different host within same CIDR is NOT suppressed",
			oldValue: "172.16.32.7/22",
			newValue: "172.16.32.8",
			want:     false,
		},
		{
			name:     "Different host (plain IPs) is NOT suppressed",
			oldValue: "172.16.32.7",
			newValue: "172.16.32.8",
			want:     false,
		},
		{
			name:     "Equal but with /32 single-host CIDR",
			oldValue: "10.0.0.1/32",
			newValue: "10.0.0.1",
			want:     true,
		},
		{
			name:     "Broader CIDR vs network address should NOT suppress",
			oldValue: "10.0.0.1/24",
			newValue: "10.0.0.0",
			want:     false,
		},
		{
			name:     "Invalid old value -> not suppressed",
			oldValue: "not-an-ip",
			newValue: "10.0.0.1",
			want:     false,
		},
		{
			name:     "Invalid new value -> not suppressed",
			oldValue: "10.0.0.1",
			newValue: "bad/32",
			want:     false,
		},
		{
			name:     "Both invalid -> not suppressed",
			oldValue: "nope",
			newValue: "nope2",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diffSuppressFuncStandaloneIPandCIDR("", tt.oldValue, tt.newValue, nil)
			if got != tt.want {
				t.Fatalf("diffSuppress(%q, %q) = %v, want %v", tt.oldValue, tt.newValue, got, tt.want)
			}
		})
	}
}
