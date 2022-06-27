package incapsula

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	ABPPolicyAPIPath = "/botmanagement/v1/policy"
)

type ABPPolicyDirective struct {
	Action      string  `json:"action"`
	ConditionID *string `json:"condition_id,omitempty"`
}

type ABPCreateUpdatePolicy struct {
	Name        string               `json:"name"`
	Description *string              `json:"description,omitempty"`
	Directives  []ABPPolicyDirective `json:"directives"`
}

type ABPPolicyResponse struct {
	ABPCreateUpdatePolicy
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
}

type ABPPoliciesResponse struct {
	Sites []ABPPolicyResponse `json:"items"`
}

func (c *Client) GetABPPolicy(policyID string) (*ABPPolicyResponse, error) {
	log.Printf("[INFO] Getting ABP policy ID: %s\n", policyID)

	if len(policyID) == 0 {
		return nil, fmt.Errorf("can't fetch empty policy ID")
	}

	resp, err := c.DoJsonRequestWithHeaders(http.MethodGet,
		fmt.Sprintf("%s%s/%s", c.config.BaseURLAPI, ABPPolicyAPIPath, policyID),
		nil,
		ReadABPPolicy)

	if err != nil {
		return nil, fmt.Errorf("error from ABP API when reading policy ID %s", policyID)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API read policy JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error status code %d from ABP API when reading policy ID %s: %s", resp.StatusCode, policyID, string(responseBody))
	}

	// Parse the JSON
	var abpPolicyResponse ABPPolicyResponse
	err = json.Unmarshal([]byte(responseBody), &abpPolicyResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON response for policy ID %s: %s. response: %s", policyID, err, string(responseBody))
	}

	return &abpPolicyResponse, nil
}

func (c *Client) GetABPPoliciesForAccountWithoutID() ([]ABPPolicyResponse, error) {
	if ID, err := c.GetABPAccountID(); err != nil {
		return nil, err
	} else {
		return c.GetABPPoliciesForAccount(ID)
	}
}

func (c *Client) GetABPPoliciesForAccount(accountID string) ([]ABPPolicyResponse, error) {
	log.Printf("[INFO] Getting ABP policies for account ID: %s\n", accountID)

	if len(accountID) == 0 {
		return nil, fmt.Errorf("can't fetch empty account ID")
	}

	resp, err := c.DoJsonRequestWithHeaders(http.MethodGet,
		fmt.Sprintf("%s%s/%s/%s", c.config.BaseURLAPI, ABPAccountAPIPath, accountID, "policy"),
		nil,
		ReadABPPolicy)

	if err != nil {
		return nil, fmt.Errorf("error from ABP API when reading policies for account ID %s", accountID)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API read policies JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error status code %d from ABP API when reading policies for account ID %s: %s", resp.StatusCode, accountID, string(responseBody))
	}

	// Parse the JSON
	var abpPoliciesResponse ABPPoliciesResponse
	err = json.Unmarshal([]byte(responseBody), &abpPoliciesResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON response for policies for account ID %s: %s. response: %s", accountID, err, string(responseBody))
	}

	return abpPoliciesResponse.Sites, nil
}

func (c *Client) updateABPPolicy(policyID string, val *ABPCreateUpdatePolicy) (*ABPPolicyResponse, error) {
	log.Printf("[INFO] Updating ABP policy ID: %s", policyID)

	valJSON, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON marshal ABP policy %v: %s", val, err)
	}
	log.Printf("[DEBUG] ABP update policy, about to send request with body: %s", valJSON)

	resp, err := c.DoJsonRequestWithHeaders(http.MethodPut,
		fmt.Sprintf("%s%s/%s", c.config.BaseURLAPI, ABPPolicyAPIPath, policyID),
		valJSON,
		UpdateABPPolicy)
	if err != nil {
		return nil, fmt.Errorf("error from ABP API while updating policy %s: %s", policyID, err)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API PUT policy data JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error status code %d from ABP API when updating policy %s: %s\n",
			resp.StatusCode, policyID, string(responseBody))
	}

	// Parse the JSON
	var updatedPolicy ABPPolicyResponse
	err = json.Unmarshal([]byte(responseBody), &updatedPolicy)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON response for ABP policy ID %s: %s\nresponse: %s\n",
			policyID, err, string(responseBody))
	}

	return &updatedPolicy, nil
}

func (c *Client) deleteABPPolicy(siteID string) error {
	log.Printf("[INFO] Deleting ABP policy %s", siteID)

	resp, err := c.DoJsonRequestWithHeaders(http.MethodDelete,
		fmt.Sprintf("%s%s/%s", c.config.BaseURLAPI, ABPPolicyAPIPath, siteID),
		nil,
		DeleteABPPolicy)
	if err != nil {
		return fmt.Errorf("error from ABP API while deleting policy ID %s: %s", siteID, err)
	}

	// Read the body
	defer resp.Body.Close()

	// Check the response code - no content for DELETE
	if resp.StatusCode != 200 {
		return fmt.Errorf("error status code %d from CSP API when deleting policy ID %s",
			resp.StatusCode, siteID)
	}
	log.Printf("[DEBUG] ABP API delete policy %s was successful", siteID)

	return nil
}

func (c *Client) createABPPolicy(val *ABPCreateUpdatePolicy) (*ABPPolicyResponse, error) {
	log.Printf("[INFO] Creating ABP policy %s", val.Name)
	accountID, err := c.GetABPAccountID()
	if err != nil {
		return nil, err
	}

	valJSON, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON marshal ABP policy %v: %s", val, err)
	}
	log.Printf("[DEBUG] ABP create policy, about to send request with body: %s", valJSON)

	resp, err := c.DoJsonRequestWithHeaders(http.MethodPost,
		fmt.Sprintf("%s%s/%s/%s", c.config.BaseURLAPI, ABPAccountAPIPath, accountID, "policy"),
		valJSON,
		CreateABPPolicy)
	if err != nil {
		return nil, fmt.Errorf("error from ABP API while creating policy under account ID %s: %s", accountID, err)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API POST policy data JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("Error status code %d from ABP API when creating policy under account ID %s: %s\n",
			resp.StatusCode, accountID, string(responseBody))
	}

	// Parse the JSON
	var createPolicy ABPPolicyResponse
	err = json.Unmarshal([]byte(responseBody), &createPolicy)
	if err != nil {
		return nil, fmt.Errorf("Error parsing JSON response for creating ABP policy under account ID %s: %s\nresponse: %s\n",
			accountID, err, string(responseBody))
	}

	return &createPolicy, nil
}
