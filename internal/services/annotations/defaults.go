package annotations

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	annotationsSDK "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// wrappedSDKv2Resources tracks the *schema.Resource pointers already wrapped
// by WithDefaultBindings, so that wrapping the same pointer twice does not
// stack a second wrapper layer and double every binding operation.
var (
	wrappedSDKv2ResourcesMu sync.Mutex
	wrappedSDKv2Resources   = make(map[*schema.Resource]bool)
)

// defaultAnnotationsMarker is set as the description of annotation keys and
// values created by the default annotations feature. The cleanup of orphaned
// keys and values (see CleanupOrphanedDefaultAnnotations) only touches items
// carrying this exact description, so that keys and values created or edited
// out-of-band are never touched.
const defaultAnnotationsMarker = "Created by the Scaleway Terraform provider (default_annotations)"

// ResolveDefaults creates or finds annotation keys and values for the given
// map of key→value name pairs, returning a resolved AnnotationDefaultsConfig
// with their API IDs. Resolution happens during provider Configure, at most
// once per Meta.
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
			KeyID:       existingKey.ID,
			Name:        valueName,
			Description: defaultAnnotationsMarker,
		}, scw.WithContext(ctx))
		if err != nil {
			return "", "", fmt.Errorf("creating annotation value %q for key %q: %w", valueName, keyName, err)
		}

		return existingKey.ID, value.ID, nil
	}

	key, err := api.CreateKey(&annotationsSDK.CreateKeyRequest{
		OrganizationID: orgID,
		Name:           keyName,
		Description:    defaultAnnotationsMarker,
	}, scw.WithContext(ctx))
	if err != nil {
		return "", "", fmt.Errorf("creating annotation key %q: %w", keyName, err)
	}

	value, err := api.CreateValue(&annotationsSDK.CreateValueRequest{
		KeyID:       key.ID,
		Name:        valueName,
		Description: defaultAnnotationsMarker,
	}, scw.WithContext(ctx))
	if err != nil {
		return "", "", fmt.Errorf("creating annotation value %q for key %q: %w", valueName, keyName, err)
	}

	return key.ID, value.ID, nil
}

// ResolveDefaultsOnMeta reads the raw default annotations from the Meta,
// resolves them to API IDs, and stores the resolved config back on the Meta.
// Resolution runs at most once per Meta and is safe for concurrent calls.
//
// Resolution happens during provider Configure, so when default_annotations
// is set, every provider operation depends on the annotations API: it
// requires the annotations:read scope (and annotations:write when new keys
// or values must be created).
func ResolveDefaultsOnMeta(ctx context.Context, m *meta.Meta) error {
	annotations := m.DefaultAnnotations()
	if len(annotations) == 0 {
		return nil
	}

	return m.EnsureDefaultAnnotationsConfig(func() (*meta.AnnotationDefaultsConfig, error) {
		return ResolveDefaults(ctx, m.ScwClient(), annotations)
	})
}

// cleanupAnnotationsAPI is the subset of the annotations API used by the
// cleanup of orphaned keys and values, extracted as an interface for unit
// testing.
type cleanupAnnotationsAPI interface {
	ListAllKeysAndValues(request *annotationsSDK.ListAllKeysAndValuesRequest, options ...scw.RequestOption) (*annotationsSDK.ListAllKeysAndValuesResponse, error)
	ListValues(request *annotationsSDK.ListValuesRequest, options ...scw.RequestOption) (*annotationsSDK.ListValuesResponse, error)
	ListBindings(request *annotationsSDK.ListBindingsRequest, options ...scw.RequestOption) (*annotationsSDK.ListBindingsResponse, error)
	DeleteValue(request *annotationsSDK.DeleteValueRequest, options ...scw.RequestOption) error
	DeleteKey(request *annotationsSDK.DeleteKeyRequest, options ...scw.RequestOption) error
}

