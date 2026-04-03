package redis

import (
	"net"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
)

func TestRedisConnectionString(t *testing.T) {
	t.Parallel()

	pubIP := net.ParseIP("51.158.1.2")
	privIP := net.ParseIP("10.0.0.5")

	tests := []struct {
		name       string
		endpoints  []*redis.Endpoint
		password   string
		tlsEnabled bool
		want       string
	}{
		{
			name: "empty endpoints",
			want: "",
		},
		{
			name: "public preferred over private",
			endpoints: []*redis.Endpoint{
				{
					Port:           6379,
					PrivateNetwork: &redis.PrivateNetwork{},
					IPs:            []net.IP{privIP},
				},
				{
					Port:          6379,
					PublicNetwork: &redis.PublicNetwork{},
					IPs:           []net.IP{pubIP},
				},
			},
			password:   "secret",
			tlsEnabled: true,
			want:       "rediss://:secret@51.158.1.2:6379/0",
		},
		{
			name: "private only",
			endpoints: []*redis.Endpoint{
				{
					Port:           6380,
					PrivateNetwork: &redis.PrivateNetwork{},
					IPs:            []net.IP{privIP},
				},
			},
			password:   "p",
			tlsEnabled: false,
			want:       "redis://:p@10.0.0.5:6380/0",
		},
		{
			name: "password with ampersand (sub-delimiter allowed unescaped in userinfo)",
			endpoints: []*redis.Endpoint{
				{
					Port:          6379,
					PublicNetwork: &redis.PublicNetwork{},
					IPs:           []net.IP{pubIP},
				},
			},
			password:   "a&b",
			tlsEnabled: false,
			want:       "redis://:a&b@51.158.1.2:6379/0",
		},
		{
			name: "no password",
			endpoints: []*redis.Endpoint{
				{
					Port:          6379,
					PublicNetwork: &redis.PublicNetwork{},
					IPs:           []net.IP{pubIP},
				},
			},
			tlsEnabled: false,
			want:       "redis://51.158.1.2:6379/0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := redisConnectionString(tt.endpoints, tt.password, tt.tlsEnabled)
			if got != tt.want {
				t.Fatalf("redisConnectionString() = %q, want %q", got, tt.want)
			}
		})
	}
}
