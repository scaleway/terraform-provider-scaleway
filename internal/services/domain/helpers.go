package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/api/std"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultWaitDomainsRegistrationRetryInterval = 10 * time.Minute
)

// NewDomainAPI returns a new domain API.
func NewDomainAPI(m any) *domain.API {
	return domain.NewAPI(meta.ExtractScwClient(m))
}

// NewRegistrarDomainAPI returns a new registrar API.
func NewRegistrarDomainAPI(m any) *domain.RegistrarAPI {
	return domain.NewRegistrarAPI(meta.ExtractScwClient(m))
}

func getRecordFromTypeAndData(dnsType domain.RecordType, data string, records []*domain.Record) (*domain.Record, error) {
	var currentRecord *domain.Record

	for _, r := range records {
		flattedData := FlattenDomainData(strings.ToLower(r.Data), r.Type).(string)
		flattenCurrentData := FlattenDomainData(strings.ToLower(data), r.Type).(string)

		if dnsType == domain.RecordTypeSRV {
			if flattedData == flattenCurrentData {
				if currentRecord != nil {
					return nil, fmt.Errorf("multiple records found with same type and data: existing record %s (ID: %s) conflicts with new record data %s", currentRecord.Data, currentRecord.ID, data)
				}

				currentRecord = r

				break
			}
		} else {
			if strings.HasPrefix(flattedData, flattenCurrentData) && r.Type == dnsType {
				if currentRecord != nil {
					return nil, fmt.Errorf("multiple records found with same type and data: existing record %s (ID: %s) conflicts with new record data %s", currentRecord.Data, currentRecord.ID, data)
				}

				currentRecord = r

				break
			}
		}
	}

	if currentRecord == nil {
		return nil, fmt.Errorf("record with type %s and data %s not found", dnsType.String(), data)
	}

	return currentRecord, nil
}

func FindDefaultReverse(address string) string {
	parts := strings.Split(address, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, "-") + ".instances.scw.cloud"
}

func ExpandContact(contactMap map[string]any) *domain.Contact {
	if contactMap == nil {
		return nil
	}

	contact := &domain.Contact{
		PhoneNumber:  contactMap["phone_number"].(string),
		LegalForm:    domain.ContactLegalForm(contactMap["legal_form"].(string)),
		Firstname:    contactMap["firstname"].(string),
		Lastname:     contactMap["lastname"].(string),
		Email:        contactMap["email"].(string),
		AddressLine1: contactMap["address_line_1"].(string),
		Zip:          contactMap["zip"].(string),
		City:         contactMap["city"].(string),
		Country:      contactMap["country"].(string),
	}

	// Optional fields
	if v, ok := contactMap["company_name"].(string); ok && v != "" {
		contact.CompanyName = v
	}

	if v, ok := contactMap["email_alt"].(string); ok && v != "" {
		contact.EmailAlt = v
	}

	if v, ok := contactMap["fax_number"].(string); ok && v != "" {
		contact.FaxNumber = v
	}

	if v, ok := contactMap["address_line_2"].(string); ok && v != "" {
		contact.AddressLine2 = v
	}

	if v, ok := contactMap["vat_identification_code"].(string); ok && v != "" {
		contact.VatIDentificationCode = v
	}

	if v, ok := contactMap["company_identification_code"].(string); ok && v != "" {
		contact.CompanyIDentificationCode = v
	}

	if v, ok := contactMap["lang"].(string); ok && v != "" {
		contact.Lang = std.LanguageCode(v)
	}

	if v, ok := contactMap["resale"].(bool); ok {
		contact.Resale = v
	}

	if v, ok := contactMap["state"].(string); ok && v != "" {
		contact.State = v
	}

	if v, ok := contactMap["whois_opt_in"].(bool); ok {
		contact.WhoisOptIn = v
	}

	if extFr, ok := contactMap["extension_fr"].(map[string]any); ok && len(extFr) > 0 {
		extension := expandContactExtension(extFr, "fr")
		if extension != nil {
			contact.ExtensionFr = extension.(*domain.ContactExtensionFR)
		}
	}

	if extEu, ok := contactMap["extension_eu"].(map[string]any); ok && len(extEu) > 0 {
		extension := expandContactExtension(extEu, "eu")
		if extension != nil {
			contact.ExtensionEu = extension.(*domain.ContactExtensionEU)
		}
	}

	if extNl, ok := contactMap["extension_nl"].(map[string]any); ok && len(extNl) > 0 {
		extension := expandContactExtension(extNl, "nl")
		if extension != nil {
			contact.ExtensionNl = extension.(*domain.ContactExtensionNL)
		}
	}

	return contact
}

