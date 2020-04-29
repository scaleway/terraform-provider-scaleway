// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package registry provides methods and message types of the registry v1 API.
package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/marshaler"
	"github.com/scaleway/scaleway-sdk-go/internal/parameter"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// always import dependencies
var (
	_ fmt.Stringer
	_ json.Unmarshaler
	_ url.URL
	_ net.IP
	_ http.Header
	_ bytes.Reader
	_ time.Time
	_ = strings.Join

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

// API: docker registry API
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type ImageStatus string

const (
	// ImageStatusUnknown is [insert doc].
	ImageStatusUnknown = ImageStatus("unknown")
	// ImageStatusReady is [insert doc].
	ImageStatusReady = ImageStatus("ready")
	// ImageStatusDeleting is [insert doc].
	ImageStatusDeleting = ImageStatus("deleting")
	// ImageStatusError is [insert doc].
	ImageStatusError = ImageStatus("error")
	// ImageStatusLocked is [insert doc].
	ImageStatusLocked = ImageStatus("locked")
)

func (enum ImageStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum ImageStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ImageStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ImageStatus(ImageStatus(tmp).String())
	return nil
}

type ImageVisibility string

const (
	// ImageVisibilityVisibilityUnknown is [insert doc].
	ImageVisibilityVisibilityUnknown = ImageVisibility("visibility_unknown")
	// ImageVisibilityInherit is [insert doc].
	ImageVisibilityInherit = ImageVisibility("inherit")
	// ImageVisibilityPublic is [insert doc].
	ImageVisibilityPublic = ImageVisibility("public")
	// ImageVisibilityPrivate is [insert doc].
	ImageVisibilityPrivate = ImageVisibility("private")
)

func (enum ImageVisibility) String() string {
	if enum == "" {
		// return default value if empty
		return "visibility_unknown"
	}
	return string(enum)
}

func (enum ImageVisibility) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ImageVisibility) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ImageVisibility(ImageVisibility(tmp).String())
	return nil
}

type ListImagesRequestOrderBy string

const (
	// ListImagesRequestOrderByCreatedAtAsc is [insert doc].
	ListImagesRequestOrderByCreatedAtAsc = ListImagesRequestOrderBy("created_at_asc")
	// ListImagesRequestOrderByCreatedAtDesc is [insert doc].
	ListImagesRequestOrderByCreatedAtDesc = ListImagesRequestOrderBy("created_at_desc")
	// ListImagesRequestOrderByNameAsc is [insert doc].
	ListImagesRequestOrderByNameAsc = ListImagesRequestOrderBy("name_asc")
	// ListImagesRequestOrderByNameDesc is [insert doc].
	ListImagesRequestOrderByNameDesc = ListImagesRequestOrderBy("name_desc")
)

func (enum ListImagesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListImagesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListImagesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListImagesRequestOrderBy(ListImagesRequestOrderBy(tmp).String())
	return nil
}

type ListNamespacesRequestOrderBy string

const (
	// ListNamespacesRequestOrderByCreatedAtAsc is [insert doc].
	ListNamespacesRequestOrderByCreatedAtAsc = ListNamespacesRequestOrderBy("created_at_asc")
	// ListNamespacesRequestOrderByCreatedAtDesc is [insert doc].
	ListNamespacesRequestOrderByCreatedAtDesc = ListNamespacesRequestOrderBy("created_at_desc")
	// ListNamespacesRequestOrderByDescriptionAsc is [insert doc].
	ListNamespacesRequestOrderByDescriptionAsc = ListNamespacesRequestOrderBy("description_asc")
	// ListNamespacesRequestOrderByDescriptionDesc is [insert doc].
	ListNamespacesRequestOrderByDescriptionDesc = ListNamespacesRequestOrderBy("description_desc")
	// ListNamespacesRequestOrderByNameAsc is [insert doc].
	ListNamespacesRequestOrderByNameAsc = ListNamespacesRequestOrderBy("name_asc")
	// ListNamespacesRequestOrderByNameDesc is [insert doc].
	ListNamespacesRequestOrderByNameDesc = ListNamespacesRequestOrderBy("name_desc")
)

