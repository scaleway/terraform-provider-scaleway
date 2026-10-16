package annotations

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	annotationsSDK "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// ResolveDefaults creates or finds annotation keys and values for the given
// map of key→value name pairs, returning a resolved AnnotationDefaultsConfig
// with their API IDs. This is called during provider Configure so that the
// IDs are cached and available when resources create bindings.
//
// If a key with the given name already exists in the organization, it is
// reused. Otherwise, a new key is created. The same applies to values under
// each key.
func ResolveDefaults(ctx context.Context, client *scw.Client, annotations map[string]string) (*meta.AnnotationDefaultsConfig, error) {
	if len(annotations) == 0 {
		return nil, nil
	}

	api := annotationsSDK.NewAPI(client)

	orgID, exists := client.GetDefaultOrganizationID()
	if !exists {
		return nil, errors.New("no default organization ID found, cannot resolve default annotations")
	}

	allKeysAndValues, err := api.ListAllKeysAndValues(&annotationsSDK.ListAllKeysAndValuesRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("listing annotation keys and values: %w", err)
	}

	existingKeys := make(map[string]*annotationsSDK.ListAllKeysAndValuesResponseKey)
	for _, key := range allKeysAndValues.Keys {
		existingKeys[key.Name] = key
	}

	dc := &meta.AnnotationDefaultsConfig{}

	keys := make([]string, 0, len(annotations))
	for k := range annotations {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, keyName := range keys {
		valueName := annotations[keyName]

		keyID, valueID, err := resolveKeyValue(ctx, api, orgID, existingKeys, keyName, valueName)
		if err != nil {
			return nil, err
		}

		dc.Items = append(dc.Items, meta.AnnotationDefault{
			KeyName:   keyName,
			KeyID:     keyID,
			ValueName: valueName,
			ValueID:   valueID,
		})
	}

	return dc, nil
}

func resolveKeyValue(
	ctx context.Context,
	api *annotationsSDK.API,
	orgID string,
	existingKeys map[string]*annotationsSDK.ListAllKeysAndValuesResponseKey,
	keyName, valueName string,
) (string, string, error) {
	if existingKey, ok := existingKeys[keyName]; ok {
		for _, value := range existingKey.Values {
			if value.Name == valueName {
				return existingKey.ID, value.ID, nil
			}
		}

		value, err := api.CreateValue(&annotationsSDK.CreateValueRequest{
			KeyID: existingKey.ID,
			Name:  valueName,
		}, scw.WithContext(ctx))
		if err != nil {
			return "", "", fmt.Errorf("creating annotation value %q for key %q: %w", valueName, keyName, err)
		}

		return existingKey.ID, value.ID, nil
	}

	key, err := api.CreateKey(&annotationsSDK.CreateKeyRequest{
		OrganizationID: orgID,
		Name:           keyName,
	}, scw.WithContext(ctx))
	if err != nil {
		return "", "", fmt.Errorf("creating annotation key %q: %w", keyName, err)
	}

	value, err := api.CreateValue(&annotationsSDK.CreateValueRequest{
		KeyID: key.ID,
		Name:  valueName,
	}, scw.WithContext(ctx))
	if err != nil {
		return "", "", fmt.Errorf("creating annotation value %q for key %q: %w", valueName, keyName, err)
	}

	return key.ID, value.ID, nil
}

// ResolveDefaultsOnMeta reads the raw default annotations from the Meta,
// resolves them to API IDs, and stores the resolved config back on the Meta.
// This is called during provider Configure.
func ResolveDefaultsOnMeta(ctx context.Context, m *meta.Meta) error {
	annotations := m.DefaultAnnotations()
	if len(annotations) == 0 {
		return nil
	}

	if m.DefaultAnnotationsConfig() != nil {
		return nil
	}

	dc, err := ResolveDefaults(ctx, m.ScwClient(), annotations)
	if err != nil {
		return err
	}

	m.SetDefaultAnnotationsConfig(dc)

	return nil
}

// CreateDefaultBindings creates annotation bindings for all default
// annotations on the resource identified by srn. It returns a map of
// key name → value name for the bindings that were created.
func CreateDefaultBindings(ctx context.Context, client *scw.Client, srn string, dc *meta.AnnotationDefaultsConfig) (map[string]string, error) {
	if dc == nil || len(dc.Items) == 0 || srn == "" {
		return nil, nil
	}

	api := annotationsSDK.NewAPI(client)
	result := make(map[string]string)

	for _, item := range dc.Items {
		_, err := api.CreateBinding(&annotationsSDK.CreateBindingRequest{
			Srn:     srn,
			ValueID: item.ValueID,
		}, scw.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("creating binding for annotation %q=%q on %q: %w", item.KeyName, item.ValueName, srn, err)
		}

		result[item.KeyName] = item.ValueName
	}

	return result, nil
}

// ReadDefaultBindings lists annotation bindings for the resource identified
// by srn and returns the subset that matches the provider's default
// annotations, as a map of key name → value name.
func ReadDefaultBindings(ctx context.Context, client *scw.Client, srn string, dc *meta.AnnotationDefaultsConfig) (map[string]string, error) {
	if dc == nil || len(dc.Items) == 0 || srn == "" {
		return nil, nil
	}

	api := annotationsSDK.NewAPI(client)

	resp, err := api.ListBindings(&annotationsSDK.ListBindingsRequest{
		Srn: &srn,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("listing bindings for %q: %w", srn, err)
	}

	defaultValueIDs := make(map[string]meta.AnnotationDefault)
	for _, item := range dc.Items {
		defaultValueIDs[item.ValueID] = item
	}

	result := make(map[string]string)

	for _, binding := range resp.Bindings {
		if item, ok := defaultValueIDs[binding.Value.ID]; ok {
			result[item.KeyName] = item.ValueName
		}
	}

	return result, nil
}

// DeleteDefaultBindings deletes all annotation bindings for the resource
// identified by srn that correspond to the provider's default annotations.
// Bindings that no longer exist (404) are silently ignored.
func DeleteDefaultBindings(ctx context.Context, client *scw.Client, srn string, dc *meta.AnnotationDefaultsConfig) error {
	if dc == nil || len(dc.Items) == 0 || srn == "" {
		return nil
	}

	api := annotationsSDK.NewAPI(client)

	resp, err := api.ListBindings(&annotationsSDK.ListBindingsRequest{
		Srn: &srn,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("listing bindings for %q: %w", srn, err)
	}

	defaultValueIDs := make(map[string]bool)
	for _, item := range dc.Items {
		defaultValueIDs[item.ValueID] = true
	}

	for _, binding := range resp.Bindings {
		if !defaultValueIDs[binding.Value.ID] {
			continue
		}

		err := api.DeleteBinding(&annotationsSDK.DeleteBindingRequest{
			BindingID: binding.ID,
		}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is404(err) {
			return fmt.Errorf("deleting binding %q: %w", binding.ID, err)
		}
	}

	return nil
}

// resolvedConfigFromMeta returns the resolved default annotations config,
// resolving it on demand if raw annotations are configured but not yet
// resolved. This handles the case where meta was injected directly (e.g.
// in acceptance tests) and provider Configure did not run the resolution.
func resolvedConfigFromMeta(ctx context.Context, m any) (*meta.AnnotationDefaultsConfig, error) {
	dc := meta.ExtractDefaultAnnotationsConfig(m)
	if dc != nil {
		return dc, nil
	}

	if metaVal, ok := m.(*meta.Meta); ok && len(metaVal.DefaultAnnotations()) > 0 {
		if err := ResolveDefaultsOnMeta(ctx, metaVal); err != nil {
			return nil, err
		}

		dc = meta.ExtractDefaultAnnotationsConfig(m)
	}

	return dc, nil
}

// CreateDefaultBindingsFromMeta is a convenience wrapper that extracts the
// resolved default annotations config from the meta value and creates
// bindings. Returns nil if no default annotations are configured.
func CreateDefaultBindingsFromMeta(ctx context.Context, m any, srn string) (map[string]string, error) {
	dc, err := resolvedConfigFromMeta(ctx, m)
	if err != nil {
		return nil, err
	}

	if dc == nil {
		return nil, nil
	}

	return CreateDefaultBindings(ctx, meta.ExtractScwClient(m), srn, dc)
}

// ReadDefaultBindingsFromMeta is a convenience wrapper that extracts the
// resolved default annotations config from the meta value and reads
// bindings. Returns nil if no default annotations are configured.
func ReadDefaultBindingsFromMeta(ctx context.Context, m any, srn string) (map[string]string, error) {
	dc, err := resolvedConfigFromMeta(ctx, m)
	if err != nil {
		return nil, err
	}

	if dc == nil {
		return nil, nil
	}

	return ReadDefaultBindings(ctx, meta.ExtractScwClient(m), srn, dc)
}

// DeleteDefaultBindingsFromMeta is a convenience wrapper that extracts the
// resolved default annotations config from the meta value and deletes
// bindings. Does nothing if no default annotations are configured.
func DeleteDefaultBindingsFromMeta(ctx context.Context, m any, srn string) error {
	dc, err := resolvedConfigFromMeta(ctx, m)
	if err != nil {
		return err
	}

	if dc == nil {
		return nil
	}

	return DeleteDefaultBindings(ctx, meta.ExtractScwClient(m), srn, dc)
}

// WithDefaultBindings wraps an SDKv2 resource so that default annotation
// bindings are created after Create and deleted before Delete, based on the
// resource's "srn" attribute. Resources without an "srn" schema field or with
// an empty SRN are no-ops.
//
// This lets the provider manage default annotation bindings for all resources
// transparently, without each resource implementation needing to know about
// annotations.
func WithDefaultBindings(r *schema.Resource) *schema.Resource {
	if r.CreateContext != nil {
		originalCreate := r.CreateContext

		r.CreateContext = func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			diags := originalCreate(ctx, d, m)
			if diags.HasError() {
				return diags
			}

			if srn, ok := d.Get("srn").(string); ok && srn != "" {
				if _, err := CreateDefaultBindingsFromMeta(ctx, m, srn); err != nil {
					return diag.FromErr(err)
				}
			}

			return diags
		}
	}

	if r.DeleteContext != nil {
		originalDelete := r.DeleteContext

		r.DeleteContext = func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			if srn, ok := d.Get("srn").(string); ok && srn != "" {
				if err := DeleteDefaultBindingsFromMeta(ctx, m, srn); err != nil {
					return diag.FromErr(err)
				}
			}

			return originalDelete(ctx, d, m)
		}
	}

	return r
}
