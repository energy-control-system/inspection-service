package brigade

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

func (c *Client) GetBrigadeByID(ctx goctx.Context, id int) (Brigade, error) {
	var response Brigade
	status, err := c.client.DoJson(ctx, http.MethodGet, fmt.Sprintf("%s/brigades/%d", c.baseURL, id), nil, &response)
	if err != nil {
		return Brigade{}, fmt.Errorf("c.client.DoJson: %w", err)
	}

	if status != http.StatusOK {
		return Brigade{}, fmt.Errorf("unexpected status code: %d", status)
	}

	return response, nil
}