func (enum ListNamespacesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListNamespacesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListNamespacesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListNamespacesRequestOrderBy(ListNamespacesRequestOrderBy(tmp).String())
	return nil
}

type ListTagsRequestOrderBy string

const (
	// ListTagsRequestOrderByCreatedAtAsc is [insert doc].
	ListTagsRequestOrderByCreatedAtAsc = ListTagsRequestOrderBy("created_at_asc")
	// ListTagsRequestOrderByCreatedAtDesc is [insert doc].
	ListTagsRequestOrderByCreatedAtDesc = ListTagsRequestOrderBy("created_at_desc")
	// ListTagsRequestOrderByNameAsc is [insert doc].
	ListTagsRequestOrderByNameAsc = ListTagsRequestOrderBy("name_asc")
	// ListTagsRequestOrderByNameDesc is [insert doc].
	ListTagsRequestOrderByNameDesc = ListTagsRequestOrderBy("name_desc")
)

func (enum ListTagsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListTagsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListTagsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListTagsRequestOrderBy(ListTagsRequestOrderBy(tmp).String())
	return nil
}

type NamespaceStatus string

const (
	// NamespaceStatusUnknown is [insert doc].
	NamespaceStatusUnknown = NamespaceStatus("unknown")
	// NamespaceStatusReady is [insert doc].
	NamespaceStatusReady = NamespaceStatus("ready")
	// NamespaceStatusDeleting is [insert doc].
	NamespaceStatusDeleting = NamespaceStatus("deleting")
	// NamespaceStatusError is [insert doc].
	NamespaceStatusError = NamespaceStatus("error")
	// NamespaceStatusLocked is [insert doc].
	NamespaceStatusLocked = NamespaceStatus("locked")
)

func (enum NamespaceStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum NamespaceStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *NamespaceStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = NamespaceStatus(NamespaceStatus(tmp).String())
	return nil
}

type TagStatus string

const (
	// TagStatusUnknown is [insert doc].
	TagStatusUnknown = TagStatus("unknown")
	// TagStatusReady is [insert doc].
	TagStatusReady = TagStatus("ready")
	// TagStatusDeleting is [insert doc].
	TagStatusDeleting = TagStatus("deleting")
	// TagStatusError is [insert doc].
	TagStatusError = TagStatus("error")
	// TagStatusLocked is [insert doc].
	TagStatusLocked = TagStatus("locked")
)

func (enum TagStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum TagStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *TagStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = TagStatus(TagStatus(tmp).String())
	return nil
}

// Image: image
type Image struct {
	// ID: the unique ID of the Image
	ID string `json:"id"`
	// Name: the Image name, unique in a namespace
	Name string `json:"name"`
	// NamespaceID: the unique ID of the Namespace the image belongs to
	NamespaceID string `json:"namespace_id"`
	// Status: the status of the image
	//
	// Default value: unknown
	Status ImageStatus `json:"status"`
	// StatusMessage: details of the image status
	StatusMessage *string `json:"status_message"`
	// Visibility: a `public` image is pullable from internet without authentication, opposed to a `private` image. `inherit` will use the namespace `is_public` parameter
	//
	// Default value: visibility_unknown
	Visibility ImageVisibility `json:"visibility"`
	// Size: image size in bytes, calculated from the size of image layers
	//
	// Image size in bytes, calculated from the size of image layers. One layer used in two tags of the same image is counted once but one layer used in two images is counted twice.
	Size scw.Size `json:"size"`
	// CreatedAt: creation date
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt: last modification date, from the user or the service
	UpdatedAt time.Time `json:"updated_at"`
	// Tags: list of docker tags of the image
	Tags []string `json:"tags"`
}

