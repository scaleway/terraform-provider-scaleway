package ephemeral

import (
	"context"
	"fmt"

	annotations "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// This file provides infrastructure for managing ephemeral resources with annotation-based identifier tracking
// (or description-based for resources equipped with a `description` field and a List API with `description` filter).
//
// This enables our Terraform provider to manage ephemeral resources by using Scaleway annotations/descriptions
// as a persistent identifying mechanism. It supports two identifying strategies:
//
//   - Annotation-based identifier (annotation_identifier): Resources are identified by a set of annotation tags.
//     On creation, we ensure there is at most one existing resource with matching annotations.
//   - Description-based identifier (description_identifier): Resources are identified by a unique description string.
//     On creation, we query via ListApiKeys filtered by description.
//
// In both cases:
//   - If multiple resources match: return an error.
//   - If replace_resource=true: Create the new resource, then delete the previous one (if any exists).
//   - If replace_resource=false: If a matching resource exists, leave it unchanged; otherwise create a new one.
//
// The term `identifier` and its derivatives are deliberately used to prevent any confusion with Terraform
// Resource Identities (https://developer.hashicorp.com/terraform/plugin/framework/resources/identity).
//
// Key considerations:
//   - Annotation quotas may apply.
//   - This mechanism has some limitations: any interference with descriptions/annotations will make the
//     pseudo-lifecycle fail. This system relies on users ensuring the identifier value they set is unique
//     and is not modified in any way.
//   - This is a custom solution to a limitation inherent to current Terraform ephemeral resources: without
//     a way to retrieve a resource from a hardcoded value in the Terraform configuration, any resource
//     created via an ephemeral block is lost to the provider (since they are not stored in state).
//     This is why currently ephemeral resources are only appropriate as "clean-state datasources",
//     not as actual resources.
//   - Our solution is far from perfect, but it is so far our only solution to allow resource creation
//     and management while keeping a clean Terraform state. We sincerely hope a Terraform-integrated
//     solution will come, maybe a `read-once` feature that we will gladly migrate to.
//     See: https://github.com/hashicorp/terraform/issues/34860

const (
	DefaultTerraformAnnotationKey = "terraform_identifier"
)

type AnnotationsAPI interface {
	ListKeys(req *annotations.ListKeysRequest, opts ...scw.RequestOption) (*annotations.ListKeysResponse, error)
	CreateKey(req *annotations.CreateKeyRequest, opts ...scw.RequestOption) (*annotations.Key, error)
	ListValues(req *annotations.ListValuesRequest, opts ...scw.RequestOption) (*annotations.ListValuesResponse, error)
	CreateValue(req *annotations.CreateValueRequest, opts ...scw.RequestOption) (*annotations.Value, error)
	DeleteValue(req *annotations.DeleteValueRequest, opts ...scw.RequestOption) error
	ListBindings(req *annotations.ListBindingsRequest, opts ...scw.RequestOption) (*annotations.ListBindingsResponse, error)
	CreateBinding(req *annotations.CreateBindingRequest, opts ...scw.RequestOption) (*annotations.Binding, error)
	DeleteBinding(req *annotations.DeleteBindingRequest, opts ...scw.RequestOption) error
}

type ResourceHandler[T any] interface {
	FindByAnnotationValue(ctx context.Context, srn string) (*T, error)
	FindByDescription(ctx context.Context, description string, organizationID string) (*T, error)
	GetByID(ctx context.Context, id string) (*T, error)
	DeleteByID(ctx context.Context, id string) error
}

type ResourceIdentifierManager[T any] struct {
	annotationsAPI  AnnotationsAPI
	ResourceHandler ResourceHandler[T]
	annotationKey   string
}

type ResourceIdentifierManagerConfig[T any] struct {
	AnnotationsAPI  AnnotationsAPI
	ResourceHandler ResourceHandler[T]
	AnnotationKey   string
}

func NewResourceIdentifierManager[T any](config ResourceIdentifierManagerConfig[T]) *ResourceIdentifierManager[T] {
	annotationKey := config.AnnotationKey
	if annotationKey == "" {
		annotationKey = DefaultTerraformAnnotationKey
	}

	return &ResourceIdentifierManager[T]{
		annotationsAPI:  config.AnnotationsAPI,
		ResourceHandler: config.ResourceHandler,
		annotationKey:   annotationKey,
	}
}

func (m *ResourceIdentifierManager[T]) GetAnnotationKeyID(ctx context.Context, organizationID string) (string, error) {
	keysResp, err := m.annotationsAPI.ListKeys(&annotations.ListKeysRequest{
		OrganizationID: organizationID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return "", fmt.Errorf("failed to list annotation keys: %w", err)
	}

	for _, key := range keysResp.Keys {
		if key.Name == m.annotationKey {
			return key.ID, nil
		}
	}

	return "", nil
}

func (m *ResourceIdentifierManager[T]) GetOrCreateAnnotationKey(ctx context.Context, organizationID string) (string, error) {
	keyID, err := m.GetAnnotationKeyID(ctx, organizationID)
	if err != nil {
		return "", err
	}

	if keyID != "" {
		return keyID, nil
	}

	newKey, err := m.annotationsAPI.CreateKey(&annotations.CreateKeyRequest{
		OrganizationID: organizationID,
		Name:           m.annotationKey,
		Description:    "Identifier key for Terraform-managed resources",
	}, scw.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("failed to create identifier key: %w", err)
	}

	return newKey.ID, nil
}

func (m *ResourceIdentifierManager[T]) GetAnnotationValueID(ctx context.Context, keyID, annotationValue string) (string, error) {
	valuesResp, err := m.annotationsAPI.ListValues(&annotations.ListValuesRequest{
		KeyID: &keyID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return "", fmt.Errorf("failed to list annotation values: %w", err)
	}

	for _, value := range valuesResp.Values {
		if value.Name == annotationValue {
			return value.ID, nil
		}
	}

	return "", nil
}

func (m *ResourceIdentifierManager[T]) GetOrCreateValue(ctx context.Context, keyID, annotationValue string) (string, error) {
	identifierValueID, err := m.GetAnnotationValueID(ctx, keyID, annotationValue)
	if err != nil {
		return "", err
	}

	if identifierValueID != "" {
		return identifierValueID, nil
	}

	newValue, err := m.annotationsAPI.CreateValue(&annotations.CreateValueRequest{
		KeyID:       keyID,
		Name:        annotationValue,
		Description: "Identifier: " + annotationValue,
	}, scw.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("failed to create annotation identifier value: %w", err)
	}

	return newValue.ID, nil
}

func (m *ResourceIdentifierManager[T]) GetAnnotationBinding(ctx context.Context, organizationID, valueID string) (*annotations.Binding, error) {
	bindingsResp, err := m.annotationsAPI.ListBindings(&annotations.ListBindingsRequest{
		OrganizationID: organizationID,
		ValueID:        &valueID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list bindings: %w", err)
	}

	if len(bindingsResp.Bindings) == 0 {
		return nil, nil
	}

	if len(bindingsResp.Bindings) > 1 {
		return nil, fmt.Errorf("found %d bindings for annotation identifier value, expected at most 1", len(bindingsResp.Bindings))
	}

	return bindingsResp.Bindings[0], nil
}

func (m *ResourceIdentifierManager[T]) FindResourceByAnnotation(ctx context.Context, annotationValue, organizationID string) (*T, error) {
	keyID, err := m.GetAnnotationKeyID(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	if keyID == "" {
		return nil, nil
	}

	valueID, err := m.GetAnnotationValueID(ctx, keyID, annotationValue)
	if err != nil {
		return nil, err
	}

	if valueID == "" {
		return nil, nil
	}

	binding, err := m.GetAnnotationBinding(ctx, organizationID, valueID)
	if err != nil {
		return nil, err
	}

	if binding == nil {
		return nil, nil
	}

	return m.ResourceHandler.FindByAnnotationValue(ctx, binding.Srn)
}

func (m *ResourceIdentifierManager[T]) FindResourceByDescription(ctx context.Context, description, organizationID string) (*T, error) {
	return m.ResourceHandler.FindByDescription(ctx, description, organizationID)
}

func (m *ResourceIdentifierManager[T]) GetOrCreateAnnotationBinding(ctx context.Context, resourceSRN, identifierValueID, organizationID string) error {
	existingBindingsResp, err := m.annotationsAPI.ListBindings(&annotations.ListBindingsRequest{
		OrganizationID: organizationID,
		ValueID:        &identifierValueID,
		Srn:            &resourceSRN,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return fmt.Errorf("failed to list existing bindings: %w", err)
	}

	if len(existingBindingsResp.Bindings) > 1 {
		return fmt.Errorf("found %d bindings for the given SRN and value ID, expected at most 1", len(existingBindingsResp.Bindings))
	}

	if len(existingBindingsResp.Bindings) == 1 {
		return nil
	}

	_, err = m.annotationsAPI.CreateBinding(&annotations.CreateBindingRequest{
		Srn:     resourceSRN,
		ValueID: identifierValueID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to create binding: %w", err)
	}

	return nil
}

func (m *ResourceIdentifierManager[T]) GetOrCreateAnnotationIdentifier(ctx context.Context, resourceSRN, annotationValue, organizationID string) error {
	annotationKeyID, err := m.GetOrCreateAnnotationKey(ctx, organizationID)
	if annotationKeyID == "" || err != nil {
		return err
	}

	identifierValueID, err := m.GetOrCreateValue(ctx, annotationKeyID, annotationValue)
	if identifierValueID == "" || err != nil {
		return err
	}

	return m.GetOrCreateAnnotationBinding(ctx, resourceSRN, identifierValueID, organizationID)
}

func (m *ResourceIdentifierManager[T]) GetResourceByID(ctx context.Context, id string) (*T, error) {
	return m.ResourceHandler.GetByID(ctx, id)
}

func (m *ResourceIdentifierManager[T]) DeleteResourceByID(ctx context.Context, id string) error {
	return m.ResourceHandler.DeleteByID(ctx, id)
}

func (m *ResourceIdentifierManager[T]) DeleteAnnotationIdentifier(ctx context.Context, annotationValue, organizationID string) error {
	keyID, err := m.GetAnnotationKeyID(ctx, organizationID)
	if err != nil {
		return err
	}

	if keyID == "" {
		return nil
	}

	valueID, err := m.GetAnnotationValueID(ctx, keyID, annotationValue)
	if err != nil {
		return err
	}

	if valueID == "" {
		return nil
	}

	binding, err := m.GetAnnotationBinding(ctx, organizationID, valueID)
	if err != nil {
		return err
	}

	if binding != nil {
		if err := m.annotationsAPI.DeleteBinding(&annotations.DeleteBindingRequest{
			BindingID: binding.ID,
		}, scw.WithContext(ctx)); err != nil {
			return fmt.Errorf("failed to delete binding: %w", err)
		}
	}

	if err := m.annotationsAPI.DeleteValue(&annotations.DeleteValueRequest{
		ValueID: valueID,
	}, scw.WithContext(ctx)); err != nil {
		return fmt.Errorf("failed to delete annotation value: %w", err)
	}

	return nil
}
