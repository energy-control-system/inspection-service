package analyzer

type ProcessImageResponse struct {
	IsBlurred    bool   `json:"IsBlurred"`
	BlurScore    string `json:"BlurScore"`
	QualityScore string `json:"QualityScore"`
	HasError     bool   `json:"HasError"`
	Filename     string `json:"Filename"`
	Dimensions   string `json:"Dimensions"`
	Channels     int    `json:"Channels"`
}
