package object

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	sdkv2schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ list.ListResource                 = (*BucketListResource)(nil)
	_ list.ListResourceWithConfigure    = (*BucketListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*BucketListResource)(nil)
)

type BucketListResource struct {
	meta *meta.Meta
}

func (r *BucketListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
}

func NewBucketListResource() list.ListResource {
	return &BucketListResource{}
}

func (r *BucketListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_bucket"
}

func (r *BucketListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Project IDs to filter for. Use '*' to list across all projects",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.Any(
							stringvalidator.OneOf("*"),
							verify.IsStringUUID(),
						),
					),
				},
			},
			"regions": listscw.RegionsAttribute("Regions to filter for. Use '*' to list from all regions"),
			"name":    listscw.NameAttribute("Name of the bucket to filter on"),
			"tags":    listscw.TagsAttribute("Tags of the bucket to filter on"),
		},
	}
}

func (r *BucketListResource) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, response *list.RawV6SchemaResponse) {
	bucketResource := ResourceBucket()

	response.ProtoV6Schema = translate.Schema(bucketResource.ProtoSchema(ctx)())
	response.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(bucketResource.ProtoIdentitySchema(ctx)())
}

type BucketListResourceModel struct {
	ProjectIDs types.List   `tfsdk:"project_ids"`
	Regions    types.List   `tfsdk:"regions"`
	Name       types.String `tfsdk:"name"`
	Tags       types.List   `tfsdk:"tags"`
}

func (m *BucketListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (m *BucketListResourceModel) GetRegions() types.List {
	return m.Regions
}

func (m *BucketListResourceModel) GetTags() types.List {
	return m.Tags
}

type bucketRow struct {
	Bucket    *s3Types.Bucket
	Region    scw.Region
	ProjectID string
}

func (r *BucketListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data BucketListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	regions, err := listscw.ExtractRegions(ctx, &data, r.meta)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing regions", "An error was encountered when listing regions: "+err.Error()),
		})

		return
	}

	projects, err := listscw.ExtractProjects(ctx, &data, r.meta)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing projects", "An error was encountered when listing projects: "+err.Error()),
		})

		return
	}

	targets := make([]bucketListTarget, 0, len(regions)*len(projects))
	for _, region := range regions {
		for _, project := range projects {
			targets = append(targets, bucketListTarget{
				Region:    region,
				ProjectID: project,
			})
		}
	}

	allRows, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target bucketListTarget) ([]bucketRow, error) {
			return r.fetchBucketRows(ctx, target, data)
		},
		func(a, b bucketRow) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.Region != b.Region {
				return strings.Compare(string(a.Region), string(b.Region))
			}

			if a.Bucket.Name == nil || b.Bucket.Name == nil {
				return 0
			}

			return strings.Compare(*a.Bucket.Name, *b.Bucket.Name)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing object buckets", "Failed to list object buckets: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, row := range allRows {
			result := req.NewListResult(ctx)
			result.DisplayName = aws.ToString(row.Bucket.Name)

			bucketResource := ResourceBucket()
			resourceData := bucketResource.Data(&terraform.InstanceState{})

			bucketID := regional.NewIDString(row.Region, aws.ToString(row.Bucket.Name))

			err := identity.SetRegionalIdentity(resourceData, row.Region, bucketID)
			if err != nil {
				result.Diagnostics.AddError(
					"Retrieving identity data",
					"An error was encountered when retrieving the identity data: "+err.Error(),
				)

				if !push(result) {
					return
				}

				continue
			}

			tfTypeIdentity, errIdentityState := resourceData.TfTypeIdentityState()
			if errIdentityState != nil {
				result.Diagnostics.AddError(
					"Converting identity data",
					"An error was encountered when converting the identity data: "+errIdentityState.Error(),
				)
			}

			identitySetDiags := result.Identity.Set(ctx, *tfTypeIdentity)
			result.Diagnostics.Append(identitySetDiags...)

			setBucketState(resourceData, row.Bucket, row.Region, row.ProjectID)

			tfTypeResource, errTfTypeResourceState := resourceData.TfTypeResourceState()
			if errTfTypeResourceState != nil {
				result.Diagnostics.AddError(
					"Converting resource state",
					"An error was encountered when converting the resource state: "+errTfTypeResourceState.Error(),
				)
			}

			resourceSetDiags := result.Resource.Set(ctx, *tfTypeResource)
			result.Diagnostics.Append(resourceSetDiags...)

			if !push(result) {
				return
			}
		}
	}
}

type bucketListTarget struct {
	Region    scw.Region
	ProjectID string
}

func (r *BucketListResource) fetchBucketRows(ctx context.Context, target bucketListTarget, data BucketListResourceModel) ([]bucketRow, error) {
	accessKey, _ := r.meta.ScwClient().GetAccessKey()
	accessKey = accessKeyWithProjectID(accessKey, target.ProjectID)

	secretKey, _ := r.meta.ScwClient().GetSecretKey()

	s3Client, err := newS3Client(ctx, target.Region.String(), accessKey, secretKey, r.meta.HTTPClient())
	if err != nil {
		return nil, fmt.Errorf("creating S3 client: %w", err)
	}

	listInput := &s3.ListBucketsInput{}

	resp, err := s3Client.ListBuckets(ctx, listInput)
	if err != nil {
		return nil, fmt.Errorf("listing buckets: %w", err)
	}

	var rows []bucketRow

	for i := range resp.Buckets {
		bucket := &resp.Buckets[i]

		// Filter by name if specified
		if !data.Name.IsNull() && !data.Name.IsUnknown() {
			expectedName := data.Name.ValueString()
			if bucket.Name == nil || *bucket.Name != expectedName {
				continue
			}
		}

		// Filter by tags if specified
		if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
			var filterTagStrings []string
			diags := data.Tags.ElementsAs(ctx, &filterTagStrings, false)
			if diags.HasError() {
				continue
			}

			if len(filterTagStrings) > 0 {
				bucketTags, err := r.getBucketTags(ctx, s3Client, bucket.Name)
				if err != nil || !tagsMatch(bucketTags, filterTagStrings) {
					continue
				}
			}
		}

		rows = append(rows, bucketRow{
			Bucket:    bucket,
			Region:    target.Region,
			ProjectID: target.ProjectID,
		})
	}

	return rows, nil
}

func (r *BucketListResource) getBucketTags(ctx context.Context, client *s3.Client, bucketName *string) ([]string, error) {
	if bucketName == nil {
		return nil, nil
	}

	resp, err := client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
		Bucket: bucketName,
	})
	if err != nil {
		return nil, err
	}

	tags := make([]string, 0, len(resp.TagSet))
	for _, tag := range resp.TagSet {
		if tag.Key != nil {
			tags = append(tags, *tag.Key)
		}
	}

	return tags, nil
}

func tagsMatch(bucketTags []string, filterTags []string) bool {
	if len(filterTags) == 0 {
		return true
	}

	// Check if all filter tags are present in bucket tags
	for _, filterTag := range filterTags {
		if !slices.Contains(bucketTags, filterTag) {
			return false
		}
	}

	return true
}

func setBucketState(resourceData *sdkv2schema.ResourceData, bucket *s3Types.Bucket, region scw.Region, projectID string) {
	if bucket.Name != nil {
		_ = resourceData.Set("name", *bucket.Name)
	}

	_ = resourceData.Set("region", region.String())
	_ = resourceData.Set("project_id", projectID)

	// Set computed fields
	_ = resourceData.Set("endpoint", objectBucketEndpointURL(aws.ToString(bucket.Name), region))
	_ = resourceData.Set("api_endpoint", objectBucketAPIEndpointURL(region))
}
