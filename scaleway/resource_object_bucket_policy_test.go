package scaleway

/*func TestAccSCWBucketPolicy_basic(t *testing.T) {
	name := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())

	expectedPolicyText := fmt.Sprintf(`{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Sid": "",
        "Effect": "Allow",
        "Principal": {
          "AWS": "*"
        },
        "Action": "s3:*",
        "Resource": [

        ]
      }
    ]
  }`, "fr-par", name, "fr-par", name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:        tt.ProviderFactories,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists("aws_s3_bucket.bucket"),
					testAccCheckBucketHasPolicy("aws_s3_bucket.bucket", expectedPolicyText),
				),
			},
			{
				ResourceName:      "aws_s3_bucket_policy.bucket",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
*/
