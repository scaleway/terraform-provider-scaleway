package webhosting

import (
	"github.com/scaleway/scaleway-sdk-go/api/webhosting/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func flattenOffer(offer *webhosting.Offer) any {
	if offer == nil {
		return []any{}
	}

	return []map[string]any{
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

func flattenOfferOptions(options []*webhosting.OfferOption) any {
	if options == nil {
		return []any{}
	}

	res := make([]map[string]any, 0, len(options))

	for _, option := range options {
		res = append(res, map[string]any{
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

func flattenOfferPrice(price *scw.Money) any {
	return price.String()
}

func flattenHostingCpanelUrls(cpanelURL *webhosting.PlatformControlPanelURLs) []map[string]any {
	return []map[string]any{
		{
			"dashboard": cpanelURL.Dashboard,
			"webmail":   cpanelURL.Webmail,
		},
	}
}

func flattenHostingOptions(options []*webhosting.OfferOption) []map[string]any {
	if options == nil {
		return nil
	}

	flattenedOptions := []map[string]any(nil)
	for _, option := range options {
		flattenedOptions = append(flattenedOptions, map[string]any{
			"id":   option.ID,
			"name": option.Name,
		})
	}

	return flattenedOptions
}

func expandOfferOptions(data any) []*webhosting.OfferOptionRequest {
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
