package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const azureRestPath = "https://management.azure.com/"

type Azure struct {
	TenantID     string `json:"tenantId"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`

	Path          string `json:"path"`
	ResourceDelta int    `json:"resourceDelta"`
	Action        string `json:"action"`

	// +optional
	Token string `json:"token,omitempty"`
}

func (c *Azure) Connect() error {
	urlToken := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", c.TenantID)

	// Set up the form values
	data := url.Values{}
	data.Set("client_id", c.ClientID)
	data.Set("resource", "https://management.azure.com/")
	data.Set("client_secret", c.ClientSecret)
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", urlToken, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read and parse the response
	responseBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(responseBody, &result)

	c.Token = result["access_token"].(string)
	return nil
}

func (c *Azure) SetResourceStatus() error {
	url := strings.TrimSuffix(azureRestPath, "/") + strings.TrimSuffix(c.Path, "/")

	var verb string
	var body []byte

	switch c.Action {
	case "scale-up":
		size, location, err := c.getVMSize(url, "up")
		if err != nil {
			return err
		}
		fmt.Printf("{\"properties\": {\"hardwareProfile\": {\"vmSize\": \"%s\"}},\"location\":\"%s\"}", size, location)
		body = []byte(fmt.Sprintf("{\"properties\": {\"hardwareProfile\": {\"vmSize\": \"%s\"}},\"location\":\"%s\"}", size, location))
		verb = "PATCH"
	case "scale-down":
		verb = "PATCH"
		size, location, err := c.getVMSize(url, "down")
		if err != nil {
			return err
		}
		body = []byte(fmt.Sprintf("{\"properties\": {\"hardwareProfile\": {\"vmSize\": \"%s\"}},\"location\":\"%s\"}", size, location))
	default:
		verb = "POST"
		url += "/" + c.Action
	}

	var req *http.Request
	var err error
	if len(body) > 0 {
		req, err = c.newRequest(verb, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = c.newRequest(verb, url, nil)
	}
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

func (c *Azure) getVMSize(Url string, direction string) (string, string, error) {
	// Get the current VM size
	url := Url
	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", "", err
	}

	// Parse current vm size
	responseBody, _ := io.ReadAll(resp.Body)
	var vmConfigAzure VMConfigAzure
	json.Unmarshal(responseBody, &vmConfigAzure)

	resp.Body.Close()

	// Get the next VM size (up or down, depending on parameter)
	url = Url + "/vmSizes"
	req, err = c.newRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}

	resp, err = c.doRequest(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	responseBody, _ = io.ReadAll(resp.Body)
	var vmSizes VMSizes
	json.Unmarshal(responseBody, &vmSizes)

	for i, vmSize := range vmSizes.Value {
		if vmSize.Name == vmConfigAzure.Properties.HardwareProfile.VMSize {
			switch direction {
			case "up":
				for j := i + 1; j < len(vmSizes.Value); j++ {
					if vmSizes.Value[j].NumberOfCores-vmSize.NumberOfCores >= int(vmSize.NumberOfCores*(c.ResourceDelta/100)) {
						fmt.Println(vmSizes.Value[j].Name)
						return vmSizes.Value[j].Name, vmConfigAzure.Properties.Location, nil
					}
				}
			case "down":
				for j := i - 1; j > 0; j-- {
					if vmSizes.Value[j].NumberOfCores-vmSize.NumberOfCores <= int(vmSize.NumberOfCores*(c.ResourceDelta/100)) {
						fmt.Println(vmSizes.Value[j].Name)
						return vmSizes.Value[j].Name, vmConfigAzure.Properties.Location, nil
					}
				}
			}
		}
	}
	return vmConfigAzure.Properties.HardwareProfile.VMSize, vmConfigAzure.Properties.Location, nil
}

func (c *Azure) newRequest(Verb string, Url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(Verb, Url, body)
	if err != nil {
		return nil, err
	}
	data := req.URL.Query()
	data.Add("api-version", "2024-03-01")
	req.URL.RawQuery = data.Encode()
	req.Header.Add("Authorization", "Bearer "+c.Token)

	return req, nil
}

func (c *Azure) doRequest(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		return nil, errors.New("CloudError: " + strconv.FormatInt(int64(resp.StatusCode), 10) + " - Error: " + resp.Status)
	}
	return resp, nil
}

type VMConfigAzure struct {
	Properties Properties `json:"properties"`
}

type Properties struct {
	HardwareProfile HardwareProfile `json:"hardwareProfile"`
	Location        string          `json:"location"`
}

type HardwareProfile struct {
	VMSize string `json:"vmSize"`
}

type VMSizes struct {
	Value []VMSize `json:"value"`
}

type VMSize struct {
	Name          string `json:"name"`
	NumberOfCores int    `json:"numberOfCores"`
	MemoryInMB    int    `json:"memoryInMB"`
}