func expandContactExtension(extensionMap map[string]any, extensionType string) any {
	if len(extensionMap) == 0 {
		return nil
	}

	switch extensionType {
	case "fr":
		return &domain.ContactExtensionFR{
			Mode:              domain.ContactExtensionFRMode(parseEnum(extensionMap, "mode", domain.ContactExtensionFRModeModeUnknown.String())),
			IndividualInfo:    parseStruct[domain.ContactExtensionFRIndividualInfo](extensionMap, "individual_info"),
			DunsInfo:          parseStruct[domain.ContactExtensionFRDunsInfo](extensionMap, "duns_info"),
			AssociationInfo:   parseStruct[domain.ContactExtensionFRAssociationInfo](extensionMap, "association_info"),
			TrademarkInfo:     parseStruct[domain.ContactExtensionFRTrademarkInfo](extensionMap, "trademark_info"),
			CodeAuthAfnicInfo: parseStruct[domain.ContactExtensionFRCodeAuthAfnicInfo](extensionMap, "code_auth_afnic_info"),
		}

	case "nl":
		legalFormRegistrationNumber := ""
		value, ok := extensionMap["legal_form_registration_number"]

		if ok {
			str, isString := value.(string)
			if isString {
				legalFormRegistrationNumber = str
			}
		}

		return &domain.ContactExtensionNL{
			LegalForm:                   domain.ContactExtensionNLLegalForm(parseEnum(extensionMap, "legal_form", domain.ContactExtensionNLLegalFormLegalFormUnknown.String())),
			LegalFormRegistrationNumber: legalFormRegistrationNumber,
		}

	case "eu":
		europeanCitizenship := ""

		if value, ok := extensionMap["european_citizenship"]; ok {
			if str, isString := value.(string); isString {
				europeanCitizenship = str
			}
		}

		return &domain.ContactExtensionEU{
			EuropeanCitizenship: europeanCitizenship,
		}

	default:
		return nil
	}
}

func ExpandNewContact(contactMap map[string]any) *domain.NewContact {
	if contactMap == nil {
		return nil
	}

	contact := &domain.NewContact{
		PhoneNumber:  contactMap["phone_number"].(string),
		LegalForm:    domain.ContactLegalForm(contactMap["legal_form"].(string)),
		Firstname:    contactMap["firstname"].(string),
		Lastname:     contactMap["lastname"].(string),
		Email:        contactMap["email"].(string),
		AddressLine1: contactMap["address_line_1"].(string),
		Zip:          contactMap["zip"].(string),
		City:         contactMap["city"].(string),
		Country:      contactMap["country"].(string),
	}

	if v, ok := contactMap["resale"].(bool); ok {
		contact.Resale = v
	} else {
		contact.Resale = false
	}

	if v, ok := contactMap["whois_opt_in"].(bool); ok {
		contact.WhoisOptIn = v
	} else {
		contact.WhoisOptIn = false
	}

	if v, ok := contactMap["company_name"].(string); ok {
		contact.CompanyName = scw.StringPtr(v)
	}

	if v, ok := contactMap["email_alt"].(string); ok {
		contact.EmailAlt = scw.StringPtr(v)
	}

	if v, ok := contactMap["fax_number"].(string); ok {
		contact.FaxNumber = scw.StringPtr(v)
	}

	if v, ok := contactMap["address_line_2"].(string); ok {
		contact.AddressLine2 = scw.StringPtr(v)
	}

	if v, ok := contactMap["vat_identification_code"].(string); ok {
		contact.VatIDentificationCode = scw.StringPtr(v)
	}

	if v, ok := contactMap["company_identification_code"].(string); ok {
		contact.CompanyIDentificationCode = scw.StringPtr(v)
	}

	if v, ok := contactMap["state"].(string); ok {
		contact.State = scw.StringPtr(v)
	}

	if extFr, ok := contactMap["extension_fr"].(map[string]any); ok && len(extFr) > 0 {
		extension := expandContactExtension(extFr, "fr")
		if extension != nil {
			contact.ExtensionFr = extension.(*domain.ContactExtensionFR)
		}
	}

	if extEu, ok := contactMap["extension_eu"].(map[string]any); ok && len(extEu) > 0 {
		extension := expandContactExtension(extEu, "eu")
		if extension != nil {
			contact.ExtensionEu = extension.(*domain.ContactExtensionEU)
		}
	}

	if extNl, ok := contactMap["extension_nl"].(map[string]any); ok && len(extNl) > 0 {
		extension := expandContactExtension(extNl, "nl")
		if extension != nil {
			contact.ExtensionNl = extension.(*domain.ContactExtensionNL)
		}
	}

	return contact
}

func parseEnum(data map[string]any, key string, defaultValue string) string {
	if value, ok := data[key].(string); ok {
		return value
	}

	return defaultValue
}

