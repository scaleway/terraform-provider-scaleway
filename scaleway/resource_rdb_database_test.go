package scaleway

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	mock "github.com/scaleway/terraform-provider-scaleway/scaleway/mocks"
	"github.com/stretchr/testify/assert"
)

func init() {
	resource.AddTestSweepers("scaleway_rdb_database", &resource.Sweeper{
		Name: "scaleway_rdb_database",
		F:    testSweepRDBInstance,
	})
}

func TestAccScalewayRdbDatabase_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	instanceName := "TestAccScalewayRdbDatabase_Basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayRdbInstanceDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_rdb_instance main {
						name = "%s"
						node_type = "db-dev-s"
						engine = "PostgreSQL-12"
						is_ha_cluster = false
						tags = [ "terraform-test", "scaleway_rdb_user", "minimal" ]
					}

					resource scaleway_rdb_database main {
						instance_id = scaleway_rdb_instance.main.id
						name = "foo"
					}`, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdbDatabaseExists(tt, "scaleway_rdb_instance.main", "scaleway_rdb_database.main"),
					resource.TestCheckResourceAttr("scaleway_rdb_database.main", "name", "foo"),
				),
			},
		},
	})
}

func testAccCheckRdbDatabaseExists(tt *TestTools, instance string, database string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instanceResource, ok := state.RootModule().Resources[instance]
		if !ok {
			return fmt.Errorf("resource not found: %s", instance)
		}

		databaseResource, ok := state.RootModule().Resources[database]
		if !ok {
			return fmt.Errorf("resource database not found: %s", database)
		}

		rdbAPI, region, _, err := rdbAPIWithRegionAndID(tt.Meta, instanceResource.Primary.ID)
		if err != nil {
			return err
		}

		region, instanceID, databaseName, err := resourceScalewayRdbDatabaseParseID(databaseResource.Primary.ID)
		if err != nil {
			return err
		}

		databases, err := rdbAPI.ListDatabases(&rdb.ListDatabasesRequest{
			Region:     region,
			InstanceID: instanceID,
			Name:       &databaseName,
			Managed:    nil,
			Owner:      nil,
			OrderBy:    "",
		})
		if err != nil {
			return err
		}

		if len(databases.Databases) != 1 {
			return fmt.Errorf("no database found")
		}

		return nil
	}
}

func TestResourceScalewayRdbDatabaseReadWithoutIDReturnDiagnotics(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)

	data := schema.ResourceData{}
	meta, _ := buildMeta(&MetaConfig{
		terraformVersion: "terraform-test-unit",
	})
	ctx := mock.NewMockContext(ctrl)

	diags := resourceScalewayRdbDatabaseRead(ctx, &data, meta)

	assert.Len(diags, 1)
	assert.Equal("can't parse user resource id: ", diags[0].Summary)
	assert.Equal(diag.Error, diags[0].Severity)
}

func TestResourceScalewayRdbDatabaseReadWithRdbErrorIdReturnDiagnotics(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)

	data := schema.ResourceData{}
	data.SetId("fr-srr/1111-11111111-111111111111/dbname")
	meta, _ := buildMeta(&MetaConfig{
		terraformVersion: "terraform-test-unit",
	})
	rdbApi := mock.NewMockRdbApiInterface(ctrl)
	rdbApi.EXPECT().ListDatabases(gomock.Any(), gomock.Any()).Return(nil, errors.New("Error"))
	meta.mockedApi = rdbApi
	ctx := mock.NewMockContext(ctrl)

	diags := resourceScalewayRdbDatabaseRead(ctx, &data, meta)

	assert.Len(diags, 1)
	assert.Equal(diag.Error, diags[0].Severity)
}

func TestResourceScalewayRdbDatabaseReadSetResourceData(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)

	data := schema.TestResourceDataRaw(t, resourceScalewayRdbDatabase().Schema, make(map[string]interface{}))
	data.SetId("fr-srr/1111-11111111-111111111111/dbname")
	meta, _ := buildMeta(&MetaConfig{
		terraformVersion: "terraform-test-unit",
	})
	rdbApi := mock.NewMockRdbApiInterface(ctrl)
	matcher := ListDatabasesRequestMatcher{
		ExpectedRegion:       "fr-srr",
		ExpectedInstanceID:   "1111-11111111-111111111111",
		ExpectedDatabaseName: "dbname",
	}
	db := rdb.Database{
		Name:    "dbname",
		Owner:   "dbowner",
		Managed: true,
		Size:    42,
	}
	dbs := make([]*rdb.Database, 0)
	dbs = append(dbs, &db)
	resp := rdb.ListDatabasesResponse{
		Databases:  dbs,
		TotalCount: 1,
	}
	rdbApi.EXPECT().ListDatabases(matcher, gomock.Any()).Return(&resp, nil)
	meta.mockedApi = rdbApi
	ctx := mock.NewMockContext(ctrl)

	diags := resourceScalewayRdbDatabaseRead(ctx, data, meta)

	assert.Len(diags, 0)
	assert.Equal("fr-srr/1111-11111111-111111111111", data.Get("instance_id"))
	assert.Equal("dbname", data.Get("name"))
	assert.Equal("dbowner", data.Get("owner"))
	assert.True(data.Get("managed").(bool))
	assert.Equal("42", data.Get("size"))
}

type CreateDatabaseRequestMatcher struct {
	ExpectedRegion       string
	ExpectedInstanceID   string
	ExpectedDatabaseName string
	errorString          string
}

func (m CreateDatabaseRequestMatcher) Matches(x interface{}) bool {
	req := x.(*rdb.CreateDatabaseRequest)

	if req.Region.String() != m.ExpectedRegion {
		return false
	}
	if req.InstanceID != m.ExpectedInstanceID {
		return false
	}
	if req.Name != m.ExpectedDatabaseName {
		return false
	}
	return true
}

func (m CreateDatabaseRequestMatcher) String() string {
	return fmt.Sprintf("is equal to (%s, %s, %s)", m.ExpectedRegion, m.ExpectedInstanceID, m.ExpectedDatabaseName)
}

type ListDatabasesRequestMatcher struct {
	ExpectedRegion       string
	ExpectedInstanceID   string
	ExpectedDatabaseName string
	errorString          string
}

func (m ListDatabasesRequestMatcher) Matches(x interface{}) bool {
	req := x.(*rdb.ListDatabasesRequest)

	if req.Region.String() != m.ExpectedRegion {
		return false
	}
	if req.InstanceID != m.ExpectedInstanceID {
		return false
	}
	if fmt.Sprintf("%s", *req.Name) != m.ExpectedDatabaseName {
		return false
	}
	return true
}

func (m ListDatabasesRequestMatcher) String() string {
	return fmt.Sprintf("is equal to (%s, %s, %s)", m.ExpectedRegion, m.ExpectedInstanceID, m.ExpectedDatabaseName)
}

func TestResourceScalewayRdbDatabaseParseIDWithWronglyFormatedIdReturnError(t *testing.T) {
	assert := assert.New(t)
	_, _, _, err := resourceScalewayRdbDatabaseParseID("notandid")
	assert.Error(err)
	assert.Equal("can't parse user resource id: notandid", err.Error())
}

func TestResourceScalewayRdbDatabaseParseID(t *testing.T) {
	assert := assert.New(t)
	region, instanceID, dbname, err := resourceScalewayRdbDatabaseParseID("region/instanceid/dbname")
	assert.NoError(err)
	assert.Equal(scw.Region("region"), region)
	assert.Equal("instanceid", instanceID)
	assert.Equal("dbname", dbname)
}
func TestResourceScalewayRdbDatabaseCreateWithRdbErrorReturnDiagnotics(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)

	raw := make(map[string]interface{})
	raw["instance_id"] = "fr-srr/1111-11111111-111111111111"
	raw["name"] = "dbname"
	data := schema.TestResourceDataRaw(t, resourceScalewayRdbDatabase().Schema, raw)
	meta, _ := buildMeta(&MetaConfig{
		terraformVersion: "terraform-test-unit",
	})
	rdbApi := mock.NewMockRdbApiInterface(ctrl)

	rdbApi.EXPECT().CreateDatabase(gomock.Any(), gomock.Any()).Return(nil, errors.New("Error"))
	meta.mockedApi = rdbApi
	ctx := mock.NewMockContext(ctrl)

	diags := resourceScalewayRdbDatabaseCreate(ctx, data, meta)

	assert.Len(diags, 1)
	assert.Equal(diag.Error, diags[0].Severity)
}
