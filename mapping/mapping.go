package mapping

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/phorest/rabbit-amazon-forwarder/connector"
	log "github.com/sirupsen/logrus"

	"github.com/phorest/rabbit-amazon-forwarder/config"
	"github.com/phorest/rabbit-amazon-forwarder/consumer"
	"github.com/phorest/rabbit-amazon-forwarder/forwarder"
	"github.com/phorest/rabbit-amazon-forwarder/lambda"
	"github.com/phorest/rabbit-amazon-forwarder/rabbitmq"
	"github.com/phorest/rabbit-amazon-forwarder/sns"
	"github.com/phorest/rabbit-amazon-forwarder/sqs"
)

type rules []ForwardingRule

type ForwardingRule struct { 
	Source      config.RabbitEntry `json:"source"`
	Destination config.AmazonEntry `json:"destination"`
	Options 	config.Options `json:"options"`
}

// Client mapping client
type Client struct {
	helper Helper
}

// Helper interface for creating consumers and forwaders
type Helper interface {
	createConsumer(entry config.RabbitEntry) consumer.Client
	createForwarder(entry config.AmazonEntry, options config.Options) forwarder.Client
}

// ConsumerForwarderMapping mapping for consumers and forwarders
type ConsumerForwarderMapping struct {
	Consumer  consumer.Client
	Forwarder forwarder.Client
}

type helperImpl struct{}

// New creates new mapping client
func New(helpers ...Helper) Client {
	var helper Helper
	helper = helperImpl{}
	if len(helpers) > 0 {
		helper = helpers[0]
	}
	return Client{helper}
}

// Load loads mappings
func (c Client) Load() ([]ConsumerForwarderMapping, error) {
	var consumerForwarderMapping []ConsumerForwarderMapping
	data, err := c.loadMappings()
	if err != nil {
		return consumerForwarderMapping, err
	}
	var rulesList rules
	if err = json.Unmarshal(data, &rulesList); err != nil {
		return consumerForwarderMapping, err
	}
	log.Info("Loading consumer - forwarder pairs")
	for _, rule := range rulesList {
		consumer := c.helper.createConsumer(rule.Source)
		forwarder := c.helper.createForwarder(rule.Destination, rule.Options)
		consumerForwarderMapping = append(consumerForwarderMapping, ConsumerForwarderMapping{consumer, forwarder})
	}
	return consumerForwarderMapping, nil
}

func (c Client) loadMappings() ([]byte, error) {
	filePath := os.Getenv(config.MappingFile)
	if filePath != "" {
		log.WithField("mappingFile", filePath).Info("Loading mapping file")
		return ioutil.ReadFile(filePath)
	} else {
		jsonConfig := os.Getenv(config.MappingJson)
		if jsonConfig == "" {
			return nil, errors.New("must provide either a mapping file or mapping json")
		}

		return []byte(jsonConfig), nil
	}
}

func (h helperImpl) createConsumer(entry config.RabbitEntry) consumer.Client {
	log.WithFields(log.Fields{
		"consumerType": entry.Type,
		"consumerName": entry.Name}).Info("Creating consumer")
	switch entry.Type {
	case rabbitmq.Type:
		rabbitConnector := connector.CreateConnector(entry.ConnectionURL)
		return rabbitmq.CreateConsumer(entry, rabbitConnector)
	}
	return nil
}

func (h helperImpl) createForwarder(entry config.AmazonEntry, options config.Options) forwarder.Client {
	log.WithFields(log.Fields{
		"forwarderType": entry.Type,
		"forwarderName": entry.Name}).Info("Creating forwarder")
	switch entry.Type {
	case sns.Type:
		return sns.CreateForwarder(entry)
	case sqs.Type:
		return sqs.CreateForwarder(entry)
	case lambda.Type:
		return lambda.CreateForwarder(entry, options)
	}
	return nil
}
