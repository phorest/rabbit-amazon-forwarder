package lambda

import (
	"fmt"
	"errors"
	"encoding/json"

	"github.com/phorest/rabbit-amazon-forwarder/config"
	"github.com/phorest/rabbit-amazon-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	log "github.com/sirupsen/logrus"
)

const (
	// Type forwarder type
	Type = "Lambda"
)

// Forwarder forwarding client
type Forwarder struct {
	name         string
	lambdaClient lambdaiface.LambdaAPI
	function     string
	forwardHeaders bool
}

type Payload struct {
	Body string `json:"body"`
	Headers map[string] interface{} `json:"headers"`
}

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.AmazonEntry, options config.Options, lambdaClient ...lambdaiface.LambdaAPI) forwarder.Client {
	var client lambdaiface.LambdaAPI
	if len(lambdaClient) > 0 {
		client = lambdaClient[0]
	} else {
		client = lambda.New(session.Must(session.NewSession()))
	}

	forwarder := Forwarder{entry.Name, client, entry.Target, options.ForwardHeaders}
	log.WithField("forwarderName", forwarder.Name()).Info("Created forwarder")
	return forwarder
}

// Name forwarder name
func (f Forwarder) Name() string {
	return f.name
}

// Push pushes message to forwarding infrastructure
func (f Forwarder) Push(messageBody string, headers map[string]interface{}) error {
	if messageBody == "" {
		return errors.New(forwarder.EmptyMessageError)
	}

	messagePayload, err := f.buildPayload(messageBody, headers)
	if err != nil {
		log.WithFields(log.Fields{
			"forwarderName": f.Name(),
			"error":         err.Error()}).Error("Could not build message payload to push")
		return err
	}

	params := &lambda.InvokeInput{
		FunctionName: aws.String(f.function),
		Payload:      messagePayload,
	}

	resp, err := f.lambdaClient.Invoke(params)
	if err != nil {
		log.WithFields(log.Fields{
			"forwarderName": f.Name(),
			"error":         err.Error()}).Error("Could not forward message")
		return err
	}
	if resp.FunctionError != nil {
		log.WithFields(log.Fields{
			"forwarderName": f.Name(),
			"functionError": *resp.FunctionError}).Errorf("Could not forward message")
		return errors.New(*resp.FunctionError)
	}
	log.WithFields(log.Fields{
		"forwarderName": f.Name(),
		"statusCode":    resp.StatusCode}).Info("Forward succeeded")
	return nil
}

func (f Forwarder) buildPayload(messageBody string, headers map[string]interface{}) ([]byte, error) {
	if f.forwardHeaders {
		payload := Payload {
			Body: messageBody,
			Headers: headers,
		}
		messagePayload, err := json.Marshal(payload)
		if (err != nil) { return nil, err }
		return messagePayload, nil
	} else {
		return []byte(messageBody), nil
	}
}
