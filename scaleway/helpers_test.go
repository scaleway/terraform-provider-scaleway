package scaleway_test

import (
	"fmt"
	"net"
	"regexp"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
	"github.com/stretchr/testify/assert"
)

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
	expected := scaleway.StringHashcode(v)
	for i := 0; i < 100; i++ {
		actual := scaleway.StringHashcode(v)
		if actual != expected {
			t.Fatalf("bad: %#v\n\t%#v", actual, expected)
		}
	}
}

func TestStringHashcode_positiveIndex(t *testing.T) {
	// "2338615298" hashes to uint32(2147483648) which is math.MinInt32
	ips := []string{"192.168.1.3", "192.168.1.5", "2338615298"}
	for _, ip := range ips {
		if index := scaleway.StringHashcode(ip); index < 0 {
			t.Fatalf("Bad Index %#v for ip %s", index, ip)
		}
	}
}

func TestAcc_GetRawConfigForKey(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestEngineVersion := testAccCheckScalewayRdbEngineGetLatestVersion(tt, postgreSQLEngineName)
	instanceUnchangedConfig := fmt.Sprintf(`
						name = "test-get-raw-config-for-key"
						engine = %q
						is_ha_cluster = false
						user_name = "my_initial_user"
						password = "thiZ_is_v&ry_s3cret"`, latestEngineVersion)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayRdbInstanceDestroy(tt),
			testAccCheckScalewayVPCPrivateNetworkDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_vpc_private_network pn {}

					resource scaleway_rdb_instance main {%s
						node_type = "db-dev-s"
						disable_backup = false
						volume_type = "lssd"
						tags = [ "terraform-test", "core", "get-raw-config" ]
						private_network {
							pn_id = "${scaleway_vpc_private_network.pn.id}"
							enable_ipam = true
						}
						settings = {
							work_mem = "4"
							max_connections = "200"
						}
					}
				`, instanceUnchangedConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRdbExists(tt, "scaleway_rdb_instance.main"),
					assertGetRawConfigResults(t, "is_ha_cluster", false, true, cty.Bool),
					assertGetRawConfigResults(t, "disable_backup", false, true, cty.Bool),
					assertGetRawConfigResults(t, "volume_type", "lssd", true, cty.String),
					assertGetRawConfigResults(t, "volume_size_in_gb", nil, false, cty.Number),
					assertGetRawConfigResults(t, "tags.0", "terraform-test", true, cty.String),
					assertGetRawConfigResults(t, "tags.1", "core", true, cty.String),
					assertGetRawConfigResults(t, "tags.2", "get-raw-config", true, cty.String),
					assertGetRawConfigResults(t, "tags.3", nil, false, cty.String),
					assertGetRawConfigResults(t, "private_network.0.ip_net", nil, false, cty.String),
					assertGetRawConfigResults(t, "private_network.0.enable_ipam", true, true, cty.Bool),
					assertGetRawConfigResults(t, "private_network.#.enable_ipam", true, true, cty.Bool),
					assertGetRawConfigResults(t, "settings.work_mem", "4", true, cty.String),
					assertGetRawConfigResults(t, "settings.max_connections", "200", true, cty.String),
					assertGetRawConfigResults(t, "settings.not_in_map", nil, false, cty.String),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_vpc_private_network pn {}

					resource scaleway_rdb_instance main {%s
						node_type = "db-dev-s"
						disable_backup = true
						volume_type = "bssd"
						volume_size_in_gb = 10
						tags = [ "terraform-test", "core", "get-raw-config-for-key" ]
						private_network {
							pn_id = "${scaleway_vpc_private_network.pn.id}"
							ip_net = "172.16.32.1/24"
						}
						settings = {
							work_mem = "2"
							max_connections = "100"
						}
					}
				`, instanceUnchangedConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRdbExists(tt, "scaleway_rdb_instance.main"),
					assertGetRawConfigResults(t, "is_ha_cluster", false, true, cty.Bool),
					assertGetRawConfigResults(t, "disable_backup", true, true, cty.Bool),
					assertGetRawConfigResults(t, "volume_type", "bssd", true, cty.String),
					assertGetRawConfigResults(t, "volume_size_in_gb", 10, true, cty.Number),
					assertGetRawConfigResults(t, "tags.0", "terraform-test", true, cty.String),
					assertGetRawConfigResults(t, "tags.1", "core", true, cty.String),
					assertGetRawConfigResults(t, "tags.2", "get-raw-config-for-key", true, cty.String),
					assertGetRawConfigResults(t, "tags.3", nil, false, cty.String),
					assertGetRawConfigResults(t, "private_network.0.ip_net", "172.16.32.1/24", true, cty.String),
					assertGetRawConfigResults(t, "private_network.0.enable_ipam", false, false, cty.Bool),
					assertGetRawConfigResults(t, "private_network.#.enable_ipam", false, false, cty.Bool),
					assertGetRawConfigResults(t, "private_network.#.not_in_list", false, false, cty.Bool),
					assertGetRawConfigResults(t, "settings.work_mem", "2", true, cty.String),
					assertGetRawConfigResults(t, "settings.max_connections", "100", true, cty.String),
					assertGetRawConfigResults(t, "settings.not_in_map", nil, false, cty.String),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_vpc_private_network pn {}
			
					resource scaleway_rdb_instance main {%s
						node_type = "db-gp-s"
						tags = [ "terraform-test", "core", "get-raw-config" ]
						private_network {
							pn_id = "${scaleway_vpc_private_network.pn.id}"
							enable_ipam = true
						}
						settings = {
							work_mem = "4"
							max_connections = "200"
						}
					}
				`, instanceUnchangedConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayRdbExists(tt, "scaleway_rdb_instance.main"),
					assertGetRawConfigResults(t, "is_ha_cluster", false, true, cty.Bool),
					assertGetRawConfigResults(t, "disable_backup", false, false, cty.Bool),
					assertGetRawConfigResults(t, "volume_type", "lssd", true, cty.String),
					assertGetRawConfigResults(t, "volume_size_in_gb", nil, false, cty.Number),
					assertGetRawConfigResults(t, "tags.0", "terraform-test", true, cty.String),
					assertGetRawConfigResults(t, "tags.1", "core", true, cty.String),
					assertGetRawConfigResults(t, "tags.2", "get-raw-config", true, cty.String),
					assertGetRawConfigResults(t, "tags.3", nil, false, cty.String),
					assertGetRawConfigResults(t, "private_network.0.ip_net", nil, false, cty.String),
					assertGetRawConfigResults(t, "private_network.0.enable_ipam", true, true, cty.Bool),
					assertGetRawConfigResults(t, "private_network.#.enable_ipam", true, true, cty.Bool),
					assertGetRawConfigResults(t, "settings.work_mem", "4", true, cty.String),
					assertGetRawConfigResults(t, "settings.max_connections", "200", true, cty.String),
					assertGetRawConfigResults(t, "settings.not_in_map", nil, false, cty.String),
				),
			},
		},
	})
}

func assertGetRawConfigResults(t *testing.T, key string, expectedValue any, expectedSet bool, ty cty.Type) resource.TestCheckFunc {
	t.Helper()
	return func(s *terraform.State) error {
		resourceTFName := "scaleway_rdb_instance.main"
		rs, ok := s.RootModule().Resources[resourceTFName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceTFName)
		}
		if rs.Primary.RawConfig.IsNull() {
			return nil
		}

		actualValue, actualSet := scaleway.GetKeyInRawConfigMap(rs.Primary.RawConfig.AsValueMap(), key, ty)
		assert.Equal(t, expectedSet, actualSet)
		assert.Equal(t, expectedValue, actualValue)

		return nil
	}
}
