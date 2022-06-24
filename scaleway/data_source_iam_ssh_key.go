package scaleway

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"golang.org/x/net/context"
)

func dataSourceScalewayIamSSHKey() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayIamSSKKey().Schema)
	addOptionalFieldsToSchema(dsSchema, "name")

	dsSchema["name"].ConflictsWith = []string{"ssh_key_id"}
	dsSchema["ssh_key_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the SSH key",
		ValidateFunc: validationUUID(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayIamSSHKeyRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayIamSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	iamAPI := iamAPI(meta)

	var sshKey *iam.SSHKey
	sshKeyID, ok := d.GetOk("ssh_key_id")
	if ok {
		res, err := iamAPI.GetSSHKey(&iam.GetSSHKeyRequest{SSHKeyID: expandID(sshKeyID)}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		sshKey = res
	} else {
		res, err := iamAPI.ListSSHKeys(&iam.ListSSHKeysRequest{
			Name:      expandStringPtr(d.Get("name")),
			ProjectID: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.SSHKeys) == 0 {
			return diag.FromErr(fmt.Errorf("no SSH Key found with the name %s", d.Get("name")))
		}
		if len(res.SSHKeys) > 1 {
			return diag.FromErr(fmt.Errorf("%d SSH Keys found with the same name %s", len(res.SSHKeys), d.Get("name")))
		}
		sshKey = res.SSHKeys[0]
	}

	d.SetId(sshKey.ID)
	_ = d.Set("ssh_key_id", sshKey.ID)

	return resourceScalewayIamSSHKeyRead(ctx, d, meta)
}
