package file

import "net/http"

const (
	ForwardedHostHeader  = "X-Forwarded-Host"
	ForwardedProtoHeader = "X-Forwarded-Proto"
)

type File struct {
	ID       int    `json:"ID"`
	FileName string `json:"FileName"`
	FileSize int64  `json:"FileSize"`
	Bucket   Bucket `json:"Bucket"`
	URL      string `json:"URL"`
}

type ForwardedHeaders struct {
	Host  string
	Proto string
}

func NewForwardedHeaders(r *http.Request) ForwardedHeaders {
	if r == nil {
		return ForwardedHeaders{}
	}

	return ForwardedHeaders{
		Host:  r.Header.Get(ForwardedHostHeader),
		Proto: r.Header.Get(ForwardedProtoHeader),
	}
}

func (h ForwardedHeaders) Apply(r *http.Request) {
	if h.Host != "" {
		r.Header.Set(ForwardedHostHeader, h.Host)
	}
	if h.Proto != "" {
		r.Header.Set(ForwardedProtoHeader, h.Proto)
	}
}

type Bucket string

const (
	BucketImages    Bucket = "images"
	BucketDocuments Bucket = "documents"
)
