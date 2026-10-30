package object

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ resource.Resource                = (*ObjectResource)(nil)
	_ resource.ResourceWithConfigure   = (*ObjectResource)(nil)
	_ resource.ResourceWithImportState = (*ObjectResource)(nil)
)

func NewObjectResource() resource.Resource {
	return &ObjectResource{}
}

type ObjectResource struct {
	meta *meta.Meta
}

type objectResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Bucket         types.String `tfsdk:"bucket"`
	Key            types.String `tfsdk:"key"`
	File           types.String `tfsdk:"file"`
	Content        types.String `tfsdk:"content"`
	ContentBase64  types.String `tfsdk:"content_base64"`
	Hash           types.String `tfsdk:"hash"`
	StorageClass   types.String `tfsdk:"storage_class"`
	Metadata       types.Map    `tfsdk:"metadata"`
	ContentType    types.String `tfsdk:"content_type"`
	Tags           types.Map    `tfsdk:"tags"`
	Visibility     types.String `tfsdk:"visibility"`
	SSECustomerKey types.String `tfsdk:"sse_customer_key"`
	Region         types.String `tfsdk:"region"`
	ProjectID      types.String `tfsdk:"project_id"`
}

func (r *ObjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object"
}

func (r *ObjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Scaleway Object Storage object.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The regional ID of the object (region/bucket/key).",
			},
			"bucket": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The bucket's name or regional ID.",
				PlanModifiers: []planmodifier.String{
					regional.LocalityPlanModifier(),
				},
			},
			"key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Key of the object",
			},
			"file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path of the file to upload, defaults to an empty file",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("content"), path.MatchRoot("content_base64")),
				},
			},
			"content": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Content of the file to upload",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("file"), path.MatchRoot("content_base64")),
				},
			},
			"content_base64": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Content of the file to upload, should be base64 encoded",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("file"), path.MatchRoot("content")),
				},
			},
			"hash": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "File hash to trigger upload",
			},
			"storage_class": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Specifies the Scaleway Object Storage class to which you want the object to transition",
				Validators: []validator.String{
					stringvalidator.OneOf(TransitionSCWStorageClassValues()...),
				},
			},
			"metadata": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Map of object's metadata, only lower case keys are allowed",
				Validators: []validator.Map{
					lowercaseMapKeysValidator{},
				},
			},
			"content_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The standard MIME type of the object's content (e.g., 'application/json', 'text/plain'). This specifies how the object should be interpreted by clients. See RFC 9110: https://www.rfc-editor.org/rfc/rfc9110.html#name-content-type",
			},
			"tags": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Map of object's tags",
			},
			"visibility": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Visibility of the object, public-read or private",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(s3Types.ObjectCannedACLPrivate),
						string(s3Types.ObjectCannedACLPublicRead),
					),
				},
			},
			"sse_customer_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Customer's encryption keys to encrypt data (SSE-C)",
				Validators: []validator.String{
					stringvalidator.LengthBetween(32, 32),
				},
			},
			"region":     regional.SchemaAttributeComputed("The region you want to attach the resource to"),
			"project_id": projectIDFrameworkAttribute(),
		},
	}
}

func (r *ObjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.meta = m
}

