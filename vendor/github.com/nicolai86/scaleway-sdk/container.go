package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ContainerData represents a  container data (S3)
type ContainerData struct {
	LastModified string `json:"last_modified"`
	Name         string `json:"name"`
	Size         string `json:"size"`
}

// GetContainerDatas represents a list of  containers data (S3)
type GetContainerDatas struct {
	Container []ContainerData `json:"container"`
}

// Container represents a  container (S3)
type Container struct {
	OrganizationDefinition `json:"organization"`
	Name                   string `json:"name"`
	Size                   string `json:"size"`
}

// GetContainers represents a list of  containers (S3)
type GetContainers struct {
	Containers []Container `json:"containers"`
}

// GetContainers returns a GetContainers
func (s *API) GetContainers() (*GetContainers, error) {
	resp, err := s.GetResponsePaginate(s.computeAPI, "containers", url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var containers GetContainers

	if err = json.Unmarshal(body, &containers); err != nil {
		return nil, err
	}
	return &containers, nil
}

// GetContainerDatas returns a GetContainerDatas
func (s *API) GetContainerDatas(container string) (*GetContainerDatas, error) {
	resp, err := s.GetResponsePaginate(s.computeAPI, fmt.Sprintf("containers/%s", container), url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var datas GetContainerDatas

	if err = json.Unmarshal(body, &datas); err != nil {
		return nil, err
	}
	return &datas, nil
}
