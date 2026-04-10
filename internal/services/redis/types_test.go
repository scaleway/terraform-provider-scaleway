//nolint:testpackage // Tests need access to unexported redis connection helpers.
package redis

import (
	"net"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
)

func assertRedisConnectionString(t *testing.T, endpoints []*redis.Endpoint, password string, tlsEnabled bool, want string) {
	t.Helper()

	got := redisConnectionString(endpoints, password, tlsEnabled)
	if got != want {
		t.Fatalf("redisConnectionString() = %q, want %q", got, want)
	}
}

func TestRedisConnectionString(t *testing.T) {
	t.Parallel()

	// Expected URIs use the Redis convention for password-only ACL auth: an empty username before the
	// password (redis(s)://:password@host:port/0), matching url.UserPassword("", password) in production code.

	pubIP := net.ParseIP("51.158.1.2")
	privIP := net.ParseIP("10.0.0.5")

	t.Run("empty endpoints", func(t *testing.T) {
		t.Parallel()
		assertRedisConnectionString(t, nil, "", false, "")
	})

	t.Run("public preferred over private", func(t *testing.T) {
		t.Parallel()

		endpoints := []*redis.Endpoint{
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
		}
		assertRedisConnectionString(t, endpoints, "secret", true, "rediss://:secret@51.158.1.2:6379/0")
	})

	t.Run("private only", func(t *testing.T) {
		t.Parallel()

		endpoints := []*redis.Endpoint{
			{
				Port:           6380,
				PrivateNetwork: &redis.PrivateNetwork{},
				IPs:            []net.IP{privIP},
			},
		}
		assertRedisConnectionString(t, endpoints, "p", false, "redis://:p@10.0.0.5:6380/0")
	})

	t.Run("password with ampersand (sub-delimiter allowed unescaped in userinfo)", func(t *testing.T) {
		t.Parallel()

		endpoints := []*redis.Endpoint{
			{
				Port:          6379,
				PublicNetwork: &redis.PublicNetwork{},
				IPs:           []net.IP{pubIP},
			},
		}
		assertRedisConnectionString(t, endpoints, "a&b", false, "redis://:a&b@51.158.1.2:6379/0")
	})

	t.Run("no password", func(t *testing.T) {
		t.Parallel()

		endpoints := []*redis.Endpoint{
			{
				Port:          6379,
				PublicNetwork: &redis.PublicNetwork{},
				IPs:           []net.IP{pubIP},
			},
		}
		assertRedisConnectionString(t, endpoints, "", false, "redis://51.158.1.2:6379/0")
	})
}
