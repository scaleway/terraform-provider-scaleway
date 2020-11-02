package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/assert"
)

func TestAccScalewayLbFrontend_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbFrontendBetaDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip_beta ip01 {}
					resource scaleway_lb_beta lb01 {
						ip_id = scaleway_lb_ip_beta.ip01.id
						name = "test-lb"
						type = "lb-s"
					}
					resource scaleway_lb_backend_beta bkd01 {
						lb_id = scaleway_lb_beta.lb01.id
						forward_protocol = "tcp"
						forward_port = 80
						proxy_protocol = "none"
					}
					resource scaleway_lb_frontend_beta frt01 {
						lb_id = scaleway_lb_beta.lb01.id
						backend_id = scaleway_lb_backend_beta.bkd01.id
						inbound_port = 80
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbFrontendBetaExists(tt, "scaleway_lb_frontend_beta.frt01"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend_beta.frt01", "inbound_port", "80"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend_beta.frt01", "timeout_client", ""),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip_beta ip01 {}
					resource scaleway_lb_beta lb01 {
						ip_id = scaleway_lb_ip_beta.ip01.id
						name = "test-lb"
						type = "lb-s"
					}
					resource scaleway_lb_backend_beta bkd01 {
						lb_id = scaleway_lb_beta.lb01.id
						forward_protocol = "tcp"
						forward_port = 80
						proxy_protocol = "none"
					}
					resource scaleway_lb_frontend_beta frt01 {
						lb_id = scaleway_lb_beta.lb01.id
						backend_id = scaleway_lb_backend_beta.bkd01.id
						name = "tf-test"
						inbound_port = 443
						timeout_client = "30s"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbFrontendBetaExists(tt, "scaleway_lb_frontend_beta.frt01"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend_beta.frt01", "name", "tf-test"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend_beta.frt01", "inbound_port", "443"),
					resource.TestCheckResourceAttr("scaleway_lb_frontend_beta.frt01", "timeout_client", "30s"),
				),
			},
		},
	})
}

