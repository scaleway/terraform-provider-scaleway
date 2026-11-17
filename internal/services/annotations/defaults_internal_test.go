package annotations

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	annotationsSDK "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testOrganizationID = "11111111-1111-1111-1111-111111111111"
	testTargetSrn      = "scw-fip-abcdef0123456789"
)

func testKeyName(name string) string         { return "key-" + name }
func testValueName(key, value string) string { return "value-" + key + "-" + value }

func testKey(name string) *annotationsSDK.ListAllKeysAndValuesResponseKey {
	return &annotationsSDK.ListAllKeysAndValuesResponseKey{
		ID:   testKeyName(name),
		Name: name,
		Values: []*annotationsSDK.ListAllKeysAndValuesResponseValue{
			{ID: testValueName(name, "platform"), Name: "platform"},
			{ID: testValueName(name, "legacy"), Name: "legacy"},
		},
	}
}

func testBinding(id, keyName, valueName string) *annotationsSDK.Binding {
	return &annotationsSDK.Binding{
		ID:        id,
		TargetSrn: testTargetSrn,
		Key:       &annotationsSDK.BindingKey{ID: testKeyName(keyName), Name: keyName},
		Value:     &annotationsSDK.BindingValue{ID: testValueName(keyName, valueName), Name: valueName},
	}
}

type fakeBindingManager struct {
	listErr   error
	createErr error
	deleteErr func(bindingID string) error
	bindings  []*annotationsSDK.Binding
	created   []annotationsSDK.CreateBindingRequest
	deleted   []string
}

func (f *fakeBindingManager) ListBindings(request *annotationsSDK.ListBindingsRequest, options ...scw.RequestOption) (*annotationsSDK.ListBindingsResponse, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}

	filtered := f.bindings
	if request.TargetSrn != nil {
		filtered = nil

		for _, binding := range f.bindings {
			if binding.TargetSrn == *request.TargetSrn {
				filtered = append(filtered, binding)
			}
		}
	}

	return &annotationsSDK.ListBindingsResponse{
		Bindings:   filtered,
		TotalCount: uint64(len(filtered)),
	}, nil
}

func (f *fakeBindingManager) CreateBinding(request *annotationsSDK.CreateBindingRequest, options ...scw.RequestOption) (*annotationsSDK.Binding, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}

	f.created = append(f.created, *request)

	return &annotationsSDK.Binding{ID: "created-binding-id", TargetSrn: request.TargetSrn}, nil
}

func (f *fakeBindingManager) DeleteBinding(request *annotationsSDK.DeleteBindingRequest, options ...scw.RequestOption) error {
	f.deleted = append(f.deleted, request.BindingID)

	if f.deleteErr != nil {
		return f.deleteErr(request.BindingID)
	}

	return nil
}

func (f *fakeBindingManager) createdBindings() []annotationsSDK.CreateBindingRequest {
	return append([]annotationsSDK.CreateBindingRequest(nil), f.created...)
}

func (f *fakeBindingManager) deletedBindings() []string {
	return append([]string(nil), f.deleted...)
}

func testKeysAndValuesResponse(keys ...*annotationsSDK.ListAllKeysAndValuesResponseKey) *annotationsSDK.ListAllKeysAndValuesResponse {
	return &annotationsSDK.ListAllKeysAndValuesResponse{Keys: keys}
}

