package list

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func OrganizationIDAttribute(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Description: description,
		Optional:    true,
		Validators: []validator.String{
			verify.IsStringUUID(),
		},
	}
}

func ProjectIDAttribute(description string) schema.ListAttribute {
	return schema.ListAttribute{
		Description: description + " Use '*' to list across all projects",
		Optional:    true,
		ElementType: types.StringType,
		Validators: []validator.List{
			listvalidator.ValueStringsAre(verify.IsStringUUID()),
		},
	}
}

func RegionAttribute(description string) schema.ListAttribute {
	return schema.ListAttribute{
		Description: description + " Use '*' to list from all regions",
		Optional:    true,
		ElementType: types.StringType,
		Validators: []validator.List{
			listvalidator.ValueStringsAre(
				stringvalidator.OneOf(append(regional.AllRegions(), "*")...)),
		},
	}
}

func TagsAttribute(description string) schema.ListAttribute {
	return schema.ListAttribute{
		Description: description,
		ElementType: types.StringType,
		Optional:    true,
	}
}

func NameAttribute(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Description: description,
		Optional:    true,
	}
}