// CleanupOrphanedDefaultAnnotations removes annotation keys and values that
// this provider created for a previous default annotations configuration and
// that are no longer referenced by the current one (given as key→value name
// pairs, possibly nil or empty):
//
//   - a value is deleted when it carries the defaultAnnotationsMarker, is not
//     the configured value of its key, and has no remaining bindings,
//   - a key is deleted when it carries the marker, is not part of the
//     current configuration, and has no remaining values.
//
// Items that do not carry the marker, or that still have bindings or
// values, are left untouched: the API refuses to delete them, and they are
// cleaned up on a later run.
func CleanupOrphanedDefaultAnnotations(ctx context.Context, client *scw.Client, current map[string]string) error {
	if client == nil {
		return errors.New("no Scaleway client found, cannot clean up default annotations")
	}

	orgID, exists := client.GetDefaultOrganizationID()
	if !exists {
		return errors.New("no default organization ID found, cannot clean up default annotations")
	}

	return cleanupOrphanedDefaultAnnotations(ctx, annotationsSDK.NewAPI(client), orgID, current)
}

func cleanupOrphanedDefaultAnnotations(ctx context.Context, api cleanupAnnotationsAPI, orgID string, current map[string]string) error {
	all, err := api.ListAllKeysAndValues(&annotationsSDK.ListAllKeysAndValuesRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("listing annotation keys and values: %w", err)
	}

	for _, key := range all.Keys {
		if key.Description != defaultAnnotationsMarker {
			continue
		}

		configuredValue, isConfigured := current[key.Name]

		for _, value := range key.Values {
			if value.Description != defaultAnnotationsMarker {
				continue
			}

			if isConfigured && value.Name == configuredValue {
				continue
			}

			if err := cleanupOrphanedValue(ctx, api, orgID, value.ID); err != nil {
				return err
			}
		}

		if isConfigured {
			continue
		}

		// The key is no longer configured: delete it unless a value still
		// remains under it (for instance a value created out-of-band after
		// the key was marked).
		remaining, err := api.ListValues(&annotationsSDK.ListValuesRequest{
			KeyID:          &key.ID,
			OrganizationID: orgID,
		}, scw.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("listing values of annotation key %q: %w", key.Name, err)
		}

		if len(remaining.Values) > 0 {
			continue
		}

		err = api.DeleteKey(&annotationsSDK.DeleteKeyRequest{
			KeyID: key.ID,
		}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is404(err) {
			return fmt.Errorf("deleting orphaned annotation key %q: %w", key.Name, err)
		}
	}

	return nil
}

// cleanupOrphanedValue deletes the value with the given ID when it has no
// remaining bindings. Bound values are left in place: the API refuses to
// delete them, and they are cleaned up on a later run.
func cleanupOrphanedValue(ctx context.Context, api cleanupAnnotationsAPI, orgID, valueID string) error {
	valueIDPtr := valueID

	bindings, err := listAllBindings(ctx, api, &annotationsSDK.ListBindingsRequest{
		ValueID:        &valueIDPtr,
		OrganizationID: orgID,
	})
	if err != nil {
		return fmt.Errorf("listing bindings of annotation value %q: %w", valueID, err)
	}

	if len(bindings) > 0 {
		return nil
	}

	err = api.DeleteValue(&annotationsSDK.DeleteValueRequest{
		ValueID: valueID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return fmt.Errorf("deleting orphaned annotation value %q: %w", valueID, err)
	}

	return nil
}

// CleanupOrphanedDefaultsOnMeta runs the best-effort cleanup of orphaned
// default annotation keys and values on the Meta, at most once per Meta and
// safe for concurrent calls.
//
// It is skipped when default_annotations has never been set, so that
// providers that do not use the feature make no annotations API call. To
// clean up the keys and values left by a previous configuration, first
// apply with default_annotations set to an empty map, then remove it from
// the configuration.
//
// This function is called from the resource wrappers (Create, Update,
// Delete) rather than from Configure, so that no deletes happen during
// plan. The cleanup runs at most once per Meta, so calling it from every
// CRUD method is safe.
func CleanupOrphanedDefaultsOnMeta(ctx context.Context, m *meta.Meta) error {
	if m.DefaultAnnotations() == nil {
		return nil
	}

	return m.CleanupOrphanedDefaultAnnotations(func() error {
		return CleanupOrphanedDefaultAnnotations(ctx, m.ScwClient(), m.DefaultAnnotations())
	})
}

// bindingsLister is the subset of the annotations API that lists bindings.
// Both *annotationsSDK.API and cleanupAnnotationsAPI satisfy it.
type bindingsLister interface {
	ListBindings(request *annotationsSDK.ListBindingsRequest, options ...scw.RequestOption) (*annotationsSDK.ListBindingsResponse, error)
}

// bindingAPI is the subset of the annotations API used by binding operations
// (create, reconcile, delete), extracted as an interface for unit testing.
type bindingAPI interface {
	bindingsLister
	CreateBinding(request *annotationsSDK.CreateBindingRequest, options ...scw.RequestOption) (*annotationsSDK.Binding, error)
	DeleteBinding(request *annotationsSDK.DeleteBindingRequest, options ...scw.RequestOption) error
}

// listAllBindings lists all bindings matching the request, handling
// pagination. The request's Page and PageSize fields are overwritten.
func listAllBindings(ctx context.Context, api bindingsLister, req *annotationsSDK.ListBindingsRequest) ([]*annotationsSDK.Binding, error) {
	pageSize := uint32(100)

	var allBindings []*annotationsSDK.Binding

	for page := int32(1); ; page++ {
		req.Page = &page
		req.PageSize = &pageSize

		resp, err := api.ListBindings(req, scw.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		allBindings = append(allBindings, resp.Bindings...)

		if uint64(len(allBindings)) >= resp.TotalCount {
			break
		}
	}

	return allBindings, nil
}

// CreateDefaultBindings creates annotation bindings for all default
// annotations on the resource identified by srn.
//
// Creation is idempotent: bindings for values already bound on the target
// (created out-of-band or by a previous partial run) are skipped, since the
// API fails when the target already has a value for a key.
func CreateDefaultBindings(ctx context.Context, client *scw.Client, srn string, dc *meta.AnnotationDefaultsConfig) error {
	if dc == nil || len(dc.Items) == 0 || srn == "" {
		return nil
	}

	return createDefaultBindings(ctx, annotationsSDK.NewAPI(client), srn, dc)
}

func createDefaultBindings(ctx context.Context, api bindingAPI, srn string, dc *meta.AnnotationDefaultsConfig) error {
	existingValueIDs, err := listBindingValueIDs(ctx, api, srn)
	if err != nil {
		return err
	}

	for _, item := range dc.Items {
		if existingValueIDs[item.ValueID] {
			continue
		}

		_, err := api.CreateBinding(&annotationsSDK.CreateBindingRequest{
			TargetSrn: srn,
			ValueID:   item.ValueID,
		}, scw.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("creating binding for annotation %q=%q on %q: %w", item.KeyName, item.ValueName, srn, err)
		}
	}

	return nil
}

// listBindingValueIDs returns the set of value IDs currently bound on the
// resource identified by srn.
func listBindingValueIDs(ctx context.Context, api bindingsLister, srn string) (map[string]bool, error) {
	existingValueIDs := make(map[string]bool)

	bindings, err := listAllBindings(ctx, api, &annotationsSDK.ListBindingsRequest{
		TargetSrn: &srn,
	})
	if err != nil {
		return nil, fmt.Errorf("listing bindings for %q: %w", srn, err)
	}

	for _, binding := range bindings {
		if binding.Value == nil {
			continue
		}

		existingValueIDs[binding.Value.ID] = true
	}

	return existingValueIDs, nil
}

// ReconcileDefaultBindings ensures the resource identified by srn carries the
// provider's default annotation bindings:
//   - bindings for the configured key→value pairs that are missing are
//     created (covers imported resources, out-of-band deletion, and keys
//     added to default_annotations after the resource was created),
//   - bindings under a default key holding a different value — left over from
//     a previous default_annotations configuration — are deleted.
//
// Bindings under keys that are not part of the current default configuration
// are left untouched: once a key is removed from default_annotations, its
// stale bindings can no longer be distinguished from user-managed ones.
//
// Reconciliation involves writes (CreateBinding/DeleteBinding) and is only
// run on Create and Update, never on Read, so that no writes happen during
// plan or refresh.
func ReconcileDefaultBindings(ctx context.Context, client *scw.Client, srn string, dc *meta.AnnotationDefaultsConfig) error {
	if dc == nil || len(dc.Items) == 0 || srn == "" {
		return nil
	}

	return reconcileDefaultBindings(ctx, annotationsSDK.NewAPI(client), srn, dc)
}

func reconcileDefaultBindings(ctx context.Context, api bindingAPI, srn string, dc *meta.AnnotationDefaultsConfig) error {
	bindings, err := listAllBindings(ctx, api, &annotationsSDK.ListBindingsRequest{
		TargetSrn: &srn,
	})
	if err != nil {
		return fmt.Errorf("listing bindings for %q: %w", srn, err)
	}

	expectedValueIDByKeyID := make(map[string]string, len(dc.Items))
	for _, item := range dc.Items {
		expectedValueIDByKeyID[item.KeyID] = item.ValueID
	}

	for _, binding := range bindings {
		if binding.Key == nil || binding.Value == nil {
			continue
		}

		expectedValueID, isDefaultKey := expectedValueIDByKeyID[binding.Key.ID]
		if !isDefaultKey || binding.Value.ID == expectedValueID {
			continue
		}

		err := api.DeleteBinding(&annotationsSDK.DeleteBindingRequest{
			BindingID: binding.ID,
		}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is404(err) {
			return fmt.Errorf("deleting stale binding %q on %q: %w", binding.ID, srn, err)
		}
	}

	return createDefaultBindings(ctx, api, srn, dc)
}

// DeleteDefaultBindings deletes annotation bindings on the resource
// identified by srn that were created by the default annotations feature:
// only bindings whose key matches one of the configured default annotation
// keys are deleted. Bindings under other keys — including user-managed ones
// (e.g. from an explicit scaleway_annotations_binding resource) and stale
// bindings from keys removed from default_annotations — are left untouched,
// consistent with ReconcileDefaultBindings.
//
// Bindings that no longer exist (404) are silently ignored.
func DeleteDefaultBindings(ctx context.Context, client *scw.Client, srn string, dc *meta.AnnotationDefaultsConfig) error {
	if dc == nil || len(dc.Items) == 0 || srn == "" {
		return nil
	}

	return deleteDefaultBindings(ctx, annotationsSDK.NewAPI(client), srn, dc)
}

func deleteDefaultBindings(ctx context.Context, api bindingAPI, srn string, dc *meta.AnnotationDefaultsConfig) error {
	defaultKeyIDs := make(map[string]bool, len(dc.Items))
	for _, item := range dc.Items {
		defaultKeyIDs[item.KeyID] = true
	}

	bindings, err := listAllBindings(ctx, api, &annotationsSDK.ListBindingsRequest{
		TargetSrn: &srn,
	})
	if err != nil {
		return fmt.Errorf("listing bindings for %q: %w", srn, err)
	}

	for _, binding := range bindings {
		if binding.Key == nil || !defaultKeyIDs[binding.Key.ID] {
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

// ReconcileDefaultBindingsFromMeta is a convenience wrapper that extracts the
// resolved default annotations config from the meta value and reconciles the
// bindings of the resource identified by srn. Does nothing if no default
// annotations are configured.
func ReconcileDefaultBindingsFromMeta(ctx context.Context, m any, srn string) error {
	dc := meta.ExtractDefaultAnnotationsConfig(m)
	if dc == nil {
		return nil
	}

	return ReconcileDefaultBindings(ctx, meta.ExtractScwClient(m), srn, dc)
}

// DeleteDefaultBindingsFromMeta is a convenience wrapper that extracts the
// resolved default annotations config from the meta value and deletes the
// default annotation bindings of the resource identified by srn. Only
// bindings under configured default keys are deleted; user-managed bindings
// are left untouched. Does nothing if no default annotations are configured.
func DeleteDefaultBindingsFromMeta(ctx context.Context, m any, srn string) error {
	dc := meta.ExtractDefaultAnnotationsConfig(m)
	if dc == nil {
		return nil
	}

	return DeleteDefaultBindings(ctx, meta.ExtractScwClient(m), srn, dc)
}

// sdkv2ResourceSrn returns the value of the resource's top-level "srn"
// attribute, or an empty string if the resource has no such attribute or it
// is not set.
func sdkv2ResourceSrn(d *schema.ResourceData) string {
	srn, _ := d.Get("srn").(string)

	return srn
}

// WithDefaultBindings wraps an SDKv2 resource so that default annotation
// bindings are managed based on the resource's "srn" attribute:
//   - after Create and after Update, the bindings are fully reconciled
//     (missing defaults are created, stale values under default keys are
//     deleted),
//   - on Read, no bindings are created or deleted: Read must not perform
//     writes since it runs during plan and refresh,
//   - before Delete, default bindings of the resource are deleted.
//   - on Create, Update and Delete, the best-effort cleanup of orphaned
//     default annotation keys and values is run (deferred from Configure
//     so that no deletes happen during plan).
//
// Only resources with a Computed top-level "srn" attribute are wrapped: the
// SRN of a managed resource is always derived from the API, never set by
// the user. Resources without a Computed "srn" field, or with an empty SRN
// at runtime, are no-ops.
//
// Wrapping the same *schema.Resource pointer twice is a no-op (see
// wrappedSDKv2Resources).
func WithDefaultBindings(r *schema.Resource) *schema.Resource {
	// SchemaMap resolves a lazily-defined schema (SchemaFunc): resources that
	// use SchemaFunc have an empty Schema field.
	srnSchema, ok := r.SchemaMap()["srn"]
	if !ok || !srnSchema.Computed {
		return r
	}

	wrappedSDKv2ResourcesMu.Lock()
	if wrappedSDKv2Resources[r] {
		wrappedSDKv2ResourcesMu.Unlock()

		return r
	}

	wrappedSDKv2Resources[r] = true
	wrappedSDKv2ResourcesMu.Unlock()

	reconcile := func(ctx context.Context, d *schema.ResourceData, m any) error {
		if srn := sdkv2ResourceSrn(d); srn != "" {
			return ReconcileDefaultBindingsFromMeta(ctx, m, srn)
		}

		return nil
	}

	runCleanup := func(ctx context.Context, m any) diag.Diagnostics {
		metaVal, ok := m.(*meta.Meta)
		if !ok || metaVal == nil {
			return nil
		}

		if err := CleanupOrphanedDefaultsOnMeta(ctx, metaVal); err != nil {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Failed to clean up orphaned default annotation keys and values",
					Detail:   err.Error(),
				},
			}
		}

		return nil
	}

	if r.CreateContext != nil {
		originalCreate := r.CreateContext

		r.CreateContext = func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			diags := runCleanup(ctx, m)
			if diags.HasError() {
				return diags
			}

			createDiags := originalCreate(ctx, d, m)
			diags = append(diags, createDiags...)

			if diags.HasError() {
				return diags
			}

			if err := reconcile(ctx, d, m); err != nil {
				return diag.FromErr(err)
			}

			return diags
		}
	}

	if r.ReadContext != nil {
		originalRead := r.ReadContext

		r.ReadContext = func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			diags := originalRead(ctx, d, m)
			if diags.HasError() {
				return diags
			}

			// The resource no longer exists: state has been cleared.
			if d.Id() == "" {
				return diags
			}

			// Read is called during plan (refresh) and must not perform
			// writes. Binding reconciliation — including creation of
			// missing default bindings — is done on Create and Update.
			return diags
		}
	}

	if r.UpdateContext != nil {
		originalUpdate := r.UpdateContext

		r.UpdateContext = func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			diags := runCleanup(ctx, m)
			if diags.HasError() {
				return diags
			}

			updateDiags := originalUpdate(ctx, d, m)
			diags = append(diags, updateDiags...)

			if diags.HasError() {
				return diags
			}

			if err := reconcile(ctx, d, m); err != nil {
				return diag.FromErr(err)
			}

			return diags
		}
	}

	if r.DeleteContext != nil {
		originalDelete := r.DeleteContext

		r.DeleteContext = func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			// If deleting the bindings fails, abort the destroy before the
			// inner Delete runs, so the resource stays in state and the
			// destroy can be retried instead of annotations errors being
			// silently ignored.
			if srn := sdkv2ResourceSrn(d); srn != "" {
				if err := DeleteDefaultBindingsFromMeta(ctx, m, srn); err != nil {
					return diag.FromErr(err)
				}
			}

			cleanupDiags := runCleanup(ctx, m)

			innerDiags := originalDelete(ctx, d, m)

			return append(cleanupDiags, innerDiags...)
		}
	}

	return r
}
