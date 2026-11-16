package annotations

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	annotationsSDK "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCleanupValue struct {
	id          string
	name        string
	description string
}

type fakeCleanupKey struct {
	values      map[string]*fakeCleanupValue
	id          string
	name        string
	description string
}

// fakeCleanupAPI holds the live state of an annotations organization for the
// cleanup unit tests. DeleteValue and DeleteKey mutate the state, so that
// ListValues reflects the deletions.
type fakeCleanupAPI struct {
	listAllErr        error
	keys              map[string]*fakeCleanupKey
	bindingsByValueID map[string]int
	deletedValues     []string
	deletedKeys       []string
}

var _ cleanupAnnotationsAPI = (*fakeCleanupAPI)(nil)

func newFakeCleanupAPI() *fakeCleanupAPI {
	return &fakeCleanupAPI{
		keys:              make(map[string]*fakeCleanupKey),
		bindingsByValueID: make(map[string]int),
	}
}

func (f *fakeCleanupAPI) ListAllKeysAndValues(_ *annotationsSDK.ListAllKeysAndValuesRequest, _ ...scw.RequestOption) (*annotationsSDK.ListAllKeysAndValuesResponse, error) {
	if f.listAllErr != nil {
		return nil, f.listAllErr
	}

	resp := &annotationsSDK.ListAllKeysAndValuesResponse{}

	for _, key := range f.keys {
		rk := &annotationsSDK.ListAllKeysAndValuesResponseKey{
			ID:          key.id,
			Name:        key.name,
			Description: key.description,
		}

		for _, value := range key.values {
			rk.Values = append(rk.Values, &annotationsSDK.ListAllKeysAndValuesResponseValue{
				ID:          value.id,
				Name:        value.name,
				Description: value.description,
			})
		}

		resp.Keys = append(resp.Keys, rk)
	}

	return resp, nil
}

func (f *fakeCleanupAPI) ListValues(req *annotationsSDK.ListValuesRequest, _ ...scw.RequestOption) (*annotationsSDK.ListValuesResponse, error) {
	resp := &annotationsSDK.ListValuesResponse{}

	if req.KeyID == nil {
		return resp, nil
	}

	key, ok := f.keys[*req.KeyID]
	if !ok {
		return resp, nil
	}

	for _, value := range key.values {
		resp.Values = append(resp.Values, &annotationsSDK.Value{
			ID:          value.id,
			Name:        value.name,
			Description: value.description,
		})
	}

	return resp, nil
}

func (f *fakeCleanupAPI) ListBindings(req *annotationsSDK.ListBindingsRequest, _ ...scw.RequestOption) (*annotationsSDK.ListBindingsResponse, error) {
	resp := &annotationsSDK.ListBindingsResponse{}

	if req.ValueID == nil {
		return resp, nil
	}

	for range f.bindingsByValueID[*req.ValueID] {
		resp.Bindings = append(resp.Bindings, &annotationsSDK.Binding{ID: "binding-" + *req.ValueID})
	}

	resp.TotalCount = uint64(len(resp.Bindings))

	return resp, nil
}

func (f *fakeCleanupAPI) DeleteValue(req *annotationsSDK.DeleteValueRequest, _ ...scw.RequestOption) error {
	f.deletedValues = append(f.deletedValues, req.ValueID)

	for _, key := range f.keys {
		delete(key.values, req.ValueID)
	}

	return nil
}

func (f *fakeCleanupAPI) DeleteKey(req *annotationsSDK.DeleteKeyRequest, _ ...scw.RequestOption) error {
	f.deletedKeys = append(f.deletedKeys, req.KeyID)

	delete(f.keys, req.KeyID)

	return nil
}

func (f *fakeCleanupAPI) addKey(keyID, name, description string) *fakeCleanupKey {
	key := &fakeCleanupKey{
		id:          keyID,
		name:        name,
		description: description,
		values:      make(map[string]*fakeCleanupValue),
	}

	f.keys[keyID] = key

	return key
}

func (f *fakeCleanupAPI) addValue(key *fakeCleanupKey, valueID, name, description string) *fakeCleanupValue {
	value := &fakeCleanupValue{
		id:          valueID,
		name:        name,
		description: description,
	}

	key.values[valueID] = value

	return value
}