func TestResolveWishlist(t *testing.T) {
	t.Run("empty wishlist", func(t *testing.T) {
		resolved, err := resolveWishlist(testKeysAndValuesResponse(testKey("team")), nil)

		require.NoError(t, err)
		assert.Empty(t, resolved)
	})

	t.Run("resolves keys and values sorted by key name", func(t *testing.T) {
		resolved, err := resolveWishlist(testKeysAndValuesResponse(testKey("team"), testKey("env")), map[string]string{"env": "platform", "team": "legacy"})

		require.NoError(t, err)
		require.Len(t, resolved, 2)
		assert.Equal(t, []DefaultAnnotation{
			{KeyName: "env", KeyID: testKeyName("env"), ValueName: "platform", ValueID: testValueName("env", "platform")},
			{KeyName: "team", KeyID: testKeyName("team"), ValueName: "legacy", ValueID: testValueName("team", "legacy")},
		}, resolved)
	})

	t.Run("missing key", func(t *testing.T) {
		_, err := resolveWishlist(testKeysAndValuesResponse(testKey("team")), map[string]string{"env": "platform"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), `annotation key "env" not found`)
	})

	t.Run("missing value", func(t *testing.T) {
		_, err := resolveWishlist(testKeysAndValuesResponse(testKey("team")), map[string]string{"team": "production"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), `annotation value "production" for key "team" not found`)
	})
}

func TestListAllBindings(t *testing.T) {
	t.Run("returns all bindings", func(t *testing.T) {
		api := &fakeBindingManager{bindings: []*annotationsSDK.Binding{testBinding("b-1", "team", "platform"), testBinding("b-2", "env", "platform")}}

		bindings, err := listAllBindings(t.Context(), api, &annotationsSDK.ListBindingsRequest{TargetSrn: &testTargetSrn})

		require.NoError(t, err)
		require.Len(t, bindings, 2)
		assert.Equal(t, "b-1", bindings[0].ID)
		assert.Equal(t, "b-2", bindings[1].ID)
	})

	t.Run("list error", func(t *testing.T) {
		api := &fakeBindingManager{listErr: errors.New("boom")}

		_, err := listAllBindings(t.Context(), api, &annotationsSDK.ListBindingsRequest{})

		require.Error(t, err)
		assert.ErrorIs(t, err, api.listErr)
	})
}

