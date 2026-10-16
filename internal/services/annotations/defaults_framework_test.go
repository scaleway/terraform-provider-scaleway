package annotations_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/annotations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubFrameworkResource is a minimal Framework resource for testing the
// binding wrapper. It implements all the optional interfaces used in this
// codebase (Configure, ImportState, Identity).
type stubFrameworkResource struct {
	srnAttribute    schema.StringAttribute
	hasSrn          bool
	createCalled    bool
	deleteCalled    bool
	configureCalled bool
	importCalled    bool
	identityCalled  bool
}

func (r *stubFrameworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stub"
}

func (r *stubFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := map[string]schema.Attribute{}

	if r.hasSrn {
		attrs["srn"] = r.srnAttribute
	}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (r *stubFrameworkResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	r.createCalled = true
}

func (r *stubFrameworkResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

func (r *stubFrameworkResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *stubFrameworkResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	r.deleteCalled = true
}

func (r *stubFrameworkResource) Configure(_ context.Context, _ resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	r.configureCalled = true
}

func (r *stubFrameworkResource) ImportState(_ context.Context, _ resource.ImportStateRequest, _ *resource.ImportStateResponse) {
	r.importCalled = true
}

func (r *stubFrameworkResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, _ *resource.IdentitySchemaResponse) {
	r.identityCalled = true
}

func wrapStub(stub *stubFrameworkResource) resource.Resource {
	factory := annotations.WithDefaultBindingsFramework(func() resource.Resource { return stub })

	return factory()
}

func TestWithDefaultBindingsFramework_WrapsCreateAndDelete(t *testing.T) {
	stub := &stubFrameworkResource{
		hasSrn:       true,
		srnAttribute: schema.StringAttribute{Computed: true},
	}

	wrapped := wrapStub(stub)

	// Configure to inject meta (no default annotations → binding ops are no-ops)
	configureReq := resource.ConfigureRequest{ProviderData: &meta.Meta{}}
	configureResp := &resource.ConfigureResponse{}

	wrapped.(resource.ResourceWithConfigure).Configure(context.Background(), configureReq, configureResp)

	require.False(t, configureResp.Diagnostics.HasError())
	assert.True(t, stub.configureCalled)

	// Create should delegate to inner
	createReq := resource.CreateRequest{}
	createResp := &resource.CreateResponse{}

	wrapped.Create(context.Background(), createReq, createResp)

	require.False(t, createResp.Diagnostics.HasError())
	assert.True(t, stub.createCalled)

	// Delete should delegate to inner
	deleteReq := resource.DeleteRequest{}
	deleteResp := &resource.DeleteResponse{}

	wrapped.Delete(context.Background(), deleteReq, deleteResp)

	require.False(t, deleteResp.Diagnostics.HasError())
	assert.True(t, stub.deleteCalled)
}

func TestWithDefaultBindingsFramework_NoSrnSchemaIsNoOp(t *testing.T) {
	stub := &stubFrameworkResource{
		hasSrn: false,
	}

	wrapped := wrapStub(stub)

	configureReq := resource.ConfigureRequest{ProviderData: &meta.Meta{}}
	configureResp := &resource.ConfigureResponse{}

	wrapped.(resource.ResourceWithConfigure).Configure(context.Background(), configureReq, configureResp)

	createReq := resource.CreateRequest{}
	createResp := &resource.CreateResponse{}

	wrapped.Create(context.Background(), createReq, createResp)

	require.False(t, createResp.Diagnostics.HasError())
	assert.True(t, stub.createCalled)
}

func TestWithDefaultBindingsFramework_RequiredSrnSkipsBindings(t *testing.T) {
	stub := &stubFrameworkResource{
		hasSrn:       true,
		srnAttribute: schema.StringAttribute{Required: true},
	}

	wrapped := wrapStub(stub)

	configureReq := resource.ConfigureRequest{ProviderData: &meta.Meta{}}
	configureResp := &resource.ConfigureResponse{}

	wrapped.(resource.ResourceWithConfigure).Configure(context.Background(), configureReq, configureResp)

	createReq := resource.CreateRequest{}
	createResp := &resource.CreateResponse{}

	wrapped.Create(context.Background(), createReq, createResp)

	// Create should succeed and call inner, but should NOT create bindings
	// because "srn" is Required (user-provided), not Computed.
	require.False(t, createResp.Diagnostics.HasError())
	assert.True(t, stub.createCalled)
}

func TestWithDefaultBindingsFramework_DelegatesOptionalInterfaces(t *testing.T) {
	stub := &stubFrameworkResource{
		hasSrn:       true,
		srnAttribute: schema.StringAttribute{Computed: true},
	}

	wrapped := wrapStub(stub)

	// Verify the wrapper implements the optional interfaces
	_, ok := wrapped.(resource.ResourceWithImportState)
	assert.True(t, ok, "wrapper should implement ResourceWithImportState")

	_, ok = wrapped.(resource.ResourceWithIdentity)
	assert.True(t, ok, "wrapper should implement ResourceWithIdentity")

	_, ok = wrapped.(resource.ResourceWithConfigure)
	assert.True(t, ok, "wrapper should implement ResourceWithConfigure")

	// ImportState should delegate to inner
	importReq := resource.ImportStateRequest{}
	importResp := &resource.ImportStateResponse{}

	wrapped.(resource.ResourceWithImportState).ImportState(context.Background(), importReq, importResp)

	assert.True(t, stub.importCalled)

	// IdentitySchema should delegate to inner
	identityReq := resource.IdentitySchemaRequest{}
	identityResp := &resource.IdentitySchemaResponse{}

	wrapped.(resource.ResourceWithIdentity).IdentitySchema(context.Background(), identityReq, identityResp)

	assert.True(t, stub.identityCalled)
}
