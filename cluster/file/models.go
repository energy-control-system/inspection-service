package file

type File struct {
	ID       int    `json:"ID"`
	FileName string `json:"FileName"`
	FileSize int64  `json:"FileSize"`
	Bucket   Bucket `json:"Bucket"`
	URL      string `json:"URL"`
}

type Bucket string

const (
	BucketImages    Bucket = "images"
	BucketDocuments Bucket = "documents"
)
