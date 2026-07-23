package mailbox

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

// DataSourceMailbox lets users look up a mailbox by its ID or by its email address.
func DataSourceMailbox() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceMailbox().SchemaFunc())

	// All fields are computed from the resource schema; expose these lookup keys.
	datasource.AddOptionalFieldsToSchema(dsSchema, "email")

	dsSchema["mailbox_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "UUID of the mailbox. Conflicts with email.",
		ValidateDiagFunc: verify.IsUUID(),
		ConflictsWith:    []string{"email"},
	}
	dsSchema["email"].ConflictsWith = []string{"mailbox_id"}

	return &schema.Resource{
		ReadContext: dataSourceMailboxRead,
		Schema:      dsSchema,
	}
}

func dataSourceMailboxRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newMailboxAPI(m)

	mailboxID, hasID := d.GetOk("mailbox_id")

	if hasID {
		d.SetId(mailboxID.(string))
		return readMailboxIntoState(ctx, d, m)
	}

	// Look up by email: list all mailboxes and filter.
	email, hasEmail := d.GetOk("email")
	if !hasEmail {
		return diag.Errorf("one of mailbox_id or email must be provided")
	}

	var foundID string
	page := int32(1)

	for {
		resp, err := api.ListMailboxes(&mailboxsdk.ListMailboxesRequest{
			Page: &page,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		for _, mb := range resp.Mailboxes {
			if mb.Email == email.(string) {
				if foundID != "" {
					return diag.Errorf("found multiple mailboxes with email %q", email)
				}
				foundID = mb.ID
			}
		}

		if uint64(int(page)*50) >= resp.TotalCount {
			break
		}
		page++
	}

	if foundID == "" {
		return diag.FromErr(fmt.Errorf("no mailbox found with email %q", email))
	}

	d.SetId(foundID)

	return readMailboxIntoState(ctx, d, m)
}
