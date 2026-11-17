package annotations

import (
	"context"
	"errors"
	"fmt"
	"sort"

	annotationsSDK "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

type DefaultAnnotation struct {
	KeyName   string
	KeyID     string
	ValueName string
	ValueID   string
}

type ResolutionError struct {
	Err error
}

func (e *ResolutionError) Error() string {
	return "resolving default annotations: " + e.Err.Error()
}

func (e *ResolutionError) Unwrap() error {
	return e.Err
}

func IsResolutionError(err error) bool {
	var resolutionErr *ResolutionError

	return errors.As(err, &resolutionErr)
}

type keysAndValuesLister interface {
	ListAllKeysAndValues(request *annotationsSDK.ListAllKeysAndValuesRequest, options ...scw.RequestOption) (*annotationsSDK.ListAllKeysAndValuesResponse, error)
}

type bindingManager interface {
	ListBindings(request *annotationsSDK.ListBindingsRequest, options ...scw.RequestOption) (*annotationsSDK.ListBindingsResponse, error)
	CreateBinding(request *annotationsSDK.CreateBindingRequest, options ...scw.RequestOption) (*annotationsSDK.Binding, error)
	DeleteBinding(request *annotationsSDK.DeleteBindingRequest, options ...scw.RequestOption) error
}

// resolveWishlist resolves the default annotations of the provider
// configuration (key name/value name) to their API IDs by listing the keys
// and values of the organization. Keys and values must exist beforehand.
// TODO: see if key/value creation/deletion can be handled properly from
// the defaults as well.
func resolveWishlist(ctx context.Context, api keysAndValuesLister, orgID string, wishlist map[string]string) ([]DefaultAnnotation, error) {
	if len(wishlist) == 0 {
		return nil, nil
	}

	all, err := api.ListAllKeysAndValues(&annotationsSDK.ListAllKeysAndValuesRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("listing annotation keys and values: %w", err)
	}

	existingKeys := make(map[string]*annotationsSDK.ListAllKeysAndValuesResponseKey, len(all.Keys))
	for _, key := range all.Keys {
		existingKeys[key.Name] = key
	}

	names := make([]string, 0, len(wishlist))
	for name := range wishlist {
		names = append(names, name)
	}

	sort.Strings(names)

	resolved := make([]DefaultAnnotation, 0, len(names))

	for _, keyName := range names {
		valueName := wishlist[keyName]

		key := existingKeys[keyName]
		if key == nil {
			return nil, fmt.Errorf("annotation key %q not found, create it beforehand (for example with the scaleway_annotations_key resource)", keyName)
		}

		valueID, found := findValueID(key, valueName)
		if !found {
			return nil, fmt.Errorf("annotation value %q for key %q not found, create it beforehand (for example with the scaleway_annotations_value resource)", valueName, keyName)
		}

		resolved = append(resolved, DefaultAnnotation{
			KeyName:   keyName,
			KeyID:     key.ID,
			ValueName: valueName,
			ValueID:   valueID,
		})
	}

	return resolved, nil
}

func findValueID(key *annotationsSDK.ListAllKeysAndValuesResponseKey, valueName string) (string, bool) {
	for _, value := range key.Values {
		if value.Name == valueName {
			return value.ID, true
		}
	}

	return "", false
}

func listAllBindings(ctx context.Context, api bindingManager, req *annotationsSDK.ListBindingsRequest) ([]*annotationsSDK.Binding, error) {
	resp, err := api.ListBindings(req, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return resp.Bindings, nil
}

// reconcileWishlist ensures the resource identified by srn carries a binding
// for each resolved default annotation: missing bindings are created and
// bindings under a default key holding a different value are deleted and
// replaced. Bindings under keys that are not part of the wishlist are left
// untouched.
func reconcileWishlist(ctx context.Context, api bindingManager, srn string, wishlist []DefaultAnnotation) error {
	bindings, err := listAllBindings(ctx, api, &annotationsSDK.ListBindingsRequest{TargetSrn: &srn})
	if err != nil {
		return fmt.Errorf("listing bindings for %q: %w", srn, err)
	}

	boundValueIDByKeyID, err := deleteStaleBindings(ctx, api, srn, bindings, wishlist)
	if err != nil {
		return err
	}

	return createMissingBindings(ctx, api, srn, boundValueIDByKeyID, wishlist)
}

func deleteStaleBindings(ctx context.Context, api bindingManager, srn string, bindings []*annotationsSDK.Binding, wishlist []DefaultAnnotation) (map[string]string, error) {
	expectedByKeyID := make(map[string]DefaultAnnotation, len(wishlist))
	for _, item := range wishlist {
		expectedByKeyID[item.KeyID] = item
	}

	boundValueIDByKeyID := make(map[string]string, len(bindings))
	for _, binding := range bindings {
		if binding.Key == nil || binding.Value == nil {
			continue
		}

		boundValueIDByKeyID[binding.Key.ID] = binding.Value.ID

		expected, isDefaultKey := expectedByKeyID[binding.Key.ID]
		if !isDefaultKey || binding.Value.ID == expected.ValueID {
			continue
		}

		if err := api.DeleteBinding(&annotationsSDK.DeleteBindingRequest{BindingID: binding.ID}, scw.WithContext(ctx)); err != nil && !httperrors.Is404(err) {
			return nil, fmt.Errorf("deleting stale binding %q on %q: %w", binding.ID, srn, err)
		}

		delete(boundValueIDByKeyID, binding.Key.ID)
	}

	return boundValueIDByKeyID, nil
}

func createMissingBindings(ctx context.Context, api bindingManager, srn string, boundValueIDByKeyID map[string]string, wishlist []DefaultAnnotation) error {
	for _, item := range wishlist {
		if boundValueIDByKeyID[item.KeyID] == item.ValueID {
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

// deleteWishlistBindings deletes the bindings on the resource identified by
// srn whose key name is part of the wishlist. Bindings under other keys are
// left untouched.
func deleteWishlistBindings(ctx context.Context, api bindingManager, srn string, wishlist map[string]string) error {
	bindings, err := listAllBindings(ctx, api, &annotationsSDK.ListBindingsRequest{TargetSrn: &srn})
	if err != nil {
		return fmt.Errorf("listing bindings for %q: %w", srn, err)
	}

	for _, binding := range bindings {
		if binding.Key == nil {
			continue
		}

		if _, isDefaultKey := wishlist[binding.Key.Name]; !isDefaultKey {
			continue
		}

		err := api.DeleteBinding(&annotationsSDK.DeleteBindingRequest{BindingID: binding.ID}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is404(err) {
			return fmt.Errorf("deleting binding %q on %q: %w", binding.ID, srn, err)
		}
	}

	return nil
}

// ReconcileDefaultBindingsForSrn resolves the default annotations from the
// provider configuration and reconciles the bindings of the resource
// identified by srn. It does nothing if no default annotations are configured.
// TODO: this function should also reconcile default_annotations with the annotations
// set in the resource config itself, with priority for the resource config annotations.
func ReconcileDefaultBindingsForSrn(ctx context.Context, m *meta.Meta, srn string) error {
	if m == nil {
		return nil
	}

	wishlist := m.DefaultAnnotations()
	if len(wishlist) == 0 || srn == "" {
		// TODO: bindings created under a previous default_annotations
		// configuration are left behind when a key is removed from, or the
		// whole default_annotations is unset in, the configuration. The
		// provider only knows the current wishlist and bindings carry no
		// provenance, so it cannot tell them apart from user-created
		// bindings. Tracking the last applied wishlist in the resource
		// state would allow deleting the bindings of the removed keys.
		return nil
	}

	client := m.ScwClient()

	orgID, exists := client.GetDefaultOrganizationID()
	if !exists {
		return &ResolutionError{Err: errors.New("no default organization ID found, cannot resolve default annotations")}
	}

	api := annotationsSDK.NewAPI(client)

	resolved, err := resolveWishlist(ctx, api, orgID, wishlist)
	if err != nil {
		return &ResolutionError{Err: err}
	}

	return reconcileWishlist(ctx, api, srn, resolved)
}

// DeleteDefaultBindingsForSrn deletes the default annotation bindings of the
// resource identified by srn. It does nothing if no default annotations are
// configured.
func DeleteDefaultBindingsForSrn(ctx context.Context, m *meta.Meta, srn string) error {
	if m == nil {
		return nil
	}

	if srn == "" || len(m.DefaultAnnotations()) == 0 {
		return nil
	}

	return deleteWishlistBindings(ctx, annotationsSDK.NewAPI(m.ScwClient()), srn, m.DefaultAnnotations())
}
