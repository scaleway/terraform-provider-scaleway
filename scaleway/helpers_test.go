package scaleway

import (
	"fmt"
	"net"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

<<<<<<< HEAD
func TestGetRandomName(t *testing.T) {
	name := newRandomName("test")
	assert.True(t, strings.HasPrefix(name, "tf-test-"))
=======
func TestIsHTTPCodeError(t *testing.T) {
	assert.True(t, isHTTPCodeError(&scw.ResponseError{StatusCode: http.StatusBadRequest}, http.StatusBadRequest))
	assert.False(t, isHTTPCodeError(nil, http.StatusBadRequest))
	assert.False(t, isHTTPCodeError(&scw.ResponseError{StatusCode: http.StatusBadRequest}, http.StatusNotFound))
	assert.False(t, isHTTPCodeError(errors.New("not an http error"), http.StatusNotFound))
}

func TestIs404Error(t *testing.T) {
	assert.True(t, is404Error(&scw.ResponseError{StatusCode: http.StatusNotFound}))
	assert.False(t, is404Error(nil))
	assert.False(t, is404Error(&scw.ResponseError{StatusCode: http.StatusBadRequest}))
}

func TestIs403Error(t *testing.T) {
	assert.True(t, is403Error(&scw.ResponseError{StatusCode: http.StatusForbidden}))
	assert.False(t, is403Error(nil))
	assert.False(t, is403Error(&scw.ResponseError{StatusCode: http.StatusBadRequest}))
>>>>>>> 8b7b03a4 (chore: migrate types to its own package)
}

func testCheckResourceAttrFunc(name string, key string, test func(string) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}
		value, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("key not found: %s", key)
		}
		err := test(value)
		if err != nil {
			return fmt.Errorf("test for %s %s did not pass test: %s", name, key, err)
		}
		return nil
	}
}

var UUIDRegex = regexp.MustCompile(`[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}`)

func testCheckResourceAttrUUID(name string, key string) resource.TestCheckFunc {
	return resource.TestMatchResourceAttr(name, key, UUIDRegex)
}

func testCheckResourceAttrIPv4(name string, key string) resource.TestCheckFunc {
	return testCheckResourceAttrFunc(name, key, func(value string) error {
		ip := net.ParseIP(value)
		if ip.To4() == nil {
			return fmt.Errorf("%s is not a valid IPv4", value)
		}
		return nil
	})
}

func testCheckResourceAttrIPv6(name string, key string) resource.TestCheckFunc {
	return testCheckResourceAttrFunc(name, key, func(value string) error {
		ip := net.ParseIP(value)
		if ip.To16() == nil {
			return fmt.Errorf("%s is not a valid IPv6", value)
		}
		return nil
	})
}

func testCheckResourceAttrIP(name string, key string) resource.TestCheckFunc {
	return testCheckResourceAttrFunc(name, key, func(value string) error {
		ip := net.ParseIP(value)
		if ip == nil {
			return fmt.Errorf("%s is not a valid IP", value)
		}
		return nil
	})
}

func TestStringHashcode(t *testing.T) {
	v := "hello, world"
	expected := StringHashcode(v)
	for i := 0; i < 100; i++ {
		actual := StringHashcode(v)
		if actual != expected {
			t.Fatalf("bad: %#v\n\t%#v", actual, expected)
		}
	}
}

func TestStringHashcode_positiveIndex(t *testing.T) {
	// "2338615298" hashes to uint32(2147483648) which is math.MinInt32
	ips := []string{"192.168.1.3", "192.168.1.5", "2338615298"}
	for _, ip := range ips {
		if index := StringHashcode(ip); index < 0 {
			t.Fatalf("Bad Index %#v for ip %s", index, ip)
		}
	}
}
