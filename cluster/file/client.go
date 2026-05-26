package file

import (
	"errors"
	"fmt"
	"io"
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

func (c *Client) Upload(ctx goctx.Context, fileName string, file io.Reader) (File, error) {
	rq, err := gohttp.NewRequest(ctx, http.MethodPost, c.baseURL+"/files", nil)
	if err != nil {
		return File{}, fmt.Errorf("NewRequest: %w", err)
	}

	data := &gohttp.MultipartData{
		Files: []gohttp.MultipartFile{{
			FieldName: "File",
			FileName:  fileName,
			Reader:    file,
		}},
	}

	if err = gohttp.WriteRequestMultipart(rq, data); err != nil {
		return File{}, fmt.Errorf("WriteRequestMultipart: %w", err)
	}

	rs, err := c.client.Do(rq)
	if err != nil {
		if rs != nil && rs.Body != nil {
			closeErr := rs.Body.Close()
			err = errors.Join(err, closeErr)
		}

		return File{}, fmt.Errorf("c.client.Do: %w", err)
	}

	if rs == nil {
		return File{}, errors.New("got nil response from server")
	}

	if rs.StatusCode != http.StatusOK {
		if rs.Body != nil {
			closeErr := rs.Body.Close()
			err = errors.Join(err, closeErr)
		}

		return File{}, errors.Join(err, fmt.Errorf("got status code %d", rs.StatusCode))
	}

	var response File
	if err = gohttp.ReadResponseJson(rs, &response); err != nil {
		return File{}, fmt.Errorf("ReadResponseJson: %w", err)
	}

	return response, nil
}

func (c *Client) GetByIDs(ctx goctx.Context, ids []int, page pagination.Pagination) ([]File, error) {
	var response []File
	status, err := c.client.DoJson(ctx, http.MethodGet, fmt.Sprintf("%s/files?%s", c.baseURL, filesQuery(ids, page)), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("c.client.DoJson: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", status)
	}

	return response, nil
}

func filesQuery(ids []int, page pagination.Pagination) string {
	values := make(url.Values)
	for _, id := range ids {
		values.Add("id", strconv.Itoa(id))
	}
	values.Set("limit", strconv.Itoa(page.Limit))
	values.Set("offset", strconv.Itoa(page.Offset))

	return values.Encode()
}