func TestReconcileWishlist(t *testing.T) {
	t.Run("no-op without wishlist or srn", func(t *testing.T) {
		api := &fakeBindingManager{}

		require.NoError(t, reconcileWishlist(t.Context(), api, testTargetSrn, nil))
		require.NoError(t, reconcileWishlist(t.Context(), api, "", nil))

		assert.Empty(t, api.createdBindings())
		assert.Empty(t, api.deletedBindings())
	})

	t.Run("creates missing bindings", func(t *testing.T) {
		api := &fakeBindingManager{}
		wishlist := []DefaultAnnotation{{KeyName: "team", KeyID: testKeyName("team"), ValueName: "platform", ValueID: testValueName("team", "platform")}}

		require.NoError(t, reconcileWishlist(t.Context(), api, testTargetSrn, wishlist))

		created := api.createdBindings()
		require.Len(t, created, 1)
		assert.Equal(t, testTargetSrn, created[0].TargetSrn)
		assert.Equal(t, testValueName("team", "platform"), created[0].ValueID)
		assert.Empty(t, api.deletedBindings())
	})

	t.Run("deletes stale bindings and recreates them", func(t *testing.T) {
		api := &fakeBindingManager{bindings: []*annotationsSDK.Binding{testBinding("b-stale", "team", "legacy")}}
		wishlist := []DefaultAnnotation{{KeyName: "team", KeyID: testKeyName("team"), ValueName: "platform", ValueID: testValueName("team", "platform")}}

		require.NoError(t, reconcileWishlist(t.Context(), api, testTargetSrn, wishlist))

		assert.Equal(t, []string{"b-stale"}, api.deletedBindings())

		created := api.createdBindings()
		require.Len(t, created, 1)
		assert.Equal(t, testValueName("team", "platform"), created[0].ValueID)
	})

	t.Run("does not recreate the expected binding when a stale duplicate is deleted", func(t *testing.T) {
		// The expected binding is listed before the stale duplicate: deleting
		// the duplicate must not make the expected binding look missing.
		api := &fakeBindingManager{bindings: []*annotationsSDK.Binding{
			testBinding("b-ok", "team", "platform"),
			testBinding("b-stale", "team", "legacy"),
		}}
		wishlist := []DefaultAnnotation{{KeyName: "team", KeyID: testKeyName("team"), ValueName: "platform", ValueID: testValueName("team", "platform")}}

		require.NoError(t, reconcileWishlist(t.Context(), api, testTargetSrn, wishlist))

		assert.Equal(t, []string{"b-stale"}, api.deletedBindings())
		assert.Empty(t, api.createdBindings())
	})

	t.Run("ignores 404 on delete", func(t *testing.T) {
		api := &fakeBindingManager{
			bindings: []*annotationsSDK.Binding{testBinding("b-stale", "team", "legacy")},
			deleteErr: func(bindingID string) error {
				return &scw.ResponseError{Message: "not found", StatusCode: http.StatusNotFound}
			},
		}
		wishlist := []DefaultAnnotation{{KeyName: "team", KeyID: testKeyName("team"), ValueName: "platform", ValueID: testValueName("team", "platform")}}

		require.NoError(t, reconcileWishlist(t.Context(), api, testTargetSrn, wishlist))

		created := api.createdBindings()
		require.Len(t, created, 1)
		assert.Equal(t, testValueName("team", "platform"), created[0].ValueID)
	})

	t.Run("propagates non-404 delete error", func(t *testing.T) {
		api := &fakeBindingManager{
			bindings: []*annotationsSDK.Binding{testBinding("b-stale", "team", "legacy")},
			deleteErr: func(bindingID string) error {
				return &scw.ResponseError{Message: "internal error", StatusCode: http.StatusInternalServerError}
			},
		}
		wishlist := []DefaultAnnotation{{KeyName: "team", KeyID: testKeyName("team"), ValueName: "platform", ValueID: testValueName("team", "platform")}}

		err := reconcileWishlist(t.Context(), api, testTargetSrn, wishlist)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "deleting stale binding")
	})

	t.Run("ignores bindings under other keys", func(t *testing.T) {
		api := &fakeBindingManager{bindings: []*annotationsSDK.Binding{testBinding("b-other", "env", "platform")}}
		wishlist := []DefaultAnnotation{{KeyName: "team", KeyID: testKeyName("team"), ValueName: "platform", ValueID: testValueName("team", "platform")}}

		require.NoError(t, reconcileWishlist(t.Context(), api, testTargetSrn, wishlist))

		assert.Empty(t, api.deletedBindings())

		created := api.createdBindings()
		require.Len(t, created, 1)
		assert.Equal(t, testValueName("team", "platform"), created[0].ValueID)
	})

	t.Run("no-op when up to date", func(t *testing.T) {
		api := &fakeBindingManager{bindings: []*annotationsSDK.Binding{testBinding("b-ok", "team", "platform")}}
		wishlist := []DefaultAnnotation{{KeyName: "team", KeyID: testKeyName("team"), ValueName: "platform", ValueID: testValueName("team", "platform")}}

		require.NoError(t, reconcileWishlist(t.Context(), api, testTargetSrn, wishlist))

		assert.Empty(t, api.createdBindings())
		assert.Empty(t, api.deletedBindings())
	})

	t.Run("propagates create error", func(t *testing.T) {
		api := &fakeBindingManager{createErr: errors.New("boom")}
		wishlist := []DefaultAnnotation{{KeyName: "team", KeyID: testKeyName("team"), ValueName: "platform", ValueID: testValueName("team", "platform")}}

		err := reconcileWishlist(t.Context(), api, testTargetSrn, wishlist)

		require.Error(t, err)
		assert.ErrorIs(t, err, api.createErr)
	})
}

