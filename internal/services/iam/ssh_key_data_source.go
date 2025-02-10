package iam

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceSSHKey() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceSSKKey().Schema)
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "project_id")

	dsSchema["name"].ConflictsWith = []string{"ssh_key_id"}
	dsSchema["ssh_key_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the SSH key",
		ValidateDiagFunc: verify.IsUUID(),
	}

	return &schema.Resource{
		ReadContext: DataSourceIamSSHKeyRead,
		Schema:      dsSchema,
	}
}

func DataSourceIamSSHKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	iamAPI := NewAPI(m)

	sshKeyID, sshKeyIDExists := d.GetOk("ssh_key_id")
	if !sshKeyIDExists {
		sshKeyName := d.Get("name").(string)

		res, err := iamAPI.ListSSHKeys(&iam.ListSSHKeysRequest{
			Name:      types.ExpandStringPtr(sshKeyName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundKey, err := datasource.FindExact(
			res.SSHKeys,
			func(s *iam.SSHKey) bool { return s.Name == sshKeyName },
			sshKeyName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		sshKeyID = foundKey.ID
	}

	d.SetId(sshKeyID.(string))

	err := d.Set("ssh_key_id", sshKeyID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceIamSSHKeyRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read iam ssh key state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("iam ssh key (%s) not found", sshKeyID)
	}

	return nil
}
