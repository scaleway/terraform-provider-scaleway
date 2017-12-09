package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// IPV6 represents a  ipv6
type IPV6 struct {
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
	Address string `json:"address"`
}

// IPV4 represents the IP's fields
type IPV4 struct {
	Organization string  `json:"organization"`
	Reverse      *string `json:"reverse"`
	ID           string  `json:"id"`
	Server       *struct {
		Identifier string `json:"id,omitempty"`
		Name       string `json:"name,omitempty"`
	} `json:"server"`
	Address string `json:"address"`
}

// IPAddress represents a  IP address
type IPAddress struct {
	// Identifier is a unique identifier for the IP address
	Identifier string `json:"id,omitempty"`

	// IP is an IPv4 address
	IP string `json:"address,omitempty"`

	// Dynamic is a flag that defines an IP that change on each reboot
	Dynamic *bool `json:"dynamic,omitempty"`
}

// GetIPS represents the response of a GET /ips/
type GetIPS struct {
	IPS []IPV4 `json:"ips"`
}

// GetIP represents the response of a GET /ips/{id_ip}
type GetIP struct {
	IP IPV4 `json:"ip"`
}

// GetIP returns a GetIP
func (s *API) GetIP(ipID string) (*GetIP, error) {
	resp, err := s.GetResponsePaginate(s.computeAPI, fmt.Sprintf("ips/%s", ipID), url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var ip GetIP

	if err = json.Unmarshal(body, &ip); err != nil {
		return nil, err
	}
	return &ip, nil
}

// GetIPS returns a GetIPS
func (s *API) GetIPS() (*GetIPS, error) {
	resp, err := s.GetResponsePaginate(s.computeAPI, "ips", url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var ips GetIPS

	if err = json.Unmarshal(body, &ips); err != nil {
		return nil, err
	}
	return &ips, nil
}

// NewIP returns a new IP
func (s *API) NewIP() (*GetIP, error) {
	var orga struct {
		Organization string `json:"organization"`
	}
	orga.Organization = s.Organization
	resp, err := s.PostResponse(s.computeAPI, "ips", orga)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusCreated}, resp)
	if err != nil {
		return nil, err
	}
	var ip GetIP

	if err = json.Unmarshal(body, &ip); err != nil {
		return nil, err
	}
	return &ip, nil
}

// AttachIP attachs an IP to a server
func (s *API) AttachIP(ipID, serverID string) error {
	var update struct {
		Address      string  `json:"address"`
		ID           string  `json:"id"`
		Reverse      *string `json:"reverse"`
		Organization string  `json:"organization"`
		Server       string  `json:"server"`
	}

	ip, err := s.GetIP(ipID)
	if err != nil {
		return err
	}
	update.Address = ip.IP.Address
	update.ID = ip.IP.ID
	update.Organization = ip.IP.Organization
	update.Server = serverID
	resp, err := s.PutResponse(s.computeAPI, fmt.Sprintf("ips/%s", ipID), update)
	if err != nil {
		return err
	}
	_, err = s.handleHTTPError([]int{http.StatusOK}, resp)
	return err
}

// DetachIP detaches an IP from a server
func (s *API) DetachIP(ipID string) error {
	ip, err := s.GetIP(ipID)
	if err != nil {
		return err
	}
	ip.IP.Server = nil
	resp, err := s.PutResponse(s.computeAPI, fmt.Sprintf("ips/%s", ipID), ip.IP)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = s.handleHTTPError([]int{http.StatusOK}, resp)
	return err
}

// DeleteIP deletes an IP
func (s *API) DeleteIP(ipID string) error {
	resp, err := s.DeleteResponse(s.computeAPI, fmt.Sprintf("ips/%s", ipID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = s.handleHTTPError([]int{http.StatusNoContent}, resp)
	return err
}
