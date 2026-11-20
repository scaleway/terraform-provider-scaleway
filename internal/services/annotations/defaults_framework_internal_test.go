package annotations

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	frameworkSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	annotationsSDK "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testFrameworkResourceType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":  tftypes.String,
		"srn": tftypes.String,
	},
}

var testFrameworkSchema = frameworkSchema.Schema{
	Attributes: map[string]frameworkSchema.Attribute{
		"id":  frameworkSchema.StringAttribute{Computed: true},
		"srn": frameworkSchema.StringAttribute{Computed: true},
	},
}

// minimal mock Framework resource with a Computed "srn" attribute
type testFrameworkResource struct {
	stateAttrs map[string]tftypes.Value
}

func (r *testFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "scaleway_test"
}

func (r *testFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = testFrameworkSchema
}

func (r *testFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.writeState(&resp.State)
}

func (r *testFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *testFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.writeState(&resp.State)
}

func (r *testFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *testFrameworkResource) writeState(state *tfsdk.State) {
	if r.stateAttrs != nil {
		state.Raw = tftypes.NewValue(testFrameworkResourceType, r.stateAttrs)
	}
}

type testIdentityFrameworkResource struct {
	testFrameworkResource
}

func (r *testIdentityFrameworkResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{},
		},
		Version: 0,
	}
}

func testFrameworkStateAttrs(id, srn string) map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"id":  tftypes.NewValue(tftypes.String, id),
		"srn": tftypes.NewValue(tftypes.String, srn),
	}
}

func newTestFrameworkState() tfsdk.State {
	return tfsdk.State{Schema: testFrameworkSchema}
}

func testFrameworkState(id, srn string) tfsdk.State {
	state := newTestFrameworkState()
	state.Raw = tftypes.NewValue(testFrameworkResourceType, testFrameworkStateAttrs(id, srn))

	return state
}

func newTestCreateResponse() resource.CreateResponse {
	return resource.CreateResponse{State: newTestFrameworkState()}
}

func newConfiguredFrameworkWrapper(t *testing.T, server *fakeAnnotationsServer, wishlist map[string]string, inner resource.Resource) *frameworkDefaultBindingsWrapper {
	t.Helper()

	m := newTestMeta(t, server, wishlist)

	wrapped := WithDefaultBindingsFramework(func() resource.Resource { return inner })()

	w, ok := wrapped.(*frameworkDefaultBindingsWrapper)
	require.True(t, ok)

	var configureResp resource.ConfigureResponse

	w.Configure(t.Context(), resource.ConfigureRequest{ProviderData: m}, &configureResp)
	assert.False(t, configureResp.Diagnostics.HasError())

	return w
}

