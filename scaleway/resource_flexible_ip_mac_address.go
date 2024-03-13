package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	flexibleip "github.com/scaleway/scaleway-sdk-go/api/flexibleip/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func resourceScalewayFlexibleIPMACAddress() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayFlexibleIPMACCreate,
		ReadContext:   resourceScalewayFlexibleIPMACRead,
		UpdateContext: resourceScalewayFlexibleIPMACUpdate,
		DeleteContext: resourceScalewayFlexibleIPMACDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultFlexibleIPTimeout),
			Read:    schema.DefaultTimeout(defaultFlexibleIPTimeout),
			Update:  schema.DefaultTimeout(defaultFlexibleIPTimeout),
			Delete:  schema.DefaultTimeout(defaultFlexibleIPTimeout),
			Default: schema.DefaultTimeout(defaultFlexibleIPTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"flexible_ip_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The ID of the flexible IP for which to generate a virtual MAC",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The type of the virtual MAC",
				ValidateFunc: validation.StringInSlice([]string{
					flexibleip.MACAddressTypeVmware.String(),
					flexibleip.MACAddressTypeXen.String(),
					flexibleip.MACAddressTypeKvm.String(),
				}, false),
			},
			"flexible_ip_ids_to_duplicate": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Description: `The IDs of the flexible IPs on which to duplicate the virtual MAC

**NOTE** : The flexible IPs need to be attached to the same server for the operation to work.`,
			},
			// computed
			"address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Virtual MAC address",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Virtual MAC status",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the virtual MAC (Format ISO 8601)",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the virtual MAC (Format ISO 8601)",
			},
			"zone": zonal.Schema(),
		},
		CustomizeDiff: CustomizeDiffLocalityCheck("flexible_ip_id"),
	}
}

func resourceScalewayFlexibleIPMACCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fipAPI, zone, err := fipAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	fipID := locality.ExpandID(d.Get("flexible_ip_id"))
	_, err = waitFlexibleIP(ctx, fipAPI, zone, fipID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := fipAPI.GenerateMACAddr(&flexibleip.GenerateMACAddrRequest{
		Zone:    zone,
		FipID:   fipID,
		MacType: flexibleip.MACAddressType(d.Get("type").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if res.MacAddress != nil {
		d.SetId(zonal.NewIDString(zone, res.MacAddress.ID))
	}

	fip, err := waitFlexibleIP(ctx, fipAPI, zone, res.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	duplicateIDs, duplicateIDsExist := d.GetOk("flexible_ip_ids_to_duplicate")
	if duplicateIDsExist {
		dupIDs := expandStrings(duplicateIDs.(*schema.Set).List())
		for _, dupID := range dupIDs {
			_, err := fipAPI.DuplicateMACAddr(&flexibleip.DuplicateMACAddrRequest{
				Zone:               zone,
				FipID:              locality.ExpandID(dupID),
				DuplicateFromFipID: fip.ID,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
			_, err = waitFlexibleIP(ctx, fipAPI, zone, locality.ExpandID(dupID), d.Timeout(schema.TimeoutCreate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceScalewayFlexibleIPMACRead(ctx, d, m)
}

func resourceScalewayFlexibleIPMACRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fipAPI, zone, err := fipAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	fip, err := fipAPI.GetFlexibleIP(&flexibleip.GetFlexibleIPRequest{
		Zone:  zone,
		FipID: locality.ExpandID(d.Get("flexible_ip_id")),
	}, scw.WithContext(ctx))
	if err != nil {
		// We check for 403 because flexible API returns 403 for a deleted IP
		if httperrors.Is404(err) || httperrors.Is403(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("flexible_ip_id", zonal.NewIDString(zone, fip.ID))
	if fip.MacAddress != nil {
		_ = d.Set("type", fip.MacAddress.MacType.String())
		_ = d.Set("address", fip.MacAddress.MacAddress)
		_ = d.Set("status", fip.MacAddress.Status.String())
		_ = d.Set("created_at", flattenTime(fip.MacAddress.CreatedAt))
		_ = d.Set("updated_at", flattenTime(fip.MacAddress.UpdatedAt))
		_ = d.Set("zone", fip.MacAddress.Zone)
	}
	_ = d.Set("flexible_ip_ids_to_duplicate", expandStrings(d.Get("flexible_ip_ids_to_duplicate").(*schema.Set).List()))

	return nil
}

func resourceScalewayFlexibleIPMACUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fipAPI, zone, err := fipAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	flexibleIP, err := waitFlexibleIP(ctx, fipAPI, zone, locality.ExpandID(d.Get("flexible_ip_id")), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("flexible_ip_id") {
		oldFipInterface, newFipInterface := d.GetChange("flexible_ip_id")
		oldFipID := locality.ExpandID(oldFipInterface.(string))
		newFipID := locality.ExpandID(newFipInterface.(string))

		res, err := fipAPI.MoveMACAddr(&flexibleip.MoveMACAddrRequest{
			Zone:     zone,
			FipID:    oldFipID,
			DstFipID: newFipID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		flexibleIP, err = waitFlexibleIP(ctx, fipAPI, zone, res.ID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("flexible_ip_ids_to_duplicate") {
		oldID, newID := d.GetChange("flexible_ip_ids_to_duplicate")
		oldIDs := expandStrings(oldID.(*schema.Set).List())
		newIDs := expandStrings(newID.(*schema.Set).List())

		// Handle added flexible IPs
		for _, newID := range newIDs {
			if !sliceContainsString(oldIDs, newID) {
				_, err = waitFlexibleIP(ctx, fipAPI, zone, locality.ExpandID(newID), d.Timeout(schema.TimeoutUpdate))
				if err != nil {
					return diag.FromErr(err)
				}
				_, err := fipAPI.DuplicateMACAddr(&flexibleip.DuplicateMACAddrRequest{
					Zone:               zone,
					FipID:              locality.ExpandID(newID),
					DuplicateFromFipID: flexibleIP.ID,
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
				_, err = waitFlexibleIP(ctx, fipAPI, zone, locality.ExpandID(newID), d.Timeout(schema.TimeoutUpdate))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
		// Handle removed flexible IPs
		for _, oldID := range oldIDs {
			if !sliceContainsString(newIDs, oldID) {
				err = fipAPI.DeleteMACAddr(&flexibleip.DeleteMACAddrRequest{
					Zone:  zone,
					FipID: locality.ExpandID(oldID),
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
				_, err = waitFlexibleIP(ctx, fipAPI, zone, locality.ExpandID(oldID), d.Timeout(schema.TimeoutUpdate))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	_, err = waitFlexibleIP(ctx, fipAPI, zone, flexibleIP.ID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayFlexibleIPMACRead(ctx, d, m)
}

func resourceScalewayFlexibleIPMACDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fipAPI, zone, err := fipAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	flexibleIP, err := waitFlexibleIP(ctx, fipAPI, zone, locality.ExpandID(d.Get("flexible_ip_id")), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	err = fipAPI.DeleteMACAddr(&flexibleip.DeleteMACAddrRequest{
		FipID: flexibleIP.MacAddress.ID,
		Zone:  zone,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) && !httperrors.Is403(err) {
		return diag.FromErr(err)
	}

	_, err = waitFlexibleIP(ctx, fipAPI, zone, locality.ExpandID(d.Get("flexible_ip_id")), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