func TestDeleteWishlistBindings(t *testing.T) {
	t.Run("no-op without wishlist or srn", func(t *testing.T) {
		api := &fakeBindingManager{bindings: []*annotationsSDK.Binding{testBinding("b-1", "team", "platform")}}

		require.NoError(t, deleteWishlistBindings(t.Context(), api, testTargetSrn, nil))
		require.NoError(t, deleteWishlistBindings(t.Context(), api, "", map[string]string{"team": "platform"}))

		assert.Empty(t, api.deletedBindings())
	})

	t.Run("deletes only bindings under default keys", func(t *testing.T) {
		api := &fakeBindingManager{bindings: []*annotationsSDK.Binding{
			testBinding("b-team", "team", "platform"),
			testBinding("b-env", "env", "platform"),
		}}

		require.NoError(t, deleteWishlistBindings(t.Context(), api, testTargetSrn, map[string]string{"team": "platform", "other": "value"}))

		assert.Equal(t, []string{"b-team"}, api.deletedBindings())
	})

	t.Run("ignores 404", func(t *testing.T) {
		api := &fakeBindingManager{
			bindings: []*annotationsSDK.Binding{testBinding("b-team", "team", "platform")},
			deleteErr: func(bindingID string) error {
				return &scw.ResponseError{Message: "not found", StatusCode: http.StatusNotFound}
			},
		}

		require.NoError(t, deleteWishlistBindings(t.Context(), api, testTargetSrn, map[string]string{"team": "platform"}))
	})

	t.Run("propagates non-404 error", func(t *testing.T) {
		api := &fakeBindingManager{
			bindings: []*annotationsSDK.Binding{testBinding("b-team", "team", "platform")},
			deleteErr: func(bindingID string) error {
				return &scw.ResponseError{Message: "internal error", StatusCode: http.StatusInternalServerError}
			},
		}

		err := deleteWishlistBindings(t.Context(), api, testTargetSrn, map[string]string{"team": "platform"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "deleting binding")
	})
}

func TestReconcileDefaultBindingsForSrn(t *testing.T) {
	t.Run("no-op without a meta", func(t *testing.T) {
		require.NoError(t, ReconcileDefaultBindingsForSrn(t.Context(), nil, testTargetSrn))
	})

	t.Run("no-op without default annotations", func(t *testing.T) {
		server := &fakeAnnotationsServer{}
		m := newTestMeta(t, server, nil)

		require.NoError(t, ReconcileDefaultBindingsForSrn(t.Context(), m, testTargetSrn))

		assert.Empty(t, server.created)
		assert.Equal(t, 0, server.listCalls)
	})

	t.Run("reconciles the default annotations", func(t *testing.T) {
		server := &fakeAnnotationsServer{keys: []*annotationsSDK.ListAllKeysAndValuesResponseKey{testKey("team")}}
		m := newTestMeta(t, server, map[string]string{"team": "platform"})

		require.NoError(t, ReconcileDefaultBindingsForSrn(t.Context(), m, testTargetSrn))

		assertCreatedBinding(t, server, testValueName("team", "platform"))
	})

	t.Run("caches the organization keys and values listing", func(t *testing.T) {
		server := &fakeAnnotationsServer{keys: []*annotationsSDK.ListAllKeysAndValuesResponseKey{testKey("team")}}
		m := newTestMeta(t, server, map[string]string{"team": "platform"})

		require.NoError(t, ReconcileDefaultBindingsForSrn(t.Context(), m, testTargetSrn))
		require.NoError(t, ReconcileDefaultBindingsForSrn(t.Context(), m, testTargetSrn))

		assert.Equal(t, 1, server.listCalls)
		assert.Len(t, server.created, 2)
	})

	t.Run("invalidating the cache triggers a new listing", func(t *testing.T) {
		server := &fakeAnnotationsServer{keys: []*annotationsSDK.ListAllKeysAndValuesResponseKey{testKey("team")}}
		m := newTestMeta(t, server, map[string]string{"team": "platform"})

		require.NoError(t, ReconcileDefaultBindingsForSrn(t.Context(), m, testTargetSrn))
		require.Equal(t, 1, server.listCalls)

		m.InvalidateAnnotationCache()

		require.NoError(t, ReconcileDefaultBindingsForSrn(t.Context(), m, testTargetSrn))

		assert.Equal(t, 2, server.listCalls)
	})

	t.Run("returns a resolution error when a key cannot be resolved", func(t *testing.T) {
		server := &fakeAnnotationsServer{}
		m := newTestMeta(t, server, map[string]string{"missing": "platform"})

		err := ReconcileDefaultBindingsForSrn(t.Context(), m, testTargetSrn)

		require.Error(t, err)
		assert.Contains(t, err.Error(), `annotation key "missing" not found`)
	})

	t.Run("returns a resolution error when the keys and values listing fails", func(t *testing.T) {
		server := &fakeAnnotationsServer{listError: http.StatusInternalServerError}
		m := newTestMeta(t, server, map[string]string{"team": "platform"})

		err := ReconcileDefaultBindingsForSrn(t.Context(), m, testTargetSrn)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "listing annotation keys and values")
	})
}

