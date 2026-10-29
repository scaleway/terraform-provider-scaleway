package list

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func ConfigureMeta(request resource.ConfigureRequest, response *resource.ConfigureResponse) *meta.Meta {
	if request.ProviderData == nil {
		return nil
	}

	m, ok := request.ProviderData.(*meta.Meta)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected List Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return nil
	}

	return m
}
