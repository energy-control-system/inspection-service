package analyzer

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

func (c *Client) ProcessImage(ctx goctx.Context, fileName string, image io.Reader) (ProcessImageResponse, error) {
	rq, err := gohttp.NewRequest(ctx, http.MethodPost, c.baseURL+"/process-image", nil)
	if err != nil {
		return ProcessImageResponse{}, fmt.Errorf("NewRequest: %w", err)
	}

	data := &gohttp.MultipartData{
		Files: []gohttp.MultipartFile{{
			FieldName: "file",
			FileName:  fileName,
			Reader:    image,
		}},
	}

	if err = gohttp.WriteRequestMultipart(rq, data); err != nil {
		return ProcessImageResponse{}, fmt.Errorf("WriteRequestMultipart: %w", err)
	}

	rs, err := c.client.Do(rq)
	if err != nil {
		if rs != nil && rs.Body != nil {
			closeErr := rs.Body.Close()
			err = errors.Join(err, closeErr)
		}

		return ProcessImageResponse{}, fmt.Errorf("c.client.Do: %w", err)
	}

	if rs == nil {
		return ProcessImageResponse{}, errors.New("got nil response from server")
	}

	if rs.StatusCode != http.StatusOK {
		if rs.Body != nil {
			closeErr := rs.Body.Close()
			err = errors.Join(err, closeErr)
		}

		return ProcessImageResponse{}, errors.Join(err, fmt.Errorf("got status code %d", rs.StatusCode))
	}

	var response ProcessImageResponse
	if err = gohttp.ReadResponseJson(rs, &response); err != nil {
		return ProcessImageResponse{}, fmt.Errorf("ReadResponseJson: %w", err)
	}

	return response, nil
}
