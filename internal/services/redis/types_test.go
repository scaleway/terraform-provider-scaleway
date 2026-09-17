//nolint:testpackage // Tests need access to unexported redis connection helpers.
package redis

import (
	"net"
	"net/url"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
)

func assertRedisConnectionString(t *testing.T, endpoints []*redis.Endpoint, userName, password string, tlsEnabled bool, want string) {
	t.Helper()

	got := redisConnectionString(endpoints, userName, password, tlsEnabled)
	if got != want {
		t.Fatalf("redisConnectionString() = %q, want %q", got, want)
	}
}

func TestRedisConnectionString(t *testing.T) {
	t.Parallel()

	// When a password is present, userinfo uses user_name + password (Redis ACL). When password is empty,
	// userinfo is omitted (e.g. password_wo).

	pubIP := net.ParseIP("51.158.1.2")
	privIP := net.ParseIP("10.0.0.5")

	t.Run("empty endpoints", func(t *testing.T) {
		t.Parallel()
		assertRedisConnectionString(t, nil, "", "", false, "")
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
		assertRedisConnectionString(t, endpoints, "redisuser", "secret", true, "rediss://redisuser:secret@51.158.1.2:6379/0")
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
		assertRedisConnectionString(t, endpoints, "redisuser", "p", false, "redis://redisuser:p@10.0.0.5:6380/0")
	})

	t.Run("password with ampersand (escaped in userinfo per url.URL)", func(t *testing.T) {
		t.Parallel()

		endpoints := []*redis.Endpoint{
			{
				Port:          6379,
				PublicNetwork: &redis.PublicNetwork{},
				IPs:           []net.IP{pubIP},
			},
		}
		want := &url.URL{
			Scheme: "redis",
			Host:   net.JoinHostPort("51.158.1.2", "6379"),
			Path:   "/0",
		}
		want.User = url.UserPassword("redisuser", "a&b")
		assertRedisConnectionString(t, endpoints, "redisuser", "a&b", false, want.String())
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
		assertRedisConnectionString(t, endpoints, "ignored", "", false, "redis://51.158.1.2:6379/0")
	})
}
