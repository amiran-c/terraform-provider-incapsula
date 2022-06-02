package incapsula

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	ABPRootAPIPath    = "/botmanagement/v1/"
	ABPAccountAPIPath = "/botmanagement/v1/account"
)

type ABPRootResponse struct {
	AccountID string `json:"account_id"`
}

type ABPAccountResponse struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	MYAccountID string `json:"my_account_id"`
}

func (c *Client) GetABPAccountWithoutID() (*ABPAccountResponse, error) {
	if ID, err := c.GetABPAccountID(); err != nil {
		return c.GetABPAccount(ID)
	} else {
		return nil, err
	}
}

func (c *Client) GetABPAccountID() (string, error) {
	resp, err := c.DoJsonRequestWithHeaders(http.MethodGet,
		fmt.Sprintf("%s%s", c.config.BaseURLAPI, ABPRootAPIPath),
		nil,
		ReadABPAccount)

	if err != nil {
		return "", fmt.Errorf("Error from ABP API when reading root API")
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API Read Account JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error status code %d from ABP API when reading root API: %s", resp.StatusCode, string(responseBody))
	}

	// Parse the JSON
	var abpRootResponse ABPRootResponse
	err = json.Unmarshal([]byte(responseBody), &abpRootResponse)
	if err != nil {
		return "", fmt.Errorf("Error parsing JSON response for root API %s. response: %s", err, string(responseBody))
	}

	return abpRootResponse.AccountID, nil
}

func (c *Client) GetABPAccount(accountID string) (*ABPAccountResponse, error) {
	log.Printf("[INFO] Getting ABP account ID: %s\n", accountID)

	if len(accountID) == 0 {
		return nil, fmt.Errorf("Cant' fetch empty account ID")
	}

	resp, err := c.DoJsonRequestWithHeaders(http.MethodGet,
		fmt.Sprintf("%s%s/%s", c.config.BaseURLAPI, ABPAccountAPIPath, accountID),
		nil,
		ReadABPAccount)

	if err != nil {
		return nil, fmt.Errorf("Error from ABP API when reading account ID %s", accountID)
	}

	// Read the body
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)

	// Dump JSON
	log.Printf("[DEBUG] ABP API Read Account JSON response: %s\n", string(responseBody))

	// Check the response code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error status code %d from ABP API when reading account ID %s: %s", resp.StatusCode, accountID, string(responseBody))
	}

	// Parse the JSON
	var abpAccountResponse ABPAccountResponse
	err = json.Unmarshal([]byte(responseBody), &abpAccountResponse)
	if err != nil {
		return nil, fmt.Errorf("Error parsing JSON response for account ID %s: %s. response: %s", accountID, err, string(responseBody))
	}

	return &abpAccountResponse, nil
}