func TestCleanupOrphanedDefaultAnnotations_DeletesOrphanedValueAndKey(t *testing.T) {
	api := newFakeCleanupAPI()

	key := api.addKey("key-1", "env", defaultAnnotationsMarker)
	api.addValue(key, "val-stale", "stale", defaultAnnotationsMarker)

	require.NoError(t, cleanupOrphanedDefaultAnnotations(t.Context(), api, "org-1", nil))

	assert.Equal(t, []string{"val-stale"}, api.deletedValues)
	assert.Equal(t, []string{"key-1"}, api.deletedKeys)
}

func TestCleanupOrphanedDefaultAnnotations_KeepsConfiguredKeyAndValue(t *testing.T) {
	api := newFakeCleanupAPI()

	key := api.addKey("key-1", "env", defaultAnnotationsMarker)
	api.addValue(key, "val-prod", "prod", defaultAnnotationsMarker)

	require.NoError(t, cleanupOrphanedDefaultAnnotations(t.Context(), api, "org-1", map[string]string{"env": "prod"}))

	assert.Empty(t, api.deletedValues)
	assert.Empty(t, api.deletedKeys)
}

func TestCleanupOrphanedDefaultAnnotations_KeepsConfiguredValueAndUnmarkedValue(t *testing.T) {
	api := newFakeCleanupAPI()

	key := api.addKey("key-1", "env", defaultAnnotationsMarker)
	api.addValue(key, "val-prod", "prod", defaultAnnotationsMarker)
	api.addValue(key, "val-stale", "stale", defaultAnnotationsMarker)
	api.addValue(key, "val-manual", "manual", "user managed")

	require.NoError(t, cleanupOrphanedDefaultAnnotations(t.Context(), api, "org-1", map[string]string{"env": "prod"}))

	assert.Equal(t, []string{"val-stale"}, api.deletedValues)
	assert.Empty(t, api.deletedKeys)
}

func TestCleanupOrphanedDefaultAnnotations_KeepsBoundValue(t *testing.T) {
	api := newFakeCleanupAPI()

	key := api.addKey("key-1", "env", defaultAnnotationsMarker)
	api.addValue(key, "val-1", "prod", defaultAnnotationsMarker)

	api.bindingsByValueID["val-1"] = 2

	require.NoError(t, cleanupOrphanedDefaultAnnotations(t.Context(), api, "org-1", nil))

	assert.Empty(t, api.deletedValues)
	assert.Empty(t, api.deletedKeys)
}

func TestCleanupOrphanedDefaultAnnotations_IgnoresUnmarkedKey(t *testing.T) {
	api := newFakeCleanupAPI()

	key := api.addKey("key-1", "env", "user key")
	api.addValue(key, "val-1", "prod", "user value")

	require.NoError(t, cleanupOrphanedDefaultAnnotations(t.Context(), api, "org-1", nil))

	assert.Empty(t, api.deletedValues)
	assert.Empty(t, api.deletedKeys)
}

func TestCleanupOrphanedDefaultAnnotations_KeepsKeyWithRemainingValues(t *testing.T) {
	api := newFakeCleanupAPI()

	key := api.addKey("key-1", "env", defaultAnnotationsMarker)
	api.addValue(key, "val-manual", "manual", "user value")

	require.NoError(t, cleanupOrphanedDefaultAnnotations(t.Context(), api, "org-1", nil))

	assert.Empty(t, api.deletedValues)
	assert.Empty(t, api.deletedKeys)
}

func TestCleanupOrphanedDefaultAnnotations_ListError(t *testing.T) {
	api := newFakeCleanupAPI()
	api.listAllErr = errors.New("boom")

	require.Error(t, cleanupOrphanedDefaultAnnotations(t.Context(), api, "org-1", nil))
}

func TestWithDefaultBindings_NonComputedSrnIsNoOp(t *testing.T) {
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"srn": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Scaleway Resource Name (SRN) of the resource",
			},
		},
	}

	wrapped := WithDefaultBindings(r)

	// A resource whose "srn" is not Computed is not wrapped: the wrap rule
	// is the same as for the Framework wrapper.
	assert.Same(t, r, wrapped)
	assert.False(t, wrappedSDKv2Resources[r])

	computed := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"srn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Scaleway Resource Name (SRN) of the resource",
			},
		},
	}

	WithDefaultBindings(computed)

	assert.True(t, wrappedSDKv2Resources[computed])
}

// --- fakeBindingAPI tests ---