// ListImagesResponse: list images response
type ListImagesResponse struct {
	// Images: paginated list of images matching filters
	Images []*Image `json:"images"`
	// TotalCount: total number of images matching filters
	TotalCount uint32 `json:"total_count"`
}

// ListNamespacesResponse: list namespaces response
type ListNamespacesResponse struct {
	// Namespaces: paginated list of namespaces matching filters
	Namespaces []*Namespace `json:"namespaces"`
	// TotalCount: total number of namespaces matching filters
	TotalCount uint32 `json:"total_count"`
}

// ListTagsResponse: list tags response
type ListTagsResponse struct {
	// Tags: paginated list of tags matching filters
	Tags []*Tag `json:"tags"`
	// TotalCount: total number of tags matching filters
	TotalCount uint32 `json:"total_count"`
}

// Namespace: namespace
type Namespace struct {
	// ID: the unique ID of the namespace
	ID string `json:"id"`
	// Name: the name of the namespace, unique in a region accross all organizations
	Name string `json:"name"`
	// Description: description of the namespace
	Description string `json:"description"`
	// OrganizationID: owner of the namespace
	OrganizationID string `json:"organization_id"`
	// Status: namespace status
	//
	// Default value: unknown
	Status NamespaceStatus `json:"status"`
	// StatusMessage: namespace status details
	StatusMessage string `json:"status_message"`
	// Endpoint: endpoint reachable by docker
	Endpoint string `json:"endpoint"`
	// IsPublic: namespace visibility policy
	IsPublic bool `json:"is_public"`
	// Size: total size of the namespace, calculated as the sum of the size of all images in the namespace
	Size scw.Size `json:"size"`
	// CreatedAt: creation date
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt: last modification date, from the user or the service
	UpdatedAt time.Time `json:"updated_at"`
	// ImageCount: number of images in the namespace
	ImageCount uint32 `json:"image_count"`
	// Region: region the namespace belongs to
	Region scw.Region `json:"region"`
}

// Tag: tag
type Tag struct {
	// ID: the unique ID of the tag
	ID string `json:"id"`
	// Name: tag name, unique for an image
	Name string `json:"name"`
	// ImageID: image ID this tag belongs to
	ImageID string `json:"image_id"`
	// Status: tag status
	//
	// Default value: unknown
	Status TagStatus `json:"status"`
	// Digest: hash of the tag actual content. Several tags of a same image may have the same digest
	Digest string `json:"digest"`
	// CreatedAt: creation date
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt: last modification date, from the user or the service
	UpdatedAt time.Time `json:"updated_at"`
}

// Service API

type ListNamespacesRequest struct {
	Region scw.Region `json:"-"`
	// Page: a positive integer to choose the page to display
	Page *int32 `json:"-"`
	// PageSize: a positive integer lower or equal to 100 to select the number of items to display
	PageSize *uint32 `json:"-"`
	// OrderBy: field by which to order the display of Images
	//
	// Default value: created_at_asc
	OrderBy ListNamespacesRequestOrderBy `json:"-"`
	// OrganizationID: filter by the namespace owner
	OrganizationID *string `json:"-"`
	// Name: filter by the namespace name (exact match)
	Name *string `json:"-"`
}

// ListNamespaces: list all your namespaces
func (s *API) ListNamespaces(req *ListNamespacesRequest, opts ...scw.RequestOption) (*ListNamespacesResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/registry/v1/regions/" + fmt.Sprint(req.Region) + "/namespaces",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListNamespacesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListNamespacesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListNamespacesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListNamespacesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Namespaces = append(r.Namespaces, results.Namespaces...)
	r.TotalCount += uint32(len(results.Namespaces))
	return uint32(len(results.Namespaces)), nil
}

type GetNamespaceRequest struct {
	Region scw.Region `json:"-"`
	// NamespaceID: the unique ID of the Namespace
	NamespaceID string `json:"-"`
}