func parseStruct[T any](data map[string]any, key string) *T {
	if nested, ok := data[key].(map[string]any); ok {
		var result T

		mapToStruct(nested, &result)

		return &result
	}

	return nil
}

func mapToStruct(data map[string]any, target any) {
	switch t := target.(type) {
	case *domain.ContactExtensionFRIndividualInfo:
		if v, ok := data["whois_opt_in"].(bool); ok {
			t.WhoisOptIn = v
		}

	case *domain.ContactExtensionFRDunsInfo:
		if v, ok := data["duns_id"].(string); ok {
			t.DunsID = v
		}

		if v, ok := data["local_id"].(string); ok {
			t.LocalID = v
		}

	case *domain.ContactExtensionFRAssociationInfo:
		if v, ok := data["publication_jo"].(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, v); err == nil {
				t.PublicationJo = &parsedTime
			}
		}

		if v, ok := data["publication_jo_page"].(float64); ok {
			t.PublicationJoPage = uint32(v)
		}

	case *domain.ContactExtensionFRTrademarkInfo:
		if v, ok := data["trademark_inpi"].(string); ok {
			t.TrademarkInpi = v
		}

	case *domain.ContactExtensionFRCodeAuthAfnicInfo:
		if v, ok := data["code_auth_afnic"].(string); ok {
			t.CodeAuthAfnic = v
		}
	}
}

func getStatusTasks(ctx context.Context, api *domain.RegistrarAPI, taskID string) (domain.TaskStatus, error) {
	var page int32 = 1

	var pageSize uint32 = 1000

	for {
		listTasksResponse, err := api.ListTasks(&domain.RegistrarAPIListTasksRequest{
			Page:     &page,
			PageSize: &pageSize,
		}, scw.WithContext(ctx))
		if err != nil {
			return "", fmt.Errorf("error retrieving tasks: %w", err)
		}

		for _, task := range listTasksResponse.Tasks {
			if task.ID == taskID {
				return task.Status, nil
			}
		}

		if len(listTasksResponse.Tasks) == 0 || uint32(len(listTasksResponse.Tasks)) < pageSize {
			break
		}

		page++
	}

	return "", fmt.Errorf("task with ID '%s' not found", taskID)
}

func SplitDomains(input *string) []string {
	if input == nil || strings.TrimSpace(*input) == "" {
		return nil
	}

	domains := strings.Split(*input, ",")

	var result []string

	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			result = append(result, domain)
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func ExtractDomainsFromTaskID(ctx context.Context, id string, registrarAPI *domain.RegistrarAPI) ([]string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ID format, expected 'projectID/domainName', got: %s", id)
	}

	taskID := parts[1]

	listTasksResponse, err := registrarAPI.ListTasks(&domain.RegistrarAPIListTasksRequest{}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("error retrieving tasks: %w", err)
	}

	for _, task := range listTasksResponse.Tasks {
		if task.ID == taskID {
			return SplitDomains(task.Domain), nil
		}
	}

	return nil, fmt.Errorf("task with ID '%s' not found", taskID)
}

func flattenContact(contact *domain.Contact) []map[string]any {
	if contact == nil {
		return nil
	}

	flattened := map[string]any{
		"phone_number":                contact.PhoneNumber,
		"legal_form":                  string(contact.LegalForm),
		"firstname":                   contact.Firstname,
		"lastname":                    contact.Lastname,
		"email":                       contact.Email,
		"address_line_1":              contact.AddressLine1,
		"zip":                         contact.Zip,
		"city":                        contact.City,
		"country":                     contact.Country,
		"company_name":                contact.CompanyName,
		"email_alt":                   contact.EmailAlt,
		"fax_number":                  contact.FaxNumber,
		"address_line_2":              contact.AddressLine2,
		"vat_identification_code":     contact.VatIDentificationCode,
		"company_identification_code": contact.CompanyIDentificationCode,
		"lang":                        string(contact.Lang),
		"resale":                      contact.Resale,
		"state":                       contact.State,
		"whois_opt_in":                contact.WhoisOptIn,
	}

	if contact.ExtensionFr != nil {
		flattened["extension_fr"] = flattenContactExtensionFR(contact.ExtensionFr)
	}

	if contact.ExtensionEu != nil {
		flattened["extension_eu"] = flattenContactExtensionEU(contact.ExtensionEu)
	}

	if contact.ExtensionNl != nil {
		flattened["extension_nl"] = flattenContactExtensionNL(contact.ExtensionNl)
	}

	return []map[string]any{flattened}
}