func TestWithDefaultBindingsFramework(t *testing.T) {
	t.Parallel()

	t.Run("create without default annotations is a no-op", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{}

		w := newConfiguredFrameworkWrapper(t, server, nil, &testFrameworkResource{
			stateAttrs: testFrameworkStateAttrs("id-1", testTargetSrn),
		})

		createResp := newTestCreateResponse()

		w.Create(t.Context(), resource.CreateRequest{}, &createResp)

		assert.False(t, createResp.Diagnostics.HasError())
		assert.Empty(t, server.created)
	})

	t.Run("create without meta is a no-op", func(t *testing.T) {
		t.Parallel()

		wrapped := WithDefaultBindingsFramework(func() resource.Resource {
			return &testFrameworkResource{stateAttrs: testFrameworkStateAttrs("id-1", testTargetSrn)}
		})()

		w, ok := wrapped.(*frameworkDefaultBindingsWrapper)
		require.True(t, ok)

		createResp := newTestCreateResponse()

		w.Create(t.Context(), resource.CreateRequest{}, &createResp)

		assert.False(t, createResp.Diagnostics.HasError())
	})

	t.Run("create reconciles the default annotations", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{keys: []*annotationsSDK.ListAllKeysAndValuesResponseKey{testKey("team")}}

		w := newConfiguredFrameworkWrapper(t, server, map[string]string{"team": "platform"}, &testFrameworkResource{
			stateAttrs: testFrameworkStateAttrs("id-1", testTargetSrn),
		})

		createResp := newTestCreateResponse()

		w.Create(t.Context(), resource.CreateRequest{}, &createResp)

		assert.False(t, createResp.Diagnostics.HasError())
		assertCreatedBinding(t, server, testValueName("team", "platform"))
	})

	t.Run("create reconciles bindings spread over multiple pages", func(t *testing.T) {
		t.Parallel()

		// The stale binding is on the second page: it can only be found (and
		// deleted) if all pages are fetched.
		server := &fakeAnnotationsServer{
			keys: []*annotationsSDK.ListAllKeysAndValuesResponseKey{testKey("team")},
			bindings: []*annotationsSDK.Binding{
				testBinding("b-env", "env", "platform"),
				testBinding("b-other", "other", "platform"),
				testBinding("b-stale", "team", "legacy"),
			},
		}

		w := newConfiguredFrameworkWrapper(t, server, map[string]string{"team": "platform"}, &testFrameworkResource{
			stateAttrs: testFrameworkStateAttrs("id-1", testTargetSrn),
		})

		createResp := newTestCreateResponse()

		w.Create(t.Context(), resource.CreateRequest{}, &createResp)

		assert.False(t, createResp.Diagnostics.HasError())

		assert.Equal(t, []string{"b-stale"}, server.deleted)
		assertCreatedBinding(t, server, testValueName("team", "platform"))
	})

	t.Run("create warns when a key cannot be resolved", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{}

		w := newConfiguredFrameworkWrapper(t, server, map[string]string{"missing": "platform"}, &testFrameworkResource{
			stateAttrs: testFrameworkStateAttrs("id-1", testTargetSrn),
		})

		createResp := newTestCreateResponse()

		w.Create(t.Context(), resource.CreateRequest{}, &createResp)

		assert.False(t, createResp.Diagnostics.HasError())

		require.Len(t, createResp.Diagnostics, 1)
		assert.Equal(t, diag.SeverityWarning, createResp.Diagnostics[0].Severity())
		assert.Equal(t, "Failed to apply default annotations", createResp.Diagnostics[0].Summary())
		assert.Contains(t, createResp.Diagnostics[0].Detail(), `annotation key "missing" not found`)
	})

	t.Run("create warns when the binding creation fails", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{
			keys:        []*annotationsSDK.ListAllKeysAndValuesResponseKey{testKey("team")},
			createError: http.StatusInternalServerError,
		}

		w := newConfiguredFrameworkWrapper(t, server, map[string]string{"team": "platform"}, &testFrameworkResource{
			stateAttrs: testFrameworkStateAttrs("id-1", testTargetSrn),
		})

		createResp := newTestCreateResponse()

		w.Create(t.Context(), resource.CreateRequest{}, &createResp)

		assert.False(t, createResp.Diagnostics.HasError())

		require.Len(t, createResp.Diagnostics, 1)
		assert.Equal(t, diag.SeverityWarning, createResp.Diagnostics[0].Severity())
		assert.Equal(t, "Failed to apply default annotations", createResp.Diagnostics[0].Summary())
	})

	t.Run("update reconciles the default annotations", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{keys: []*annotationsSDK.ListAllKeysAndValuesResponseKey{testKey("team")}}

		w := newConfiguredFrameworkWrapper(t, server, map[string]string{"team": "platform"}, &testFrameworkResource{
			stateAttrs: testFrameworkStateAttrs("id-1", testTargetSrn),
		})

		updateResp := resource.UpdateResponse{State: newTestFrameworkState()}

		w.Update(t.Context(), resource.UpdateRequest{}, &updateResp)

		assert.False(t, updateResp.Diagnostics.HasError())
		assertCreatedBinding(t, server, testValueName("team", "platform"))
	})

	t.Run("update without default annotations is a no-op", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{}

		w := newConfiguredFrameworkWrapper(t, server, nil, &testFrameworkResource{
			stateAttrs: testFrameworkStateAttrs("id-1", testTargetSrn),
		})

		updateResp := resource.UpdateResponse{State: newTestFrameworkState()}

		w.Update(t.Context(), resource.UpdateRequest{}, &updateResp)

		assert.False(t, updateResp.Diagnostics.HasError())
		assert.Empty(t, server.created)
	})

	t.Run("delete removes the default bindings", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{bindings: []*annotationsSDK.Binding{testBinding("b-team", "team", "platform")}}

		w := newConfiguredFrameworkWrapper(t, server, map[string]string{"team": "platform"}, &testFrameworkResource{})

		var deleteResp resource.DeleteResponse

		w.Delete(t.Context(), resource.DeleteRequest{State: testFrameworkState("id-1", testTargetSrn)}, &deleteResp)

		assert.False(t, deleteResp.Diagnostics.HasError())

		assert.Equal(t, []string{"b-team"}, server.deleted)
	})

	t.Run("import fails cleanly when the inner resource does not support import", func(t *testing.T) {
		t.Parallel()

		wrapped := WithDefaultBindingsFramework(func() resource.Resource { return &testFrameworkResource{} })()

		w, ok := wrapped.(*frameworkDefaultBindingsWrapper)
		require.True(t, ok)

		var importResp resource.ImportStateResponse

		w.ImportState(t.Context(), resource.ImportStateRequest{ID: "id-1"}, &importResp)

		require.True(t, importResp.Diagnostics.HasError())
		assert.Equal(t, "Resource Import Not Implemented", importResp.Diagnostics[0].Summary())
	})

	t.Run("identity resources are wrapped with identity support", func(t *testing.T) {
		t.Parallel()

		wrapped := WithDefaultBindingsFramework(func() resource.Resource { return &testIdentityFrameworkResource{} })()

		w, ok := wrapped.(*frameworkDefaultBindingsIdentityWrapper)
		require.True(t, ok)

		require.Implements(t, (*resource.ResourceWithIdentity)(nil), w)

		var identityResp resource.IdentitySchemaResponse

		w.IdentitySchema(t.Context(), resource.IdentitySchemaRequest{}, &identityResp)

		assert.False(t, identityResp.Diagnostics.HasError())
		assert.Contains(t, identityResp.IdentitySchema.Attributes, "id")
	})
}
