package scaleway

/*

// getInstanceAPIWithZone returns a new instance API and the zone for a Create request
func getS3ClientWithRegion(d *schema.ResourceData, m interface{}) (*instance.API, utils.Region, error) {
	meta := m.(*Meta)

	region, err := getRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	defaultS3Region, err := utils.ParseRegion(*meta.s3Client.Config.Region)
	if err != nil {
		return nil, "", err
	}

	if region != defaultS3Region {
		meta.getS3ClientWithRegion(region)
	}

	s3 := meta.s3Client.Config.Region

	return instanceApi, region, err
}

*/
