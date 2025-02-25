package webhosting

import (
	"github.com/scaleway/scaleway-sdk-go/api/webhosting/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func flattenOffer(offer *webhosting.Offer) interface{} {
	if offer == nil {
		return []interface{}{}
	}
	return []map[string]interface{}{
		{
			"id":                     offer.ID,
			"name":                   offer.Name,
			"billing_operation_path": offer.BillingOperationPath,
			"available":              offer.Available,
			"control_panel_name":     offer.ControlPanelName,
			"end_of_life":            offer.EndOfLife,
			"quota_warning":          string(offer.QuotaWarning),
			"price":                  flattenOfferPrice(offer.Price),
			"options":                flattenOfferOptions(offer.Options),
		},
	}
}

func flattenOfferOptions(options []*webhosting.OfferOption) interface{} {
	if options == nil {
		return []interface{}{}
	}
	res := make([]map[string]interface{}, 0, len(options))
	for _, option := range options {
		res = append(res, map[string]interface{}{
			"id":                     option.ID,
			"name":                   string(option.Name),
			"billing_operation_path": option.BillingOperationPath,
			"min_value":              option.MinValue,
			"current_value":          option.CurrentValue,
			"max_value":              option.MaxValue,
			"quota_warning":          string(option.QuotaWarning),
			"price":                  flattenOfferPrice(option.Price),
		})
	}
	return res
}

func flattenOfferPrice(price *scw.Money) interface{} {
	return price.String()
}

func flattenHostingCpanelUrls(cpanelURL *webhosting.PlatformControlPanelURLs) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"dashboard": cpanelURL.Dashboard,
			"webmail":   cpanelURL.Webmail,
		},
	}
}

func flattenHostingOptions(options []*webhosting.OfferOption) []map[string]interface{} {
	if options == nil {
		return nil
	}

	flattenedOptions := []map[string]interface{}(nil)
	for _, option := range options {
		flattenedOptions = append(flattenedOptions, map[string]interface{}{
			"id":   option.ID,
			"name": option.Name,
		})
	}

	return flattenedOptions
}

func expandOfferOptions(data interface{}) []*webhosting.OfferOptionRequest {
	optionIDs := types.ExpandStrings(data)
	offerOptions := make([]*webhosting.OfferOptionRequest, 0, len(optionIDs))
	for _, id := range optionIDs {
		if id == "" {
			continue
		}
		offerOptions = append(offerOptions, &webhosting.OfferOptionRequest{
			ID:       id,
			Quantity: 1,
		})
	}
	return offerOptions
}
