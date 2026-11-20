package provider_test

import (
	"context"
	"reflect"
	"runtime"
	"slices"
	"testing"

	actionFramework "github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	ephemeralFramework "github.com/hashicorp/terraform-plugin-framework/ephemeral"
	providerFramework "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/scaleway/terraform-provider-scaleway/v2/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func extractResourceDescriptions(ctx context.Context, r resource.Resource) (string, string) {
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	return resp.Schema.Description, resp.Schema.MarkdownDescription
}

func extractResourceSchemaMethod(r resource.Resource) (reflect.Type, reflect.Method, bool) {
	resourceType := reflect.TypeOf(r)
	method, found := resourceType.MethodByName("Schema")

	return resourceType, method, found
}

func TestProviderResourceDescriptionsAreNotEmpty(t *testing.T) {
	p := provider.NewFrameworkProvider(nil)()
	for _, r := range p.Resources(t.Context()) {
		description, markdownDescription := extractResourceDescriptions(t.Context(), r())
		resourceType, method, found := extractResourceSchemaMethod(r())

		if found {
			funcPtr := method.Func.Pointer()
			fn := runtime.FuncForPC(funcPtr)
			file, line := fn.FileLine(funcPtr)
			assert.NotEmpty(t, description, "Please fill up Description field in %s schema, %s:%d", resourceType, file, line)
			assert.NotEmpty(t, markdownDescription, "Please fill up MarkdownDescription field in %s schema, %s:%d", resourceType, file, line)
		} else {
			t.Errorf("No Schema function found for the resource %s", resourceType)
		}
	}
}

func extractDataSourceDescriptions(ctx context.Context, d datasource.DataSource) (string, string) {
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)

	return resp.Schema.Description, resp.Schema.MarkdownDescription
}

func extractDataSourceSchemaMethod(d datasource.DataSource) (reflect.Type, reflect.Method, bool) {
	dataSourceType := reflect.TypeOf(d)
	method, found := dataSourceType.MethodByName("Schema")

	return dataSourceType, method, found
}

func TestProviderDataSourceDescriptionsAreNotEmpty(t *testing.T) {
	p := provider.NewFrameworkProvider(nil)()
	for _, d := range p.DataSources(t.Context()) {
		description, markdownDescription := extractDataSourceDescriptions(t.Context(), d())
		dataSourceType, method, found := extractDataSourceSchemaMethod(d())

		if found {
			funcPtr := method.Func.Pointer()
			fn := runtime.FuncForPC(funcPtr)
			file, line := fn.FileLine(funcPtr)
			assert.NotEmpty(t, description, "Please fill up Description field in %s schema, %s:%d", dataSourceType, file, line)
			assert.NotEmpty(t, markdownDescription, "Please fill up MarkdownDescription field in %s schema, %s:%d", dataSourceType, file, line)
		} else {
			t.Errorf("No Schema function found for the data source %s", dataSourceType)
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

func validateDefaultAnnotationsConfig(t *testing.T, values map[string]tftypes.Value) []string {
	t.Helper()

	ctx := t.Context()

	p := provider.NewFrameworkProvider(nil)().(providerFramework.ProviderWithValidateConfig)

	var schemaResp providerFramework.SchemaResponse

	p.Schema(ctx, providerFramework.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	terraformType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)

	attrs := make(map[string]tftypes.Value, len(terraformType.AttributeTypes))

	for name, attrType := range terraformType.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}

	if values != nil {
		attrs["default_annotations"] = tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, values)
	}

	var resp providerFramework.ValidateConfigResponse

	p.ValidateConfig(ctx, providerFramework.ValidateConfigRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(terraformType, attrs),
			Schema: schemaResp.Schema,
		},
	}, &resp)

	summary := make([]string, 0, len(resp.Diagnostics))

	for _, diagnostic := range resp.Diagnostics {
		summary = append(summary, diagnostic.Summary())
	}

	return summary
}

func TestProviderValidateDefaultAnnotationsConfig(t *testing.T) {
	t.Parallel()

	t.Run("null default_annotations is valid", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, validateDefaultAnnotationsConfig(t, nil))
	})

	t.Run("valid default_annotations", func(t *testing.T) {
		t.Parallel()

		summary := validateDefaultAnnotationsConfig(t, map[string]tftypes.Value{
			"team": tftypes.NewValue(tftypes.String, "platform"),
			"env":  tftypes.NewValue(tftypes.String, "production"),
		})

		assert.Empty(t, summary)
	})

	t.Run("invalid key names are reported", func(t *testing.T) {
		t.Parallel()

		summary := validateDefaultAnnotationsConfig(t, map[string]tftypes.Value{
			"2bad":  tftypes.NewValue(tftypes.String, "value"),
			"team!": tftypes.NewValue(tftypes.String, "value"),
		})

		assert.ElementsMatch(t, []string{
			"Invalid default annotation key",
			"Invalid default annotation key",
		}, summary)
	})

	t.Run("invalid value names are reported", func(t *testing.T) {
		t.Parallel()

		summary := validateDefaultAnnotationsConfig(t, map[string]tftypes.Value{
			"team": tftypes.NewValue(tftypes.String, ""),
			"env":  tftypes.NewValue(tftypes.String, "2value"),
		})

		assert.ElementsMatch(t, []string{
			"Invalid default annotation value",
			"Invalid default annotation value",
		}, summary)
	})
}