func TestAccScalewayLbAcl_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbFrontendBetaDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_lb_ip_beta ip01 {}
					resource scaleway_lb_beta lb01 {
						ip_id = scaleway_lb_ip_beta.ip01.id
						name = "test-lb-acl"
						type = "lb-s"
					}
					resource scaleway_lb_backend_beta bkd01 {
						lb_id = scaleway_lb_beta.lb01.id
						forward_protocol = "http"
						forward_port = 80
						proxy_protocol = "none"
					}
					resource scaleway_lb_frontend_beta frt01 {
						lb_id = scaleway_lb_beta.lb01.id
						backend_id = scaleway_lb_backend_beta.bkd01.id
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
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayACLAreCorrect(tt, "scaleway_lb_frontend_beta.frt01", []*lb.ACL{
						{
							Name: "test-acl",
							Match: &lb.ACLMatch{
								IPSubnet:        scw.StringSlicePtr([]string{"192.168.0.1", "192.168.0.2", "192.168.10.0/24"}),
								HTTPFilter:      lb.ACLHTTPFilterACLHTTPFilterNone,
								HTTPFilterValue: []*string{},
								Invert:          true,
							},
							Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						},
						{
							Match: &lb.ACLMatch{
								IPSubnet:        scw.StringSlicePtr([]string{"0.0.0.0/0"}),
								HTTPFilter:      lb.ACLHTTPFilterPathBegin,
								HTTPFilterValue: scw.StringSlicePtr([]string{"criteria1", "criteria2"}),
								Invert:          true,
							},
							Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						},
						{
							Match: &lb.ACLMatch{
								IPSubnet:        scw.StringSlicePtr([]string{"0.0.0.0/0"}),
								HTTPFilter:      lb.ACLHTTPFilterPathBegin,
								HTTPFilterValue: scw.StringSlicePtr([]string{"criteria1", "criteria2"}),
								Invert:          false,
							},
							Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						},
						{
							Match: &lb.ACLMatch{
								IPSubnet:        scw.StringSlicePtr([]string{"0.0.0.0/0"}),
								HTTPFilter:      lb.ACLHTTPFilterACLHTTPFilterNone,
								HTTPFilterValue: []*string{},
								Invert:          false,
							},
							Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						},
						{
							Match: &lb.ACLMatch{
								IPSubnet:        scw.StringSlicePtr([]string{"0.0.0.0/0"}),
								HTTPFilter:      lb.ACLHTTPFilterACLHTTPFilterNone,
								HTTPFilterValue: []*string{},
							},
							Action: &lb.ACLAction{Type: lb.ACLActionTypeDeny},
						},
					}),
				),
			},
			{
				Config: `
					resource scaleway_lb_ip_beta ip01 {}
					resource scaleway_lb_beta lb01 {
						ip_id = scaleway_lb_ip_beta.ip01.id
						name = "test-lb-acl"
						type = "lb-s"
					}
					resource scaleway_lb_backend_beta bkd01 {
						lb_id = scaleway_lb_beta.lb01.id
						forward_protocol = "http"
						forward_port = 80
						proxy_protocol = "none"
					}
					resource scaleway_lb_frontend_beta frt01 {
						lb_id = scaleway_lb_beta.lb01.id
						backend_id = scaleway_lb_backend_beta.bkd01.id
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
					testAccCheckScalewayACLAreCorrect(tt, "scaleway_lb_frontend_beta.frt01", []*lb.ACL{
						{
							Match: &lb.ACLMatch{
								IPSubnet:        scw.StringSlicePtr([]string{"10.0.0.10"}),
								HTTPFilter:      lb.ACLHTTPFilterPathBegin,
								HTTPFilterValue: scw.StringSlicePtr([]string{"foo", "bar"}),
								Invert:          false,
							},
							Action: &lb.ACLAction{Type: lb.ACLActionTypeAllow},
						},
					}),
				),
			},
		},
	})
}

func testAccCheckScalewayACLAreCorrect(tt *TestTools, frontendName string, expectedAcls []*lb.ACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		//define a wrapper for acl comparison
		testCompareAcls := func(testAcl, apiAcl lb.ACL) bool {
			//drop some values which are not part of the testing acl structure
			apiAcl.ID = ""
			apiAcl.Frontend = nil
			//if we do not pass any name, then drop it from comparison
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

		lbAPI, region, ID, err := lbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		//fetch our acls from the scaleway
		resACL, err := lbAPI.ListACLs(&lb.ListACLsRequest{
			Region:     region,
			FrontendID: ID,
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error on getting acl list [%s]", err)
		}

		//verify that the count of api acl is the same as we are expecting it to be
		if len(expectedAcls) != len(resACL.ACLs) {
			return fmt.Errorf("acl count is wrong")
		}
		//convert them to map indexed by the acl index
		aclMap := make(map[int32]*lb.ACL)
		for _, acl := range resACL.ACLs {
			aclMap[acl.Index] = acl
		}

		//check that every index is set up correctly
		for i := 1; i <= len(expectedAcls); i++ {
			if _, found := aclMap[int32(i)]; !found {
				return fmt.Errorf("cannot find an index set [%d]", i)
			}
			if !testCompareAcls(*expectedAcls[i-1], *aclMap[int32(i)]) {
				return fmt.Errorf("two acls are not equal on stage %d", i)
			}
		}
		//check the actual data

		return nil
	}
}
func testAccCheckScalewayLbFrontendBetaExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		lbAPI, region, ID, err := lbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = lbAPI.GetFrontend(&lb.GetFrontendRequest{
			FrontendID: ID,
			Region:     region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayLbFrontendBetaDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_lb_frontend_beta" {
				continue
			}

			lbAPI, region, ID, err := lbAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = lbAPI.GetFrontend(&lb.GetFrontendRequest{
				Region:     region,
				FrontendID: ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("LB Frontend (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}

func TestAclEqual(t *testing.T) {
	aclA := &lb.ACL{
		Name: "test-acl",
		Match: &lb.ACLMatch{
			IPSubnet:        scw.StringSlicePtr([]string{"192.168.0.1", "192.168.0.2", "192.168.10.0/24"}),
			HTTPFilter:      lb.ACLHTTPFilterACLHTTPFilterNone,
			HTTPFilterValue: nil,
			Invert:          true,
		},
		Action:   &lb.ACLAction{Type: lb.ACLActionTypeAllow},
		Frontend: nil,
		Index:    1,
	}
	aclB := &lb.ACL{
		Name: "test-acl",
		Match: &lb.ACLMatch{
			IPSubnet:        scw.StringSlicePtr([]string{"192.168.0.1", "192.168.0.2", "192.168.10.0/24"}),
			HTTPFilter:      lb.ACLHTTPFilterACLHTTPFilterNone,
			HTTPFilterValue: nil,
			Invert:          true,
		},
		Action:   &lb.ACLAction{Type: lb.ACLActionTypeAllow},
		Frontend: nil,
		Index:    1,
	}
	assert.True(t, aclEquals(aclA, aclB))

	//change name
	aclA.Name = "nope"
	assert.False(t, aclEquals(aclA, aclB))
	aclA.Name = aclB.Name

	//check action
	aclA.Action = nil
	assert.False(t, aclEquals(aclA, aclB))
	aclA.Action = &lb.ACLAction{Type: lb.ACLActionTypeAllow}
	assert.True(t, aclEquals(aclA, aclB))
	aclA.Action = &lb.ACLAction{Type: lb.ACLActionTypeDeny}
	assert.False(t, aclEquals(aclA, aclB))
	aclA.Action = &lb.ACLAction{Type: lb.ACLActionTypeAllow}
	assert.True(t, aclEquals(aclA, aclB))

	//check match
	aclA.Match.IPSubnet = scw.StringSlicePtr([]string{"192.168.0.1", "192.168.0.2", "192.168.10.0/24", "0.0.0.0"})
	assert.False(t, aclEquals(aclA, aclB))
}
