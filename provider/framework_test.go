package provider_test

import (
	"context"
	"reflect"
	"runtime"
	"slices"
	"testing"

	actionFramework "github.com/hashicorp/terraform-plugin-framework/action"
	ephemeralFramework "github.com/hashicorp/terraform-plugin-framework/ephemeral"
	providerFramework "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/provider"
	"github.com/stretchr/testify/assert"
)

func extractActionDescriptions(ctx context.Context, a actionFramework.Action) (string, string) {
	resp := &actionFramework.SchemaResponse{}
	a.Schema(ctx, actionFramework.SchemaRequest{}, resp)

	return resp.Schema.Description, resp.Schema.MarkdownDescription
}

func extractActionSchemaMethod(action actionFramework.Action) (reflect.Type, reflect.Method, bool) {
	methodName := "Schema"
	actionType := reflect.TypeOf(action)
	method, found := actionType.MethodByName(methodName)

	return actionType, method, found
}

func TestProviderActionDescriptionsAreNotEmpty(t *testing.T) {
	p := provider.NewFrameworkProvider(nil)().(providerFramework.ProviderWithActions)
	for _, action := range p.Actions(t.Context()) {
		description, markdownDescription := extractActionDescriptions(t.Context(), action())
		actionType, method, found := extractActionSchemaMethod(action())

		if found {
			funcPtr := method.Func.Pointer()
			fn := runtime.FuncForPC(funcPtr)
			file, line := fn.FileLine(funcPtr)
			assert.NotEmpty(t, description, "Please fill up Description field in %s schema, %s:%d", actionType, file, line)
			assert.NotEmpty(t, markdownDescription, "Please fill up MarkdownDescription field in %s schema, %s:%d", actionType, file, line)
		} else {
			t.Errorf("No Schema function found of the action %s", actionType)
		}
	}
}

func extractEphemeralDescriptions(ctx context.Context, a ephemeralFramework.EphemeralResource) (string, string) {
	resp := &ephemeralFramework.SchemaResponse{}
	a.Schema(ctx, ephemeralFramework.SchemaRequest{}, resp)

	return resp.Schema.Description, resp.Schema.MarkdownDescription
}

func extractEphemeralSchemaMethod(ephemeral ephemeralFramework.EphemeralResource) (reflect.Type, reflect.Method, bool) {
	methodName := "Schema"
	ephemeralType := reflect.TypeOf(ephemeral)
	method, found := ephemeralType.MethodByName(methodName)

	return ephemeralType, method, found
}

func TestProviderEphemeralDescriptionsAreNotEmpty(t *testing.T) {
	p := provider.NewFrameworkProvider(nil)().(providerFramework.ProviderWithEphemeralResources)
	for _, ephemeral := range p.EphemeralResources(t.Context()) {
		description, markdownDescription := extractEphemeralDescriptions(t.Context(), ephemeral())
		ephemeralType, method, found := extractEphemeralSchemaMethod(ephemeral())

		if found {
			funcPtr := method.Func.Pointer()
			fn := runtime.FuncForPC(funcPtr)
			file, line := fn.FileLine(funcPtr)
			assert.NotEmpty(t, description, "Please fill up Description field in %s schema, %s:%d", ephemeralType, file, line)
			assert.NotEmpty(t, markdownDescription, "Please fill up MarkdownDescription field in %s schema, %s:%d", ephemeralType, file, line)
		} else {
			t.Errorf("No Schema function found for the ephemeral resource %s", ephemeralType)
		}
	}
}

func TestProviderActionNameDoesNotContainAction(t *testing.T) {
	exceptions := []string{"_instance_server_action", "_baremetal_server_action"}

	p := provider.NewFrameworkProvider(nil)().(providerFramework.ProviderWithActions)
	for _, action := range p.Actions(t.Context()) {
		resp := &actionFramework.MetadataResponse{}
		action().Metadata(t.Context(), actionFramework.MetadataRequest{}, resp)

		if !slices.Contains(exceptions, resp.TypeName) {
			assert.NotContains(t, resp.TypeName, "action")
		}
	}
}

// TestProviderAnnotationsBindingResourceIsWrapped verifies that
// scaleway_annotations_binding is registered with the default annotation
// binding wrapper, so that it participates in default_annotations like
// any other resource with a Computed "srn" attribute.
func TestProviderAnnotationsBindingResourceIsWrapped(t *testing.T) {
	p := provider.NewFrameworkProvider(nil)()
	found := false

	for _, factory := range p.Resources(t.Context()) {
		r := factory()

		resp := &resource.SchemaResponse{}
		r.Schema(t.Context(), resource.SchemaRequest{}, resp)

		if resp.Schema.Attributes["srn"] != nil {
			if strAttr, ok := resp.Schema.Attributes["srn"].(schema.StringAttribute); ok && strAttr.Computed {
				// Verify this is the binding resource by checking its type name.
				metaResp := &resource.MetadataResponse{}
				r.Metadata(t.Context(), resource.MetadataRequest{ProviderTypeName: "scaleway"}, metaResp)

				if metaResp.TypeName == "scaleway_annotations_binding" {
					found = true
				}
			}
		}
	}

	assert.True(t, found, "scaleway_annotations_binding must be registered with the default annotation binding wrapper")
}