// GetNamespace: get a namespace
//
// Get the namespace associated with the given id.
func (s *API) GetNamespace(req *GetNamespaceRequest, opts ...scw.RequestOption) (*Namespace, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.NamespaceID) == "" {
		return nil, errors.New("field NamespaceID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/registry/v1/regions/" + fmt.Sprint(req.Region) + "/namespaces/" + fmt.Sprint(req.NamespaceID) + "",
		Headers: http.Header{},
	}

	var resp Namespace

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateNamespaceRequest struct {
	Region scw.Region `json:"-"`
	// Name: define a namespace name
	Name string `json:"name"`
	// Description: define a description
	Description string `json:"description"`
	// OrganizationID: define the namespace owner
	OrganizationID string `json:"organization_id"`
	// IsPublic: define the default visibility policy
	IsPublic bool `json:"is_public"`
}

// CreateNamespace: create a new namespace
func (s *API) CreateNamespace(req *CreateNamespaceRequest, opts ...scw.RequestOption) (*Namespace, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/registry/v1/regions/" + fmt.Sprint(req.Region) + "/namespaces",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Namespace

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateNamespaceRequest struct {
	Region scw.Region `json:"-"`
	// NamespaceID: namespace ID to update
	NamespaceID string `json:"-"`
	// Description: define a description
	Description *string `json:"description"`
	// IsPublic: define the default visibility policy
	IsPublic *bool `json:"is_public"`
}

// UpdateNamespace: update an existing namespace
//
// Update the namespace associated with the given id.
func (s *API) UpdateNamespace(req *UpdateNamespaceRequest, opts ...scw.RequestOption) (*Namespace, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.NamespaceID) == "" {
		return nil, errors.New("field NamespaceID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/registry/v1/regions/" + fmt.Sprint(req.Region) + "/namespaces/" + fmt.Sprint(req.NamespaceID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Namespace

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteNamespaceRequest struct {
	Region scw.Region `json:"-"`
	// NamespaceID: the unique ID of the Namespace
	NamespaceID string `json:"-"`
}

// DeleteNamespace: delete an existing namespace
//
// Delete the namespace associated with the given id.
func (s *API) DeleteNamespace(req *DeleteNamespaceRequest, opts ...scw.RequestOption) (*Namespace, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.NamespaceID) == "" {
		return nil, errors.New("field NamespaceID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/registry/v1/regions/" + fmt.Sprint(req.Region) + "/namespaces/" + fmt.Sprint(req.NamespaceID) + "",
		Headers: http.Header{},
	}

	var resp Namespace

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListImagesRequest struct {
	Region scw.Region `json:"-"`
	// Page: a positive integer to choose the page to display
	Page *int32 `json:"-"`
	// PageSize: a positive integer lower or equal to 100 to select the number of items to display
	PageSize *uint32 `json:"-"`
	// OrderBy: field by which to order the display of Images
	//
	// Default value: created_at_asc
	OrderBy ListImagesRequestOrderBy `json:"-"`
	// NamespaceID: filter by the Namespace ID
	NamespaceID *string `json:"-"`
	// Name: filter by the Image name (exact match)
	Name *string `json:"-"`
	// OrganizationID: filter by Organization ID
	OrganizationID *string `json:"-"`
}

// ListImages: list all your images
func (s *API) ListImages(req *ListImagesRequest, opts ...scw.RequestOption) (*ListImagesResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "namespace_id", req.NamespaceID)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/registry/v1/regions/" + fmt.Sprint(req.Region) + "/images",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListImagesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListImagesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Images = append(r.Images, results.Images...)
	r.TotalCount += uint32(len(results.Images))
	return uint32(len(results.Images)), nil
}

type GetImageRequest struct {
	Region scw.Region `json:"-"`
	// ImageID: the unique ID of the Image
	ImageID string `json:"-"`
}

// GetImage: get a image
//
// Get the image associated with the given id.
func (s *API) GetImage(req *GetImageRequest, opts ...scw.RequestOption) (*Image, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ImageID) == "" {
		return nil, errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/registry/v1/regions/" + fmt.Sprint(req.Region) + "/images/" + fmt.Sprint(req.ImageID) + "",
		Headers: http.Header{},
	}

	var resp Image

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateImageRequest struct {
	Region scw.Region `json:"-"`
	// ImageID: image ID to update
	ImageID string `json:"-"`
	// Visibility: a `public` image is pullable from internet without authentication, opposed to a `private` image. `inherit` will use the namespace `is_public` parameter
	//
	// Default value: visibility_unknown
	Visibility ImageVisibility `json:"visibility"`
}

// UpdateImage: update an existing image
//
// Update the image associated with the given id.
func (s *API) UpdateImage(req *UpdateImageRequest, opts ...scw.RequestOption) (*Image, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ImageID) == "" {
		return nil, errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/registry/v1/regions/" + fmt.Sprint(req.Region) + "/images/" + fmt.Sprint(req.ImageID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Image

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteImageRequest struct {
	Region scw.Region `json:"-"`
	// ImageID: the unique ID of the Image
	ImageID string `json:"-"`
}

// DeleteImage: delete an image
//
// Delete the image associated with the given id.
func (s *API) DeleteImage(req *DeleteImageRequest, opts ...scw.RequestOption) (*Image, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ImageID) == "" {
		return nil, errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/registry/v1/regions/" + fmt.Sprint(req.Region) + "/images/" + fmt.Sprint(req.ImageID) + "",
		Headers: http.Header{},
	}

	var resp Image

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListTagsRequest struct {
	Region scw.Region `json:"-"`
	// ImageID: the unique ID of the image
	ImageID string `json:"-"`
	// Page: a positive integer to choose the page to display
	Page *int32 `json:"-"`
	// PageSize: a positive integer lower or equal to 100 to select the number of items to display
	PageSize *uint32 `json:"-"`
	// OrderBy: field by which to order the display of Images
	//
	// Default value: created_at_asc
	OrderBy ListTagsRequestOrderBy `json:"-"`
	// Name: filter by the tag name (exact match)
	Name *string `json:"-"`
}

// ListTags: list all your tags
func (s *API) ListTags(req *ListTagsRequest, opts ...scw.RequestOption) (*ListTagsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ImageID) == "" {
		return nil, errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/registry/v1/regions/" + fmt.Sprint(req.Region) + "/images/" + fmt.Sprint(req.ImageID) + "/tags",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListTagsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListTagsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListTagsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListTagsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Tags = append(r.Tags, results.Tags...)
	r.TotalCount += uint32(len(results.Tags))
	return uint32(len(results.Tags)), nil
}

type GetTagRequest struct {
	Region scw.Region `json:"-"`
	// TagID: the unique ID of the Tag
	TagID string `json:"-"`
}

// GetTag: get a tag
//
// Get the tag associated with the given id.
func (s *API) GetTag(req *GetTagRequest, opts ...scw.RequestOption) (*Tag, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.TagID) == "" {
		return nil, errors.New("field TagID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/registry/v1/regions/" + fmt.Sprint(req.Region) + "/tags/" + fmt.Sprint(req.TagID) + "",
		Headers: http.Header{},
	}

	var resp Tag

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteTagRequest struct {
	Region scw.Region `json:"-"`
	// TagID: the unique ID of the tag
	TagID string `json:"-"`
	// Force: if two tags share the same digest the deletion will fail unless this parameter is set to true
	Force bool `json:"-"`
}

// DeleteTag: delete a tag
//
// Delete the tag associated with the given id.
func (s *API) DeleteTag(req *DeleteTagRequest, opts ...scw.RequestOption) (*Tag, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	query := url.Values{}
	parameter.AddToQuery(query, "force", req.Force)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.TagID) == "" {
		return nil, errors.New("field TagID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/registry/v1/regions/" + fmt.Sprint(req.Region) + "/tags/" + fmt.Sprint(req.TagID) + "",
		Query:   query,
		Headers: http.Header{},
	}

	var resp Tag

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
