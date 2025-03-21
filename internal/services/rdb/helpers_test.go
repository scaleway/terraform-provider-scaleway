package rdb_test

import (
	"reflect"
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb"
)

func TestPrivilegeV1SchemaUpgradeFunc(t *testing.T) {
	v0Schema := map[string]interface{}{
		"id":            "fr-par/11111111-1111-1111-1111-111111111111",
		"region":        "fr-par",
		"database_name": "database",
		"user_name":     "username",
	}
	v1Schema := map[string]interface{}{
		"id":            "fr-par/11111111-1111-1111-1111-111111111111/database/username",
		"region":        "fr-par",
		"database_name": "database",
		"user_name":     "username",
	}

	actual, err := rdb.PrivilegeV1SchemaUpgradeFunc(t.Context(), v0Schema, nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(v1Schema, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", v1Schema, actual)
	}
}
