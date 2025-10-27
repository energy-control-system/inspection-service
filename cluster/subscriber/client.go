package subscriber

import (
	"fmt"
	"net/http"

	"github.com/sunshineOfficial/golib/goctx"
	"github.com/sunshineOfficial/golib/gohttp"
)

type Client struct {
	client  gohttp.Client
	baseURL string
}

func NewClient(client gohttp.Client, baseURL string) *Client {
	return &Client{
		client:  client,
		baseURL: baseURL,
	}
}

func (c *Client) GetLastContractByObjectID(ctx goctx.Context, objectID int) (Contract, error) {
	var response Contract
	status, err := c.client.DoJson(ctx, http.MethodGet, fmt.Sprintf("%s/contracts/objects/%d/last", c.baseURL, objectID), nil, &response)
	if err != nil {
		return Contract{}, fmt.Errorf("c.client.DoJson: %w", err)
	}

	if status != http.StatusOK {
		return Contract{}, fmt.Errorf("unexpected status code: %d", status)
	}

	return response, nil
}

func (c *Client) GetObjectByDeviceID(ctx goctx.Context, deviceID int) (Object, error) {
	var response Object
	status, err := c.client.DoJson(ctx, http.MethodGet, fmt.Sprintf("%s/objects/devices/%d", c.baseURL, deviceID), nil, &response)
	if err != nil {
		return Object{}, fmt.Errorf("c.client.DoJson: %w", err)
	}

	if status != http.StatusOK {
		return Object{}, fmt.Errorf("unexpected status code: %d", status)
	}

	return response, nil
}

func (c *Client) GetObjectBySealID(ctx goctx.Context, sealID int) (Object, error) {
	var response Object
	status, err := c.client.DoJson(ctx, http.MethodGet, fmt.Sprintf("%s/objects/seals/%d", c.baseURL, sealID), nil, &response)
	if err != nil {
		return Object{}, fmt.Errorf("c.client.DoJson: %w", err)
	}

	if status != http.StatusOK {
		return Object{}, fmt.Errorf("unexpected status code: %d", status)
	}

	return response, nil
}
