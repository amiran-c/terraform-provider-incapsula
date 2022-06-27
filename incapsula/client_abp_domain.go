package incapsula

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	ABPDomainAPIPath = "/botmanagement/v1/domain"
)

type ABPUpdateDomain struct {
	SiteID string `json:"site_id"`
	ChallengeIPLookupMode string `json:"challenge_ip_lookup_mode"`
	AnalysisIPLookupMode string `json:"analysis_ip_lookup_mode"`
	CookieScope string `json:"cookiescope"`
	CaptchaSettings string `json:"captcha_settings"`
	LogRegion string `json:"log_region"`
	NoJSInjectionPaths []string `json:"no_js_injection_paths"`
	ObfuscatePath string `json:"obfuscate_path"`
	CookieMode string `json:"cookie_mode"`
	UnmaskedHeaders []string `json:"unmasked_headers"`
}

type ABPCreateDomain struct {
	ABPUpdateDomain
	Criteria string `json:"criteria"`
	EncryptionKeyID string `json:"encryption_key_id"`
}

type ABPDomainResponse struct {
	ABPCreateDomain
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	MobileAPIObfuscatePath string `json:"mobile_api_obfuscate_path"`
}

type ABPDomainsResponse struct {
	Domains []ABPDomainResponse `json:"items"`
}

func (c *Client) GetABPDomain(domainID string) (*ABPDomainResponse, error) {
	log.Printf("[INFO] Getting ABP domain ID: %s\n", domainID)

	if len(domainID) == 0 {
		return nil, fmt.Errorf("can't fetch empty domain ID")
	}

	resp, err := c.DoJsonRequestWithHeaders(http.MethodGet,
		fmt.Sprintf("%s%s/%s", c.config.BaseURLAPI, ABPDomainAPIPath, domainID),
		nil,
		ReadABPDomain)

	if err != nil {
		return nil, fmt.Errorf("error from ABP API when reading domain ID %s", domainID)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API read policy JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error status code %d from ABP API when reading domain ID %s: %s", resp.StatusCode, domainID, string(responseBody))
	}

	// Parse the JSON
	var abpDomainResponse ABPDomainResponse
	err = json.Unmarshal([]byte(responseBody), &abpDomainResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON response for domain ID %s: %s. response: %s", domainID, err, string(responseBody))
	}

	return &abpDomainResponse, nil
}

func (c *Client) GetABPDomainsForAccount(accountID string) ([]ABPDomainResponse, error) {
	log.Printf("[INFO] Getting ABP domains for account ID: %s\n", accountID)

	if len(accountID) == 0 {
		return nil, fmt.Errorf("can't fetch empty account ID")
	}

	resp, err := c.DoJsonRequestWithHeaders(http.MethodGet,
		fmt.Sprintf("%s%s/%s/%s", c.config.BaseURLAPI, ABPAccountAPIPath, accountID, "domain"),
		nil,
		ReadABPDomain)

	if err != nil {
		return nil, fmt.Errorf("error from ABP API when reading domains for account ID %s", accountID)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API read domains JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error status code %d from ABP API when reading domains for account ID %s: %s", resp.StatusCode, accountID, string(responseBody))
	}

	// Parse the JSON
	var abpDomainsResponse ABPDomainsResponse
	err = json.Unmarshal([]byte(responseBody), &abpDomainsResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON response for domains for account ID %s: %s. response: %s", accountID, err, string(responseBody))
	}

	return abpDomainsResponse.Domains, nil
}

func (c *Client) GetABPDomainsForAccountWithoutID() ([]ABPDomainResponse, error) {
	if ID, err := c.GetABPAccountID(); err != nil {
		return nil, err
	} else {
		return c.GetABPDomainsForAccount(ID)
	}
}

func (c *Client) updateABPDomain(domainID string, val *ABPUpdateDomain) (*ABPDomainResponse, error) {
	log.Printf("[INFO] Updating ABP domain ID: %s", domainID)

	valJSON, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON marshal ABP domain %v: %s", val, err)
	}
	log.Printf("[DEBUG] ABP update domain, about to send request with body: %s", valJSON)

	resp, err := c.DoJsonRequestWithHeaders(http.MethodPut,
		fmt.Sprintf("%s%s/%s", c.config.BaseURLAPI, ABPDomainAPIPath, domainID),
		valJSON,
		UpdateABPDomain)
	if err != nil {
		return nil, fmt.Errorf("error from ABP API while updating domain %s: %s", domainID, err)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API PUT domain data JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error status code %d from ABP API when updating domain %s: %s\n",
			resp.StatusCode, domainID, string(responseBody))
	}

	// Parse the JSON
	var updatedDomain ABPDomainResponse
	err = json.Unmarshal([]byte(responseBody), &updatedDomain)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON response for ABP domain ID %s: %s\nresponse: %s\n",
			domainID, err, string(responseBody))
	}

	return &updatedDomain, nil
}

func (c *Client) deleteABPDomain(domainID string) error {
	log.Printf("[INFO] Deleting ABP domain %s", domainID)

	resp, err := c.DoJsonRequestWithHeaders(http.MethodDelete,
		fmt.Sprintf("%s%s/%s", c.config.BaseURLAPI, ABPDomainAPIPath, domainID),
		nil,
		DeleteABPDomain)
	if err != nil {
		return fmt.Errorf("error from ABP API while deleting domain ID %s: %s", domainID, err)
	}

	// Read the body
	defer resp.Body.Close()

	// Check the response code - no content for DELETE
	if resp.StatusCode != 200 {
		return fmt.Errorf("error status code %d from CSP API when deleting domain ID %s",
			resp.StatusCode, domainID)
	}
	log.Printf("[DEBUG] ABP API delete domain %s was successful", domainID)

	return nil
}

func (c *Client) createABPDomain(val *ABPCreateDomain) (*ABPDomainResponse, error) {
	log.Printf("[INFO] Creating ABP domain")
	accountID, err := c.GetABPAccountID()
	if err != nil {
		return nil, err
	}

	valJSON, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON marshal ABP domain %v: %s", val, err)
	}
	log.Printf("[DEBUG] ABP create domain, about to send request with body: %s", valJSON)

	resp, err := c.DoJsonRequestWithHeaders(http.MethodPost,
		fmt.Sprintf("%s%s/%s/%s", c.config.BaseURLAPI, ABPAccountAPIPath, accountID, "domain"),
		valJSON,
		CreateABPDomain)
	if err != nil {
		return nil, fmt.Errorf("error from ABP API while creating domain under account ID %s: %s", accountID, err)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API POST domain data JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("Error status code %d from ABP API when creating domain under account ID %s: %s\n",
			resp.StatusCode, accountID, string(responseBody))
	}

	// Parse the JSON
	var createDomain ABPDomainResponse
	err = json.Unmarshal([]byte(responseBody), &createDomain)
	if err != nil {
		return nil, fmt.Errorf("Error parsing JSON response for creating ABP domain under account ID %s: %s\nresponse: %s\n",
			accountID, err, string(responseBody))
	}

	return &createDomain, nil
}
