package provider_test

import (
	"context"
	"reflect"
	"runtime"
	"testing"

	actionFramework "github.com/hashicorp/terraform-plugin-framework/action"
	providerFramework "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/scaleway/terraform-provider-scaleway/v2/provider"
	"github.com/stretchr/testify/assert"
)

func extractDescriptions(ctx context.Context, a actionFramework.Action) (string, string) {
	resp := &actionFramework.SchemaResponse{}
	a.Schema(ctx, actionFramework.SchemaRequest{}, resp)

	return resp.Schema.Description, resp.Schema.MarkdownDescription
}

func TestProviderActionDescriptionsAreNotEmpty(t *testing.T) {
	p := provider.NewFrameworkProvider(nil)().(providerFramework.ProviderWithActions)
	for _, action := range p.Actions(t.Context()) {
		description, markdownDescription := extractDescriptions(t.Context(), action())
		actionType, method, found := extractSchemaMethod(action())

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

func extractSchemaMethod(action actionFramework.Action) (reflect.Type, reflect.Method, bool) {
	methodName := "Schema"
	actionType := reflect.TypeOf(action)
	method, found := actionType.MethodByName(methodName)

	return actionType, method, found
}
