package task

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

func (c *Client) GetTaskByID(ctx goctx.Context, id int) (Task, error) {
	var response Task
	status, err := c.client.DoJson(ctx, http.MethodGet, fmt.Sprintf("%s/tasks/%d", c.baseURL, id), nil, &response)
	if err != nil {
		return Task{}, fmt.Errorf("c.client.DoJson: %w", err)
	}

	if status != http.StatusOK {
		return Task{}, fmt.Errorf("unexpected status code: %d", status)
	}

	return response, nil
}
