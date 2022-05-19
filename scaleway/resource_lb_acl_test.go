package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestAccScalewayLbAcl_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbFrontendDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip ip01 {}
					resource scaleway_lb lb01 {
						ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb-acl"
						type = "lb-s"
					}
					resource scaleway_lb_backend bkd01 {
						lb_id = scaleway_lb.lb01.id
						forward_protocol = "http"
						forward_port = 80
						proxy_protocol = "none"
					}
					resource scaleway_lb_frontend frt01 {
						lb_id = scaleway_lb.lb01.id
						backend_id = scaleway_lb_backend.bkd01.id
						name = "tf-test"
						inbound_port = 80
						timeout_client = "30s"
						acl {
							name = "test-acl"
							action {
								type = "allow"
							}
							match {
								ip_subnet = ["192.168.0.1", "192.168.0.2", "192.168.10.0/24"]
								http_filter = "acl_http_filter_none"
								http_filter_value = []
								invert = "true"
							}
						}
						acl {
							action {
								type = "allow"
							}
							match {
								ip_subnet = ["0.0.0.0/0"]
								http_filter = "path_begin"
								http_filter_value = ["criteria1","criteria2"]
								invert = "true"
							}
						}
						acl {
							action {
								type = "allow"
							}
							match {
								ip_subnet = ["0.0.0.0/0"]
								http_filter = "path_begin"
								http_filter_value = ["criteria1","criteria2"]
							}
						}
						acl {
							action {
								type = "allow"
							}
							match {
								ip_subnet = ["0.0.0.0/0"]
								http_filter = "acl_http_filter_none"
								http_filter_value = []
							}
						}
						acl {
							match {
								http_filter_value = []
								ip_subnet = ["0.0.0.0/0"]
							}
							action {
								type = "deny"
							}
						}

						acl {
							match {
								ip_subnet = ["0.0.0.0/0"]
								http_filter = "http_header_match"
								http_filter_value = ["example.com"]
								http_filter_option = "host"
							}

							action {
								type = "allow"
							}
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayACLAreCorrect(tt, "scaleway_lb_frontend.frt01", []*lbSDK.ACL{
						{
							Name: "test-acl",
							Match: &lbSDK.ACLMatch{
								IPSubnet:        scw.StringSlicePtr([]string{"192.168.0.1", "192.168.0.2", "192.168.10.0/24"}),
								HTTPFilter:      lbSDK.ACLHTTPFilterACLHTTPFilterNone,
								HTTPFilterValue: []*string{},
								Invert:          true,
							},
							Action: &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeAllow},
						},
						{
							Match: &lbSDK.ACLMatch{
								IPSubnet:        scw.StringSlicePtr([]string{"0.0.0.0/0"}),
								HTTPFilter:      lbSDK.ACLHTTPFilterPathBegin,
								HTTPFilterValue: scw.StringSlicePtr([]string{"criteria1", "criteria2"}),
								Invert:          true,
							},
							Action: &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeAllow},
						},
						{
							Match: &lbSDK.ACLMatch{
								IPSubnet:        scw.StringSlicePtr([]string{"0.0.0.0/0"}),
								HTTPFilter:      lbSDK.ACLHTTPFilterPathBegin,
								HTTPFilterValue: scw.StringSlicePtr([]string{"criteria1", "criteria2"}),
								Invert:          false,
							},
							Action: &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeAllow},
						},
						{
							Match: &lbSDK.ACLMatch{
								IPSubnet:        scw.StringSlicePtr([]string{"0.0.0.0/0"}),
								HTTPFilter:      lbSDK.ACLHTTPFilterACLHTTPFilterNone,
								HTTPFilterValue: []*string{},
								Invert:          false,
							},
							Action: &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeAllow},
						},
						{
							Match: &lbSDK.ACLMatch{
								IPSubnet:        scw.StringSlicePtr([]string{"0.0.0.0/0"}),
								HTTPFilter:      lbSDK.ACLHTTPFilterACLHTTPFilterNone,
								HTTPFilterValue: []*string{},
							},
							Action: &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeDeny},
						},
						{
							Match: &lbSDK.ACLMatch{
								IPSubnet:         scw.StringSlicePtr([]string{"0.0.0.0/0"}),
								HTTPFilter:       lbSDK.ACLHTTPFilterHTTPHeaderMatch,
								HTTPFilterValue:  scw.StringSlicePtr([]string{"example.com"}),
								HTTPFilterOption: scw.StringPtr("host"),
							},
							Action: &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeAllow},
						},
					}),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {}
					resource scaleway_lb lb01 {
						ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb-acl"
						type = "lb-s"
					}
					resource scaleway_lb_backend bkd01 {
						lb_id = scaleway_lb.lb01.id
						forward_protocol = "http"
						forward_port = 80
						proxy_protocol = "none"
					}
					resource scaleway_lb_frontend frt01 {
						lb_id = scaleway_lb.lb01.id
						backend_id = scaleway_lb_backend.bkd01.id
						name = "tf-test"
						inbound_port = 80
						timeout_client = "30s"
						acl {
							action {
								type = "allow"
							}
							match {
								ip_subnet = ["10.0.0.10"]
								http_filter = "path_begin"
								http_filter_value = ["foo","bar"]
							}
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayACLAreCorrect(tt, "scaleway_lb_frontend.frt01", []*lbSDK.ACL{
						{
							Match: &lbSDK.ACLMatch{
								IPSubnet:        scw.StringSlicePtr([]string{"10.0.0.10"}),
								HTTPFilter:      lbSDK.ACLHTTPFilterPathBegin,
								HTTPFilterValue: scw.StringSlicePtr([]string{"foo", "bar"}),
								Invert:          false,
							},
							Action: &lbSDK.ACLAction{Type: lbSDK.ACLActionTypeAllow},
						},
					}),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip ip01 {}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_lb_ip.ip01", "id"),
				),
			},
		},
	})
}

