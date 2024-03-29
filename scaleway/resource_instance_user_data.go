package scaleway

import (
	"bytes"
	"context"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceScalewayInstanceUserData() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayInstanceUserDataCreate,
		ReadContext:   resourceScalewayInstanceUserDataRead,
		UpdateContext: resourceScalewayInstanceUserDataUpdate,
		DeleteContext: resourceScalewayInstanceUserDataDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(DefaultInstanceServerWaitTimeout),
			Read:    schema.DefaultTimeout(DefaultInstanceServerWaitTimeout),
			Update:  schema.DefaultTimeout(DefaultInstanceServerWaitTimeout),
			Delete:  schema.DefaultTimeout(DefaultInstanceServerWaitTimeout),
			Default: schema.DefaultTimeout(DefaultInstanceServerWaitTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The ID of the server",
				ValidateFunc: verify.IsUUIDWithLocality(),
			},
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The key of the user data to set.",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The value of the user data to set.",
			},
			"zone": zonal.Schema(),
		},
		CustomizeDiff: CustomizeDiffLocalityCheck("server_id"),
	}
}

func resourceScalewayInstanceUserDataCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	serverID := locality.ExpandID(d.Get("server_id").(string))
	server, err := waitForInstanceServer(ctx, instanceAPI, zone, serverID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	key := d.Get("key").(string)
	value := bytes.NewBufferString(d.Get("value").(string))

	userDataRequest := &instance.SetServerUserDataRequest{
		Zone:     zone,
		ServerID: server.ID,
		Key:      key,
		Content:  value,
	}

	if v, ok := d.GetOk("zone"); ok {
		userDataRequest.Zone = scw.Zone(v.(string))
	}

	err = instanceAPI.SetServerUserData(userDataRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewNestedIDString(zone, key, server.ID))

	return resourceScalewayInstanceUserDataRead(ctx, d, m)
}

func resourceScalewayInstanceUserDataRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, id, key, err := InstanceAPIWithZoneAndNestedID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	server, err := waitForInstanceServer(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	requestGetUserData := &instance.GetServerUserDataRequest{
		Zone:     zone,
		ServerID: server.ID,
		Key:      key,
	}

	if v, ok := d.GetOk("zone"); ok {
		requestGetUserData.Zone = scw.Zone(v.(string))
		zone = requestGetUserData.Zone
	}

	serverUserDataRawValue, err := instanceAPI.GetServerUserData(requestGetUserData, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	userDataValue, err := io.ReadAll(serverUserDataRawValue)
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("server_id", zonal.NewID(zone, server.ID).String())
	_ = d.Set("key", key)
	_ = d.Set("value", string(userDataValue))
	_ = d.Set("zone", zone.String())

	return nil
}

func resourceScalewayInstanceUserDataUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, id, key, err := InstanceAPIWithZoneAndNestedID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	server, err := waitForInstanceServer(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	userDataRequest := &instance.SetServerUserDataRequest{
		Zone:     zone,
		ServerID: server.ID,
		Key:      key,
	}

	if v, ok := d.GetOk("zone"); ok {
		userDataRequest.Zone = scw.Zone(v.(string))
	}

	if d.HasChanges("value") {
		value := d.Get("value")
		userDataRequest.Content = bytes.NewBufferString(value.(string))
	}

	err = instanceAPI.SetServerUserData(userDataRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForInstanceServer(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayInstanceUserDataRead(ctx, d, m)
}

func resourceScalewayInstanceUserDataDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, id, key, err := InstanceAPIWithZoneAndNestedID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deleteUserData := &instance.DeleteServerUserDataRequest{
		ServerID: locality.ExpandID(id),
		Key:      key,
		Zone:     zone,
	}

	if v, ok := d.GetOk("zone"); ok {
		deleteUserData.Zone = scw.Zone(v.(string))
	}

	err = instanceAPI.DeleteServerUserData(deleteUserData, scw.WithContext(ctx))

	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
