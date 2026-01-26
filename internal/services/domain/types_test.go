package domain_test

import (
	"testing"

	domainSDK "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/domain"
)

func TestFlattenDomainData(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		dnsZone    string
		recordType domainSDK.RecordType
		expected   string
	}{
		{
			name:       "SRV record with domain duplication",
			data:       "0 0 1234 foo.noet-ia.com.noet-ia.com.",
			dnsZone:    "noet-ia.com",
			recordType: domainSDK.RecordTypeSRV,
			expected:   "0 0 1234 foo.noet-ia.com",
		},
		{
			name:       "SRV record without domain duplication",
			data:       "0 0 1234 foo.noet-ia.com",
			dnsZone:    "example.com",
			recordType: domainSDK.RecordTypeSRV,
			expected:   "0 0 1234 foo.noet-ia.com",
		},
		{
			name:       "SRV record with complex domain duplication",
			data:       "10 5 8080 service.example.com.example.com.",
			dnsZone:    "example.com",
			recordType: domainSDK.RecordTypeSRV,
			expected:   "10 5 8080 service.example.com",
		},
		{
			name:       "SRV record with no duplication pattern",
			data:       "0 0 1234 foo.bar.com",
			dnsZone:    "example.com",
			recordType: domainSDK.RecordTypeSRV,
			expected:   "0 0 1234 foo.bar.com",
		},
		{
			name:       "SRV record with trailing dot only",
			data:       "0 0 1234 foo.example.com.",
			dnsZone:    "example.com",
			recordType: domainSDK.RecordTypeSRV,
			expected:   "0 0 1234 foo",
		},
		{
			name:       "SRV record with real test case",
			data:       "0 0 1234 foo.example.com.test-srv-duplication.scaleway-terraform.com.",
			dnsZone:    "test-srv-duplication.scaleway-terraform.com",
			recordType: domainSDK.RecordTypeSRV,
			expected:   "0 0 1234 foo.example.com",
		},
		{
			name:       "SRV record with Terraform data (3 parts)",
			data:       "0 1234 foo.example.com",
			dnsZone:    "example.com",
			recordType: domainSDK.RecordTypeSRV,
			expected:   "0 0 1234 foo.example.com",
		},
		{
			name:       "MX record",
			data:       "10 mail.example.com",
			dnsZone:    "example.com",
			recordType: domainSDK.RecordTypeMX,
			expected:   "mail.example.com",
		},
		{
			name:       "TXT record",
			data:       "\"v=spf1 include:_spf.example.com ~all\"",
			dnsZone:    "example.com",
			recordType: domainSDK.RecordTypeTXT,
			expected:   "v=spf1 include:_spf.example.com ~all",
		},
		{
			name:       "A record",
			data:       "192.168.1.1",
			dnsZone:    "example.com",
			recordType: domainSDK.RecordTypeA,
			expected:   "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.FlattenDomainData(tt.data, tt.recordType, tt.dnsZone)
			if result != tt.expected {
				t.Errorf("flattenDomainData(%q, %v, %q) = %q, want %q", tt.data, tt.recordType, tt.dnsZone, result, tt.expected)
			}
		})
	}
}

func TestNormalizeSRVData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		dnsZone  string
		expected string
	}{
		{
			name:     "SRV with weight 0",
			input:    "0 0 8080 server.example.com",
			dnsZone:  "example.com",
			expected: "0 0 8080 server.example.com",
		},
		{
			name:     "SRV with weight 1",
			input:    "0 1 8080 server.example.com",
			dnsZone:  "example.com",
			expected: "0 1 8080 server.example.com",
		},
		{
			name:     "SRV with high weight",
			input:    "0 100 8080 server.example.com",
			dnsZone:  "example.com",
			expected: "0 100 8080 server.example.com",
		},
		{
			name:     "SRV with priority 10",
			input:    "10 0 8080 server.example.com",
			dnsZone:  "example.com",
			expected: "10 0 8080 server.example.com",
		},
		{
			name:     "SRV with port 443",
			input:    "0 0 443 server.example.com",
			dnsZone:  "example.com",
			expected: "0 0 443 server.example.com",
		},
		{
			name:     "SRV with complex target",
			input:    "0 0 8080 server.subdomain.example.com",
			dnsZone:  "example.com",
			expected: "0 0 8080 server.subdomain.example.com",
		},
		{
			name:     "SRV with trailing dot",
			input:    "0 0 8080 server.example.com.",
			dnsZone:  "example.com",
			expected: "0 0 8080 server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.NormalizeSRVData(tt.input, tt.dnsZone)
			if result != tt.expected {
				t.Errorf("normalizeSRVData(%q, %q) = %q, want %q", tt.input, tt.dnsZone, result, tt.expected)
			}
		})
	}
}

func TestRemoveZoneDomainSuffix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		dnsZone  string
		expected string
	}{
		{
			name:     "simple domain",
			input:    "server.example.com",
			dnsZone:  "example.com",
			expected: "server.example.com",
		},
		{
			name:     "with trailing dot",
			input:    "server.example.com.",
			dnsZone:  "example.com",
			expected: "server",
		},
		{
			name:     "with zone domain duplication",
			input:    "server.example.com.zone.tld",
			dnsZone:  "zone.tld",
			expected: "server.example.com.zone.tld",
		},
		{
			name:     "with zone domain duplication and trailing dot",
			input:    "server.example.com.zone.tld.",
			dnsZone:  "zone.tld",
			expected: "server.example.com",
		},
		{
			name:     "with complex zone domain duplication",
			input:    "server.example.com.subdomain.zone.tld",
			dnsZone:  "zone.tld",
			expected: "server.example.com.subdomain.zone.tld",
		},
		{
			name:     "with complex zone domain duplication and trailing dot",
			input:    "server.example.com.subdomain.zone.tld.",
			dnsZone:  "zone.tld",
			expected: "server.example.com.subdomain",
		},
		{
			name:     "no domain duplication pattern",
			input:    "server.example.com.other.tld",
			dnsZone:  "zone.tld",
			expected: "server.example.com.other.tld",
		},
		{
			name:     "no domain duplication pattern with trailing dot",
			input:    "server.example.com.other.tld.",
			dnsZone:  "zone.tld",
			expected: "server.example.com.other.tld.",
		},
		{
			name:     "single word",
			input:    "server",
			dnsZone:  "example.com",
			expected: "server",
		},
		{
			name:     "single word with trailing dot",
			input:    "server.",
			dnsZone:  "example.com",
			expected: "server.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.RemoveZoneDomainSuffix(tt.input, tt.dnsZone)
			if result != tt.expected {
				t.Errorf("removeZoneDomainSuffix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
