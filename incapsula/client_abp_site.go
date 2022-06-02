package incapsula

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	ABPSiteAPIPath = "/botmanagement/v1/site"
)

const (
	ABPSelectorCriteriaPostback   = "postback"
	ABPSelectorCriteriaPathPrefix = "path_prefix"
	ABPSelectorCriteriaPathRegex  = "path_regex"
)

type ABPCreateSiteSelector struct {
	PolicyID *string `json:"policy_id,omitempty"`
	Criteria struct {
		Postback   *string `json:"postback,omitempty"`
		PathPrefix *string `json:"path_prefix,omitempty"`
		PathRegex  *string `json:"path_regex,omitempty"`
	} `json:"criteria"`
	AnalysisSettings struct {
		RateLimiting string `json:"rate_limiting"`
	} `json:"analysis_settings"`
}

type ABPSiteSelector struct {
	ABPCreateSiteSelector
	DerivedID *string `json:"derived_id,omitempty"`
}

type ABPUpdateSite struct {
	Name      string            `json:"name"`
	Selectors []ABPSiteSelector `json:"selectors"`
}

type ABPCreateSite struct {
	ABPUpdateSite
	MxHostnameID *string `json:"mx_hostname_id,omitempty"`
}

type ABPSiteResponse struct {
	ABPCreateSite
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
}

type ABPSitesResponse struct {
	Sites []ABPSiteResponse `json:"items"`
}

func (c *Client) GetABPSite(siteID string) (*ABPSiteResponse, error) {
	log.Printf("[INFO] Getting ABP site ID: %s\n", siteID)

	if len(siteID) == 0 {
		return nil, fmt.Errorf("Cant' fetch empty site ID")
	}

	resp, err := c.DoJsonRequestWithHeaders(http.MethodGet,
		fmt.Sprintf("%s%s/%s", c.config.BaseURLAPI, ABPSiteAPIPath, siteID),
		nil,
		ReadABPSite)

	if err != nil {
		return nil, fmt.Errorf("Error from ABP API when reading site ID %s", siteID)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API read site JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error status code %d from ABP API when reading site ID %s: %s", resp.StatusCode, siteID, string(responseBody))
	}

	// Parse the JSON
	var abpSiteResponse ABPSiteResponse
	err = json.Unmarshal([]byte(responseBody), &abpSiteResponse)
	if err != nil {
		return nil, fmt.Errorf("Error parsing JSON response for site ID %s: %s. response: %s", siteID, err, string(responseBody))
	}

	return &abpSiteResponse, nil
}

func (c *Client) GetABPSitesForAccountWithoutID() ([]ABPSiteResponse, error) {
	if ID, err := c.GetABPAccountID(); err != nil {
		return nil, err
	} else {
		return c.GetABPSitesForAccount(ID)
	}
}

func (c *Client) GetABPSitesForAccount(accountID string) ([]ABPSiteResponse, error) {
	log.Printf("[INFO] Getting ABP sites for account ID: %s\n", accountID)

	if len(accountID) == 0 {
		return nil, fmt.Errorf("Cant' fetch empty account ID")
	}

	log.Printf("ABP get sites, fetching URL: %s%s/%s/%s", c.config.BaseURLAPI, ABPAccountAPIPath, accountID, "site")
	resp, err := c.DoJsonRequestWithHeaders(http.MethodGet,
		fmt.Sprintf("%s%s/%s/%s", c.config.BaseURLAPI, ABPAccountAPIPath, accountID, "site"),
		nil,
		ReadABPSite)

	if err != nil {
		return nil, fmt.Errorf("Error from ABP API when reading sites for account ID %s", accountID)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API read sites JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error status code %d from ABP API when reading sites for account ID %s: %s", resp.StatusCode, accountID, string(responseBody))
	}

	// Parse the JSON
	var abpSitesResponse ABPSitesResponse
	err = json.Unmarshal([]byte(responseBody), &abpSitesResponse)
	if err != nil {
		return nil, fmt.Errorf("Error parsing JSON response for sites for account ID %s: %s. response: %s", accountID, err, string(responseBody))
	}

	return abpSitesResponse.Sites, nil
}

func (c *Client) updateABPSite(siteID string, val *ABPUpdateSite) (*ABPSiteResponse, error) {
	log.Printf("[INFO] Updating ABP site ID: %s", siteID)

	valJSON, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON marshal ABP site %v: %s", val, err)
	}
	log.Printf("[DEBUG] ABP update site, about to send request with body: %s", valJSON)

	resp, err := c.DoJsonRequestWithHeaders(http.MethodPut,
		fmt.Sprintf("%s%s/%s", c.config.BaseURLAPI, ABPSiteAPIPath, siteID),
		valJSON,
		UpdateABPSite)
	if err != nil {
		return nil, fmt.Errorf("error from ABP API while updating site %s: %s", siteID, err)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API PUT site data JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error status code %d from ABP API when updating site %s: %s\n",
			resp.StatusCode, siteID, string(responseBody))
	}

	// Parse the JSON
	var updatedSite ABPSiteResponse
	err = json.Unmarshal([]byte(responseBody), &updatedSite)
	if err != nil {
		return nil, fmt.Errorf("Error parsing JSON response for ABP site ID %s: %s\nresponse: %s\n",
			siteID, err, string(responseBody))
	}

	return &updatedSite, nil
}

func (c *Client) deleteABPSite(siteID string) error {
	log.Printf("[INFO] Deleting ABP site %s", siteID)

	resp, err := c.DoJsonRequestWithHeaders(http.MethodDelete,
		fmt.Sprintf("%s%s/%s", c.config.BaseURLAPI, ABPSiteAPIPath, siteID),
		nil,
		DeleteABPSite)
	if err != nil {
		return fmt.Errorf("error from ABP API while deleting site ID %s: %s", siteID, err)
	}

	// Read the body
	defer resp.Body.Close()

	// Check the response code - no content for DELETE
	if resp.StatusCode != 200 {
		return fmt.Errorf("error status code %d from CSP API when deleting site ID %s",
			resp.StatusCode, siteID)
	}
	log.Printf("[DEBUG] ABP API delete site %s was successful", siteID)

	return nil
}

func (c *Client) createABPSite(val *ABPCreateSite) (*ABPSiteResponse, error) {
	log.Printf("[INFO] Creating ABP site %s", val.Name)
	accountID, err := c.GetABPAccountID()
	if err != nil {
		return nil, err
	}

	valJSON, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON marshal ABP site %v: %s", val, err)
	}
	log.Printf("[DEBUG] ABP create site, about to send request with body: %s", valJSON)

	resp, err := c.DoJsonRequestWithHeaders(http.MethodPost,
		fmt.Sprintf("%s%s/%s/%s", c.config.BaseURLAPI, ABPAccountAPIPath, accountID, "site"),
		valJSON,
		CreateABPSite)
	if err != nil {
		return nil, fmt.Errorf("error from ABP API while creating site under account ID %s: %s", accountID, err)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API POST site data JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("Error status code %d from ABP API when creating site under account ID %s: %s\n",
			resp.StatusCode, accountID, string(responseBody))
	}

	// Parse the JSON
	var updatedSite ABPSiteResponse
	err = json.Unmarshal([]byte(responseBody), &updatedSite)
	if err != nil {
		return nil, fmt.Errorf("Error parsing JSON response for creating ABP site under account ID %s: %s\nresponse: %s\n",
			accountID, err, string(responseBody))
	}

	return &updatedSite, nil
}
