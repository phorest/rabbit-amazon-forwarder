package config

const (
	// MappingFile mapping file environment variable
	MappingFile = "MAPPING_FILE"
	MappingJson = "MAPPING_JSON"
	CaCertFile  = "CA_CERT_FILE"
	CertFile    = "CERT_FILE"
	KeyFile     = "KEY_FILE"
)

// RabbitEntry RabbitMQ mapping entry
type RabbitEntry struct {
	Type          string   `json:"type"`
	Name          string   `json:"name"`
	ConnectionURL string   `json:"connection"`
	ExchangeName  string   `json:"topic"`
	ExchangeType  string   `json:"exchangeType"`
	QueueName     string   `json:"queue"`
	RoutingKey    string   `json:"routing"`
	RoutingKeys   []string `json:"routingKeys"`
}

// AmazonEntry SQS/SNS mapping entry
type AmazonEntry struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Target string `json:"target"`
}

type Options struct {
	ForwardHeaders bool `json:"forwardHeaders"`
}