// fakeBindingAPI is an in-memory fake of the annotations binding API for
// unit testing createDefaultBindings, reconcileDefaultBindings and
// deleteDefaultBindings. ListBindings respects pagination parameters.
type fakeBindingAPI struct {
	created   []*annotationsSDK.CreateBindingRequest
	deleted   []string
	listErr   error
	createErr error
	deleteErr error
	bindings  []*annotationsSDK.Binding
}

var _ bindingAPI = (*fakeBindingAPI)(nil)

func newFakeBindingAPI(bindings ...*annotationsSDK.Binding) *fakeBindingAPI {
	return &fakeBindingAPI{bindings: bindings}
}

func (f *fakeBindingAPI) ListBindings(req *annotationsSDK.ListBindingsRequest, _ ...scw.RequestOption) (*annotationsSDK.ListBindingsResponse, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}

	var filtered []*annotationsSDK.Binding

	for _, b := range f.bindings {
		if req.TargetSrn != nil && b.TargetSrn != *req.TargetSrn {
			continue
		}

		if req.ValueID != nil && (b.Value == nil || b.Value.ID != *req.ValueID) {
			continue
		}

		filtered = append(filtered, b)
	}

	total := uint64(len(filtered))

	page := int32(1)
	if req.Page != nil {
		page = *req.Page
	}

	pageSize := uint64(0)
	if req.PageSize != nil {
		pageSize = uint64(*req.PageSize)
	}

	if pageSize == 0 {
		pageSize = uint64(len(filtered))
	}

	start := uint64(page-1) * pageSize
	if start >= total {
		return &annotationsSDK.ListBindingsResponse{TotalCount: total}, nil
	}

	end := start + pageSize

	end = min(end, total)

	return &annotationsSDK.ListBindingsResponse{
		Bindings:   filtered[start:end],
		TotalCount: total,
	}, nil
}

func (f *fakeBindingAPI) CreateBinding(req *annotationsSDK.CreateBindingRequest, _ ...scw.RequestOption) (*annotationsSDK.Binding, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}

	f.created = append(f.created, req)

	binding := &annotationsSDK.Binding{
		ID:    fmt.Sprintf("binding-%d", len(f.created)),
		Srn:   req.TargetSrn,
		Value: &annotationsSDK.BindingValue{ID: req.ValueID},
	}
	f.bindings = append(f.bindings, binding)

	return binding, nil
}

func (f *fakeBindingAPI) DeleteBinding(req *annotationsSDK.DeleteBindingRequest, _ ...scw.RequestOption) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}

	f.deleted = append(f.deleted, req.BindingID)

	var remaining []*annotationsSDK.Binding

	for _, b := range f.bindings {
		if b.ID != req.BindingID {
			remaining = append(remaining, b)
		}
	}

	f.bindings = remaining

	return nil
}

func TestCreateDefaultBindings_CreatesMissing(t *testing.T) {
	api := newFakeBindingAPI()

	dc := &meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", KeyID: "key-1", ValueName: "prod", ValueID: "val-1"},
			{KeyName: "team", KeyID: "key-2", ValueName: "dev", ValueID: "val-2"},
		},
	}

	require.NoError(t, createDefaultBindings(t.Context(), api, "srn://test", dc))

	assert.Len(t, api.created, 2)
	assert.Equal(t, "val-1", api.created[0].ValueID)
	assert.Equal(t, "val-2", api.created[1].ValueID)
}

func TestCreateDefaultBindings_SkipsExisting(t *testing.T) {
	api := newFakeBindingAPI(&annotationsSDK.Binding{
		ID:        "b-1",
		TargetSrn: "srn://test",
		Value:     &annotationsSDK.BindingValue{ID: "val-1"},
	})

	dc := &meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", KeyID: "key-1", ValueName: "prod", ValueID: "val-1"},
			{KeyName: "team", KeyID: "key-2", ValueName: "dev", ValueID: "val-2"},
		},
	}

	require.NoError(t, createDefaultBindings(t.Context(), api, "srn://test", dc))

	assert.Len(t, api.created, 1)
	assert.Equal(t, "val-2", api.created[0].ValueID)
}