// fakeAnnotationsServer serves the annotations API endpoints with the given
// keys and initial bindings. It records the create and delete binding calls.
type fakeAnnotationsServer struct {
	keys        []*annotationsSDK.ListAllKeysAndValuesResponseKey
	bindings    []*annotationsSDK.Binding
	created     []annotationsSDK.CreateBindingRequest
	deleted     []string
	createError int
	listError   int
	listCalls   int
}

func (s *fakeAnnotationsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/annotations/v1/all-keys-and-values":
		s.listCalls++

		if s.listError != 0 {
			http.Error(w, "internal error", s.listError)

			return
		}

		writeTestJSON(w, &annotationsSDK.ListAllKeysAndValuesResponse{Keys: s.keys})
	case r.Method == http.MethodGet && r.URL.Path == "/annotations/v1/bindings":
		const pageSize = 2

		page := 1
		if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
			page = p
		}

		start := min((page-1)*pageSize, len(s.bindings))

		writeTestJSON(w, &annotationsSDK.ListBindingsResponse{Bindings: s.bindings[start:min(start+pageSize, len(s.bindings))], TotalCount: uint64(len(s.bindings))})
	case r.Method == http.MethodPost && r.URL.Path == "/annotations/v1/bindings":
		var req annotationsSDK.CreateBindingRequest

		_ = json.NewDecoder(r.Body).Decode(&req)

		s.created = append(s.created, req)

		if s.createError != 0 {
			http.Error(w, "internal error", s.createError)

			return
		}

		writeTestJSON(w, &annotationsSDK.Binding{ID: "created-binding-id", TargetSrn: req.TargetSrn})
	case r.Method == http.MethodDelete && len(r.URL.Path) > len("/annotations/v1/bindings/"):
		s.deleted = append(s.deleted, r.URL.Path[len("/annotations/v1/bindings/"):])

		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func writeTestJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return
	}
}

func assertCreatedBinding(t *testing.T, server *fakeAnnotationsServer, valueID string) {
	t.Helper()

	require.Len(t, server.created, 1)
	assert.Equal(t, testTargetSrn, server.created[0].TargetSrn)
	assert.Equal(t, valueID, server.created[0].ValueID)
}

func newTestMeta(t *testing.T, server *fakeAnnotationsServer, wishlist map[string]string) *meta.Meta {
	t.Helper()

	httpServer := httptest.NewServer(server)
	t.Cleanup(httpServer.Close)

	profile := &scw.Profile{
		APIURL:                &httpServer.URL,
		AccessKey:             new("SCWTESTACCESSKEY0000"),
		SecretKey:             new("22222222-2222-2222-2222-222222222222"),
		DefaultOrganizationID: &testOrganizationID,
	}

	m, err := meta.NewMetaFromProfile(t.Context(), profile, &meta.CredentialsSource{}, nil, nil, "test", &http.Client{Transport: http.DefaultTransport}, wishlist)
	require.NoError(t, err)

	return m
}
