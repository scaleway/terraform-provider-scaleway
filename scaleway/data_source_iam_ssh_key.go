package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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

	sshKeyID, sshKeyIDExists := d.GetOk("ssh_key_id")
	if !sshKeyIDExists {
		sshKeyName := d.Get("name").(string)
		res, err := iamAPI.ListSSHKeys(&iam.ListSSHKeysRequest{
			Name:      expandStringPtr(sshKeyName),
			ProjectID: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundKey, err := findExact(
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

	diags := resourceScalewayIamSSHKeyRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read iam ssh key state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("iam ssh key (%s) not found", sshKeyID)
	}

	return nil
}
