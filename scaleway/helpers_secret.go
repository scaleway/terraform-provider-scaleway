package scaleway

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultSecretTimeout = 5 * time.Minute
)

// secretAPIWithRegion returns a new Secret API and the region for a Create request
func secretAPIWithRegion(d *schema.ResourceData, m interface{}) (*secret.API, scw.Region, error) {
	meta := m.(*Meta)
	api := secret.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// secretAPIWithRegionAndDefault returns a new Secret API and the region for a Create request
func secretAPIWithRegionAndDefault(d *schema.ResourceData, m interface{}, defaultRegion scw.Region) (*secret.API, scw.Region, error) {
	meta := m.(*Meta)
	api := secret.NewAPI(meta.scwClient)

	region, err := extractRegionWithDefault(d, meta, defaultRegion)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// secretAPIWithRegionAndProjectID returns a new Secret API, with region and projectID
func secretAPIWithRegionAndProjectID(d *schema.ResourceData, m interface{}) (*secret.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := secret.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", "", err
	}

	projectID, _, err := extractProjectID(d, meta)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, projectID, nil
}

// secretAPIWithRegionAndID returns a Secret API with locality and ID extracted from the state
func secretAPIWithRegionAndID(m interface{}, id string) (*secret.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := secret.NewAPI(meta.scwClient)

	region, id, err := parseRegionalID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

// secretVersionAPIWithRegionAndID returns a Secret API with locality and Nested ID extracted from the state
func secretVersionAPIWithRegionAndID(m interface{}, id string) (*secret.API, scw.Region, string, string, error) {
	meta := m.(*Meta)

	region, id, revision, err := parseLocalizedNestedID(id)
	if err != nil {
		return nil, "", "", "", err
	}

	api := secret.NewAPI(meta.scwClient)
	return api, scw.Region(region), id, revision, nil
}

func isBase64Encoded(data []byte) bool {
	_, err := base64.StdEncoding.DecodeString(string(data))
	return err == nil
}

func base64Encoded(data []byte) string {
	if isBase64Encoded(data) {
		return string(data)
	}
	return base64.StdEncoding.EncodeToString(data)
}

func getSecretFolderByID(ctx context.Context, api *secret.API, region scw.Region, id string) (*secret.Folder, error) {
	listResp, err := api.ListFolders(&secret.ListFoldersRequest{
		Region: region,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	for _, folder := range listResp.Folders {
		if folder.ID == id {
			return folder, nil
		}
	}

	return nil, fmt.Errorf("failed to find folder with given id %q in region %q", id, region)
}
