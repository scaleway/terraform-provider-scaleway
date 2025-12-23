package provider_test

import (
	"context"
	"reflect"
	"runtime"
	"testing"

	actionFramework "github.com/hashicorp/terraform-plugin-framework/action"
	ephemeralFramework "github.com/hashicorp/terraform-plugin-framework/ephemeral"
	providerFramework "github.com/hashicorp/terraform-plugin-framework/provider"
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