func flattenContactExtensionFR(ext *domain.ContactExtensionFR) []map[string]any {
	if ext == nil {
		return nil
	}

	return []map[string]any{
		{
			"mode":                 string(ext.Mode),
			"individual_info":      flattenContactExtensionFRIndividualInfo(ext.IndividualInfo),
			"duns_info":            flattenContactExtensionFRDunsInfo(ext.DunsInfo),
			"association_info":     flattenContactExtensionFRAssociationInfo(ext.AssociationInfo),
			"trademark_info":       flattenContactExtensionFRTrademarkInfo(ext.TrademarkInfo),
			"code_auth_afnic_info": flattenContactExtensionFRCodeAuthAfnicInfo(ext.CodeAuthAfnicInfo),
		},
	}
}

func flattenContactExtensionFRIndividualInfo(info *domain.ContactExtensionFRIndividualInfo) []map[string]any {
	if info == nil {
		return nil
	}

	return []map[string]any{
		{
			"whois_opt_in": info.WhoisOptIn,
		},
	}
}

func flattenContactExtensionFRDunsInfo(info *domain.ContactExtensionFRDunsInfo) []map[string]any {
	if info == nil {
		return nil
	}

	return []map[string]any{
		{
			"duns_id":  info.DunsID,
			"local_id": info.LocalID,
		},
	}
}

func flattenContactExtensionFRAssociationInfo(info *domain.ContactExtensionFRAssociationInfo) []map[string]any {
	if info == nil {
		return nil
	}

	return []map[string]any{
		{
			"publication_jo":      info.PublicationJo.Format(time.RFC3339),
			"publication_jo_page": info.PublicationJoPage,
		},
	}
}

func flattenContactExtensionFRTrademarkInfo(info *domain.ContactExtensionFRTrademarkInfo) []map[string]any {
	if info == nil {
		return nil
	}

	return []map[string]any{
		{
			"trademark_inpi": info.TrademarkInpi,
		},
	}
}

func flattenContactExtensionFRCodeAuthAfnicInfo(info *domain.ContactExtensionFRCodeAuthAfnicInfo) []map[string]any {
	if info == nil {
		return nil
	}

	return []map[string]any{
		{
			"code_auth_afnic": info.CodeAuthAfnic,
		},
	}
}

func flattenContactExtensionEU(ext *domain.ContactExtensionEU) []map[string]any {
	if ext == nil {
		return nil
	}

	return []map[string]any{
		{
			"european_citizenship": ext.EuropeanCitizenship,
		},
	}
}

func flattenContactExtensionNL(ext *domain.ContactExtensionNL) []map[string]any {
	if ext == nil {
		return nil
	}

	return []map[string]any{
		{
			"legal_form":                     string(ext.LegalForm),
			"legal_form_registration_number": ext.LegalFormRegistrationNumber,
		},
	}
}

func waitForTaskCompletion(ctx context.Context, registrarAPI *domain.RegistrarAPI, taskID string, duration int) error {
	timeout := time.Duration(duration) * time.Second

	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		status, err := getStatusTasks(ctx, registrarAPI, taskID)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("failed to retrieve task status: %w", err))
		}

		if status == domain.TaskStatusPending || status == domain.TaskStatusWaitingPayment || status == domain.TaskStatusNew {
			return retry.RetryableError(errors.New("task is not yet complete, retrying"))
		}

		if status == domain.TaskStatusSuccess {
			return nil
		}

		if status == domain.TaskStatusError {
			return retry.NonRetryableError(fmt.Errorf("task failed for domain: %s", taskID))
		}

		return retry.NonRetryableError(fmt.Errorf("unexpected task status: %v", status))
	})
}

func FlattenDSRecord(dsRecords []*domain.DSRecord) []any {
	if len(dsRecords) == 0 {
		return []any{}
	}

	results := make([]any, 0, len(dsRecords))

	for _, dsRecord := range dsRecords {
		item := map[string]any{
			"key_id":    dsRecord.KeyID,
			"algorithm": string(dsRecord.Algorithm),
		}

		if dsRecord.Digest != nil {
			digest := map[string]any{
				"type":   string(dsRecord.Digest.Type),
				"digest": dsRecord.Digest.Digest,
			}

			if dsRecord.Digest.PublicKey != nil {
				digest["public_key"] = []any{
					map[string]any{
						"key": dsRecord.Digest.PublicKey.Key,
					},
				}
			}

			item["digest"] = []any{digest}
		}

		if dsRecord.PublicKey != nil {
			item["public_key"] = []any{
				map[string]any{
					"key": dsRecord.PublicKey.Key,
				},
			}
		}

		results = append(results, item)
	}

	return results
}

func BuildZoneName(subdomain, domain string) string {
	if subdomain == "" {
		return domain
	}

	return fmt.Sprintf("%s.%s", subdomain, domain)
}
