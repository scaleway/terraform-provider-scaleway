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

func extractName(ctx context.Context, a actionFramework.Action) string {
	metadataResponse := &actionFramework.MetadataResponse{}
	a.Metadata(ctx, actionFramework.MetadataRequest{}, metadataResponse)

	return metadataResponse.TypeName
}

func extractDescriptions(ctx context.Context, a actionFramework.Action) (string, string) {
	resp := &actionFramework.SchemaResponse{}
	a.Schema(ctx, actionFramework.SchemaRequest{}, resp)

	return resp.Schema.Description, resp.Schema.MarkdownDescription
}

func TestProviderActionDescriptionAreNotEmpty(t *testing.T) {
	p := provider.NewFrameworkProvider(nil)().(providerFramework.ProviderWithActions)
	for _, action := range p.Actions(t.Context()) {
		// name := extractName(t.Context(), action())
		description, markdownDescription := extractDescriptions(t.Context(), action())

		methodName := "Schema"
		actionType := reflect.TypeOf(action())
		method, found := actionType.MethodByName(methodName)
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
