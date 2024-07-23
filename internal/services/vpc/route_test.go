package vpc_test

/*func TestAccVPCRoute_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      vpcchecks.CheckPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "%s"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					vpcchecks.IsPrivateNetworkPresent(
						tt,
						"scaleway_vpc_private_network.pn01",
					),
					resource.TestCheckResourceAttrSet(
						"scaleway_vpc_private_network.pn01",
						"vpc_id"),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.pn01",
						"name",
						privateNetworkName,
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.pn01",
						"region",
						"fr-par",
					),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_vpc_private_network pn01 {
						name = "%s"
						tags = ["tag0", "tag1"]
					}
				`, privateNetworkName),
				Check: resource.ComposeTestCheckFunc(
					vpcchecks.IsPrivateNetworkPresent(
						tt,
						"scaleway_vpc_private_network.pn01",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.pn01",
						"tags.0",
						"tag0",
					),
					resource.TestCheckResourceAttr(
						"scaleway_vpc_private_network.pn01",
						"tags.1",
						"tag1",
					),
				),
			},
		},
	})
}
*/