func testAccCheckScalewayACLAreCorrect(tt *TestTools, frontendName string, expectedAcls []*lbSDK.ACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// define a wrapper for acl comparison
		testCompareAcls := func(testAcl, apiAcl lbSDK.ACL) bool {
			// drop some values which are not part of the testing acl structure
			apiAcl.ID = ""
			apiAcl.Frontend = nil
			// if we do not pass any name, then drop it from comparison
			if testAcl.Name == "" {
				testAcl.Name = apiAcl.Name
			}
			return aclEquals(&testAcl, &apiAcl)
		}

		rs, ok := s.RootModule().Resources[frontendName]
		if !ok {
			return fmt.Errorf("resource not found: %s", frontendName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id is not set")
		}

		lbAPI, zone, ID, err := lbAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		// fetch our acls from the scaleway
		resACL, err := lbAPI.ListACLs(&lbSDK.ZonedAPIListACLsRequest{
			Zone:       zone,
			FrontendID: ID,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error on getting acl list [%w]", err)
		}

		// verify that the count of api acl is the same as we are expecting it to be
		if len(expectedAcls) != len(resACL.ACLs) {
			return fmt.Errorf("acl count is wrong")
		}
		// convert them to map indexed by the acl index
		aclMap := make(map[int32]*lbSDK.ACL)
		for _, acl := range resACL.ACLs {
			aclMap[acl.Index] = acl
		}

		// check that every index is set up correctly
		for i := 1; i <= len(expectedAcls); i++ {
			if _, found := aclMap[int32(i)]; !found {
				return fmt.Errorf("cannot find an index set [%d]", i)
			}
			if !testCompareAcls(*expectedAcls[i-1], *aclMap[int32(i)]) {
				return fmt.Errorf("two acls are not equal on stage %d", i)
			}
		}
		// check the actual data

		return nil
	}
}