func (r *ObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan objectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, s3Client, err := r.resolveBucketAndClient(ctx, plan.Bucket, plan.Region, plan.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	key := plan.Key.ValueString()

	putReq := &s3.PutObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		StorageClass: s3Types.StorageClass(plan.StorageClass.ValueString()),
		Metadata:     expandStringMap(ctx, plan.Metadata, &resp.Diagnostics),
	}

	if !plan.ContentType.IsNull() && !plan.ContentType.IsUnknown() && plan.ContentType.ValueString() != "" {
		putReq.ContentType = aws.String(plan.ContentType.ValueString())
	}

	if visibility := plan.Visibility.ValueString(); visibility != "" {
		putReq.ACL = s3Types.ObjectCannedACL(visibility)
	}

	if err := setObjectEncryption(plan.SSECustomerKey, putReq); err != nil {
		resp.Diagnostics.AddError("Failed to encrypt customer key", err.Error())

		return
	}

	body, closeFn, err := objectBody(plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read object body", err.Error())

		return
	}

	if closeFn != nil {
		defer closeFn()
	}

	putReq.Body = body

	_, err = s3Client.PutObject(ctx, putReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to put object", err.Error())

		return
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() && len(plan.Tags.Elements()) > 0 {
		_, err = s3Client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Tagging: &s3Types.Tagging{
				TagSet: expandObjectTags(ctx, plan.Tags, &resp.Diagnostics),
			},
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to put object tagging", err.Error())

			return
		}
	}

	plan.ID = types.StringValue(regional.NewIDString(region, objectID(bucket, key)))

	r.readInto(ctx, &plan, s3Client, region, bucket, key, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state objectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, key, s3Client, err := r.clientFromState(ctx, state.ID, state.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	r.readInto(ctx, &state, s3Client, region, bucket, key, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  objectResourceModel
		state objectResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, oldBucket, oldKey, s3Client, err := r.clientFromState(ctx, state.ID, state.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	bucketUpdated := regional.ExpandID(plan.Bucket.ValueString()).ID
	keyUpdated := plan.Key.ValueString()

	reupload := !plan.File.Equal(state.File) || !plan.Hash.Equal(state.Hash)

	if reupload {
		putReq := &s3.PutObjectInput{
			Bucket:       aws.String(bucketUpdated),
			Key:          aws.String(keyUpdated),
			StorageClass: s3Types.StorageClass(plan.StorageClass.ValueString()),
			Metadata:     expandStringMap(ctx, plan.Metadata, &resp.Diagnostics),
			ACL:          s3Types.ObjectCannedACL(plan.Visibility.ValueString()),
		}

		if !plan.ContentType.IsNull() && !plan.ContentType.IsUnknown() && plan.ContentType.ValueString() != "" {
			putReq.ContentType = aws.String(plan.ContentType.ValueString())
		}

		if err := setObjectEncryption(plan.SSECustomerKey, putReq); err != nil {
			resp.Diagnostics.AddError("Failed to encrypt customer key", err.Error())

			return
		}

		body, closeFn, err := objectBody(plan)
		if err != nil {
			resp.Diagnostics.AddError("Failed to read object body", err.Error())

			return
		}

		if closeFn != nil {
			defer closeFn()
		}

		putReq.Body = body

		_, err = s3Client.PutObject(ctx, putReq)
	} else {
		copyReq := &s3.CopyObjectInput{
			Bucket:       aws.String(bucketUpdated),
			Key:          aws.String(keyUpdated),
			StorageClass: s3Types.StorageClass(plan.StorageClass.ValueString()),
			CopySource:   aws.String(fmt.Sprintf("%s/%s", oldBucket, oldKey)),
			Metadata:     expandStringMap(ctx, plan.Metadata, &resp.Diagnostics),
			ACL:          s3Types.ObjectCannedACL(plan.Visibility.ValueString()),
		}

		if !plan.ContentType.IsNull() && !plan.ContentType.IsUnknown() && plan.ContentType.ValueString() != "" {
			copyReq.ContentType = aws.String(plan.ContentType.ValueString())
		}

		if err := setCopyObjectEncryption(plan.SSECustomerKey, copyReq); err != nil {
			resp.Diagnostics.AddError("Failed to encrypt customer key", err.Error())

			return
		}

		_, err = s3Client.CopyObject(ctx, copyReq)
	}

	if err != nil {
		resp.Diagnostics.AddError("Failed to update object", err.Error())

		return
	}

	if !plan.Key.Equal(state.Key) || !plan.Bucket.Equal(state.Bucket) {
		_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Key:    aws.String(oldKey),
			Bucket: aws.String(oldBucket),
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to delete old object", err.Error())

			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		_, err := s3Client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
			Bucket: aws.String(bucketUpdated),
			Key:    aws.String(keyUpdated),
			Tagging: &s3Types.Tagging{
				TagSet: expandObjectTags(ctx, plan.Tags, &resp.Diagnostics),
			},
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to put object tagging", err.Error())

			return
		}
	}

	plan.ID = types.StringValue(regional.NewIDString(scw.Region(state.Region.ValueString()), objectID(bucketUpdated, keyUpdated)))

	r.readInto(ctx, &plan, s3Client, scw.Region(state.Region.ValueString()), bucketUpdated, keyUpdated, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state objectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, bucket, key, s3Client, err := r.clientFromState(ctx, state.ID, state.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	_, err = s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete object", err.Error())

		return
	}
}

func (r *ObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ObjectResource) resolveBucketAndClient(ctx context.Context, bucketAttr, regionAttr, projectIDAttr types.String) (scw.Region, string, *s3.Client, error) {
	region, err := meta.ExtractFrameworkRegion(regionAttr, r.meta.ScwClient())
	if err != nil {
		return "", "", nil, err
	}

	regionalID := regional.ExpandID(bucketAttr.ValueString())
	bucket := regionalID.ID

	if regionalID.Region != "" && regionalID.Region != region {
		region = regionalID.Region
	}

	projectID, err := meta.ExtractFrameworkProjectID(projectIDAttr, r.meta.ScwClient())
	if err != nil {
		return "", "", nil, err
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, r.meta, region.String(), projectID)
	if err != nil {
		return "", "", nil, err
	}

	return region, bucket, s3Client, nil
}

func (r *ObjectResource) clientFromState(ctx context.Context, idAttr, projectIDAttr types.String) (scw.Region, string, string, *s3.Client, error) {
	region, key, bucket, err := regional.ParseNestedID(idAttr.ValueString())
	if err != nil {
		return "", "", "", nil, err
	}

	projectID, err := meta.ExtractFrameworkProjectID(projectIDAttr, r.meta.ScwClient())
	if err != nil {
		return "", "", "", nil, err
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, r.meta, region.String(), projectID)
	if err != nil {
		return "", "", "", nil, err
	}

	return region, bucket, key, s3Client, nil
}

func (r *ObjectResource) readInto(
	ctx context.Context,
	model *objectResourceModel,
	s3Client *s3.Client,
	region scw.Region,
	bucket string,
	key string,
	diags *diag.Diagnostics,
) {
	model.Region = types.StringValue(region.String())

	if model.Bucket.IsNull() {
		model.Bucket = types.StringValue(regional.NewIDString(region, bucket))
	}

	model.Key = types.StringValue(key)

	headReq := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if !model.SSECustomerKey.IsNull() && !model.SSECustomerKey.IsUnknown() && model.SSECustomerKey.ValueString() != "" {
		headReq.SSECustomerKey = aws.String(base64.StdEncoding.EncodeToString([]byte(model.SSECustomerKey.ValueString())))
		headReq.SSECustomerAlgorithm = aws.String("AES256")
	}

	obj, err := s3Client.HeadObject(ctx, headReq)
	if err != nil {
		diags.AddError("Failed to head object", err.Error())

		return
	}

	for k, v := range obj.Metadata {
		if k != strings.ToLower(k) {
			obj.Metadata[strings.ToLower(k)] = v
			delete(obj.Metadata, k)
		}
	}

	metadataVal, d := types.MapValueFrom(ctx, types.StringType, obj.Metadata)
	diags.Append(d...)
	model.Metadata = metadataVal

	if obj.ContentType != nil {
		model.ContentType = types.StringValue(aws.ToString(obj.ContentType))
	}

	tags, err := s3Client.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		diags.AddError("Failed to get object tagging", err.Error())

		return
	}

	model.Tags = flattenObjectTags(ctx, tags.TagSet, diags)

	acl, err := s3Client.GetObjectAcl(ctx, &s3.GetObjectAclInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		diags.AddError("Failed to get object acl", err.Error())

		return
	}

	if objectIsPublic(acl) {
		model.Visibility = types.StringValue(string(s3Types.ObjectCannedACLPublicRead))
	} else {
		model.Visibility = types.StringValue(string(s3Types.ObjectCannedACLPrivate))
	}

	if acl != nil && acl.Owner != nil {
		model.ProjectID = types.StringValue(aws.ToString(NormalizeOwnerID(acl.Owner.ID)))
	}
}

func setObjectEncryption(sseCustomerKey types.String, putReq *s3.PutObjectInput) error {
	if sseCustomerKey.IsNull() || sseCustomerKey.IsUnknown() || sseCustomerKey.ValueString() == "" {
		return nil
	}

	digestMD5, encryption, err := EncryptCustomerKey(sseCustomerKey.ValueString())
	if err != nil {
		return err
	}

	putReq.SSECustomerAlgorithm = aws.String("AES256")
	putReq.SSECustomerKeyMD5 = &digestMD5
	putReq.SSECustomerKey = encryption

	return nil
}

func setCopyObjectEncryption(sseCustomerKey types.String, copyReq *s3.CopyObjectInput) error {
	if sseCustomerKey.IsNull() || sseCustomerKey.IsUnknown() || sseCustomerKey.ValueString() == "" {
		return nil
	}

	digestMD5, encryption, err := EncryptCustomerKey(sseCustomerKey.ValueString())
	if err != nil {
		return err
	}

	copyReq.CopySourceSSECustomerAlgorithm = aws.String("AES256")
	copyReq.CopySourceSSECustomerKeyMD5 = &digestMD5
	copyReq.CopySourceSSECustomerKey = encryption

	return nil
}

// objectBody builds the request body from one of file/content/content_base64,
// defaulting to an empty object. The returned close function (if non-nil) must
// be deferred by the caller to release any opened file handle.
func objectBody(plan objectResourceModel) (body io.Reader, closeFn func() error, err error) {
	if file := plan.File.ValueString(); file != "" && !plan.File.IsNull() {
		f, ferr := os.Open(file)
		if ferr != nil {
			return nil, nil, ferr
		}

		return f, f.Close, nil
	}

	if content := plan.Content.ValueString(); content != "" && !plan.Content.IsNull() {
		return bytes.NewReader([]byte(content)), nil, nil
	}

	if encoded := plan.ContentBase64.ValueString(); encoded != "" && !plan.ContentBase64.IsNull() {
		raw := []byte(encoded)
		decoded := make([]byte, base64.StdEncoding.DecodedLen(len(raw)))

		n, derr := base64.StdEncoding.Decode(decoded, raw)
		if derr != nil {
			return nil, nil, derr
		}

		return bytes.NewReader(decoded[:n]), nil, nil
	}

	return bytes.NewReader([]byte{}), nil, nil
}

func expandStringMap(ctx context.Context, m types.Map, diags *diag.Diagnostics) map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}

	out := map[string]string{}
	ddiags := m.ElementsAs(ctx, &out, false)
	diags.Append(ddiags...)

	return out
}

func expandObjectTags(ctx context.Context, tagsMap types.Map, diags *diag.Diagnostics) []s3Types.Tag {
	if tagsMap.IsNull() || tagsMap.IsUnknown() {
		return nil
	}

	out := map[string]string{}
	ddiags := tagsMap.ElementsAs(ctx, &out, false)
	diags.Append(ddiags...)

	tags := make([]s3Types.Tag, 0, len(out))
	for k, v := range out {
		tags = append(tags, s3Types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	return tags
}

func flattenObjectTags(ctx context.Context, tagSet []s3Types.Tag, diags *diag.Diagnostics) types.Map {
	if len(tagSet) == 0 {
		return types.MapNull(types.StringType)
	}

	tags := make(map[string]string, len(tagSet))
	for _, t := range tagSet {
		tags[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}

	val, d := types.MapValueFrom(ctx, types.StringType, tags)
	diags.Append(d...)

	return val
}

// lowercaseMapKeysValidator validates that all keys of a map attribute are
// lowercase, mirroring the SDKv2 validateMapKeyLowerCase helper.
type lowercaseMapKeysValidator struct{}

func (lowercaseMapKeysValidator) Description(_ context.Context) string {
	return "Map keys must be lowercase."
}

func (v lowercaseMapKeysValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (lowercaseMapKeysValidator) ValidateMap(_ context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	for k := range req.ConfigValue.Elements() {
		if strings.ToLower(k) != k {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtMapKey(k),
				"Invalid map content",
				fmt.Sprintf("key (%s) should be lowercase", k),
			)
		}
	}
}

func objectID(bucket, key string) string {
	return fmt.Sprintf("%s/%s", bucket, key)
}

func objectIsPublic(acl *s3.GetObjectAclOutput) bool {
	for _, grant := range acl.Grants {
		if grant.Grantee != nil &&
			grant.Grantee.Type == s3Types.TypeGroup &&
			grant.Grantee.URI != nil &&
			*grant.Grantee.URI == AllUsersURI {
			return true
		}
	}

	return false
}

func EncryptCustomerKey(encryptionKeyStr string) (string, *string, error) {
	encryptionKey := []byte(encryptionKeyStr)
	h := md5.New() //nolint:gosec

	_, err := h.Write(encryptionKey)
	if err != nil {
		return "", nil, err
	}

	digest := h.Sum(nil)
	digestMD5 := base64.StdEncoding.EncodeToString(digest)
	encryption := aws.String(base64.StdEncoding.EncodeToString(encryptionKey))

	return digestMD5, encryption, nil
}
