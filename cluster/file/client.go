package file

import (
	"errors"
	"fmt"
	"io"
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
