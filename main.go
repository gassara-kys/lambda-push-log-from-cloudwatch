package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	lambda.Start(handler)
}

type lambdaConfig struct {
	Description string `default:"Some log messages were detected in CloudWatchLogs(subscription filter)."` // DESCRIPTION
	SNSTopicArn string `required:"true" split_words:"true"`                                                // SNS_TOPIC_ARN
}

func handler(ctx context.Context, e events.CloudwatchLogsEvent) error {
	var conf lambdaConfig
	err := envconfig.Process("", &conf)
	if err != nil {
		log.Fatalf("Failed to load enviroment variables, err=%+v", err)
	}

	msgs, err := parseMsg(e.AWSLogs)
	if err != nil {
		log.Fatalf("Failed to parse message, err=%+v", err)
	}
	if err := publishSNS(conf.SNSTopicArn, conf.Description, msgs); err != nil {
		log.Fatalf("Failed to publish SNS message, err=%+v", err)
	}
	return nil
}

func parseMsg(awsLogs events.CloudwatchLogsRawData) (*[]string, error) {
	msgs := []string{}
	data, err := awsLogs.Parse()
	if err != nil {
		return nil, err
	}

	for _, logEvent := range data.LogEvents {
		msgs = append(msgs, logEvent.Message)
	}
	return &msgs, nil
}

// type messageParam struct {
// 	Description string
// 	Logs        string
// }

// var messageTemplate = `
// Description:
// {{.Description}}

// Logs:
// {{.Logs}}
// `

func publishSNS(topicArn, desc string, logs *[]string) error {
	// tmpl, err := template.New("").Parse(messageTemplate)
	// if err != nil {
	// 	return err
	// }

	msg := fmt.Sprintln("Description:")
	msg += fmt.Sprintln(desc)
	msg += fmt.Sprintln()
	msg += fmt.Sprintln("Logs:")
	for _, l := range *logs {
		msg += fmt.Sprintln(l)
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sns.New(sess)
	result, err := svc.Publish(&sns.PublishInput{
		Message:  aws.String(msg),
		TopicArn: aws.String(topicArn),
	})
	if err != nil {
		return err
	}
	log.Printf("Succeded publish SNS message: %+v", result)
	return nil
}