func TestReconcileDefaultBindings_DeletesStaleAndCreatesMissing(t *testing.T) {
	api := newFakeBindingAPI(
		&annotationsSDK.Binding{
			ID:        "b-stale",
			TargetSrn: "srn://test",
			Key:       &annotationsSDK.BindingKey{ID: "key-1"},
			Value:     &annotationsSDK.BindingValue{ID: "val-old"},
		},
		&annotationsSDK.Binding{
			ID:        "b-user",
			TargetSrn: "srn://test",
			Key:       &annotationsSDK.BindingKey{ID: "key-user"},
			Value:     &annotationsSDK.BindingValue{ID: "val-user"},
		},
	)

	dc := &meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", KeyID: "key-1", ValueName: "prod", ValueID: "val-prod"},
		},
	}

	require.NoError(t, reconcileDefaultBindings(t.Context(), api, "srn://test", dc))

	assert.Equal(t, []string{"b-stale"}, api.deleted)
	assert.Len(t, api.created, 1)
	assert.Equal(t, "val-prod", api.created[0].ValueID)
}

func TestReconcileDefaultBindings_KeepsNonDefaultKeyBindings(t *testing.T) {
	api := newFakeBindingAPI(&annotationsSDK.Binding{
		ID:        "b-user",
		TargetSrn: "srn://test",
		Key:       &annotationsSDK.BindingKey{ID: "key-user"},
		Value:     &annotationsSDK.BindingValue{ID: "val-user"},
	})

	dc := &meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", KeyID: "key-1", ValueName: "prod", ValueID: "val-prod"},
		},
	}

	require.NoError(t, reconcileDefaultBindings(t.Context(), api, "srn://test", dc))

	assert.Empty(t, api.deleted)
	assert.Len(t, api.created, 1)
	assert.Equal(t, "val-prod", api.created[0].ValueID)
}

func TestReconcileDefaultBindings_KeepsMatchingBinding(t *testing.T) {
	api := newFakeBindingAPI(&annotationsSDK.Binding{
		ID:        "b-1",
		TargetSrn: "srn://test",
		Key:       &annotationsSDK.BindingKey{ID: "key-1"},
		Value:     &annotationsSDK.BindingValue{ID: "val-prod"},
	})

	dc := &meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", KeyID: "key-1", ValueName: "prod", ValueID: "val-prod"},
		},
	}

	require.NoError(t, reconcileDefaultBindings(t.Context(), api, "srn://test", dc))

	assert.Empty(t, api.deleted)
	assert.Empty(t, api.created)
}

func TestDeleteDefaultBindings_OnlyDeletesDefaultKeys(t *testing.T) {
	api := newFakeBindingAPI(
		&annotationsSDK.Binding{
			ID:        "b-default",
			TargetSrn: "srn://test",
			Key:       &annotationsSDK.BindingKey{ID: "key-1"},
			Value:     &annotationsSDK.BindingValue{ID: "val-1"},
		},
		&annotationsSDK.Binding{
			ID:        "b-user",
			TargetSrn: "srn://test",
			Key:       &annotationsSDK.BindingKey{ID: "key-user"},
			Value:     &annotationsSDK.BindingValue{ID: "val-user"},
		},
	)

	dc := &meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", KeyID: "key-1", ValueName: "prod", ValueID: "val-1"},
		},
	}

	require.NoError(t, deleteDefaultBindings(t.Context(), api, "srn://test", dc))

	assert.Equal(t, []string{"b-default"}, api.deleted)
}

func TestDeleteDefaultBindings_NoBindingsForSrn(t *testing.T) {
	api := newFakeBindingAPI(&annotationsSDK.Binding{
		ID:    "b-1",
		Srn:   "srn://other",
		Key:   &annotationsSDK.BindingKey{ID: "key-1"},
		Value: &annotationsSDK.BindingValue{ID: "val-1"},
	})

	dc := &meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", KeyID: "key-1", ValueName: "prod", ValueID: "val-1"},
		},
	}

	require.NoError(t, deleteDefaultBindings(t.Context(), api, "srn://test", dc))

	assert.Empty(t, api.deleted)
}

func TestListAllBindings_Pagination(t *testing.T) {
	srn := "srn://test"

	bindings := make([]*annotationsSDK.Binding, 0, 250)

	for i := range 250 {
		bindings = append(bindings, &annotationsSDK.Binding{
			ID:        fmt.Sprintf("b-%d", i),
			TargetSrn: srn,
		})
	}

	api := newFakeBindingAPI(bindings...)

	result, err := listAllBindings(t.Context(), api, &annotationsSDK.ListBindingsRequest{
		TargetSrn: &srn,
	})
	require.NoError(t, err)

	assert.Len(t, result, 250)
}
