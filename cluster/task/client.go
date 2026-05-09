package task

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sunshineOfficial/golib/goctx"
	"github.com/sunshineOfficial/golib/gohttp"
	"github.com/sunshineOfficial/golib/pagination"
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

func (c *Client) GetTasksByBrigade(ctx goctx.Context, brigadeID int, page pagination.Pagination) ([]Task, error) {
	var response []Task
	status, err := c.client.DoJson(ctx, http.MethodGet, fmt.Sprintf("%s/tasks/brigade/%d?%s", c.baseURL, brigadeID, pageQuery(page)), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("c.client.DoJson: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", status)
	}

	return response, nil
}

func pageQuery(page pagination.Pagination) string {
	values := make(url.Values)
	values.Set("limit", strconv.Itoa(page.Limit))
	values.Set("offset", strconv.Itoa(page.Offset))

	return values.Encode()
}
