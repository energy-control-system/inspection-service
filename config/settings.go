package config

type Settings struct {
	Port      int       `json:"port"`
	Databases Databases `json:"databases"`
	Cluster   Cluster   `json:"cluster"`
	Templates Templates `json:"templates"`
}

type Databases struct {
	Postgres string `json:"postgres"`
	Kafka    Kafka  `json:"kafka"`
}

type Kafka struct {
	Brokers []string `json:"brokers"`
	Topics  Topics   `json:"topics"`
}

type Topics struct {
	Inspections string `json:"inspections"`
	Tasks       string `json:"tasks"`
}

type Cluster struct {
	AnalyzerService   string `json:"analyzerService"`
	BrigadeService    string `json:"brigadeService"`
	FileService       string `json:"fileService"`
	SubscriberService string `json:"subscriberService"`
	TaskService       string `json:"taskService"`
}

type Templates struct {
	Universal string `json:"universal"`
	Control   string `json:"control"`
}
