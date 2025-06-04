package snsnotifier

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
)

// NotificationMessage represents a message to be sent via SNS
type NotificationMessage struct {
	TopicArn   string            `json:"topic_arn"`
	Message    string            `json:"message"`
	Subject    string            `json:"subject,omitempty"`
	GroupID    string            `json:"group_id,omitempty"`   // For FIFO topics
	DedupeID   string            `json:"dedupe_id,omitempty"`  // For FIFO topics
	Attributes map[string]string `json:"attributes,omitempty"` // Message attributes for filtering
}

// Notifier handles SNS operations similar to how MQTTSubscriber handles MQTT
type Notifier struct {
	snsClient *sns.Client
	ctx       context.Context
}

// NewNotifier creates a new SNS notifier with AWS credentials loaded from environment/config
func NewNotifier(ctx context.Context) (*Notifier, error) {
	// Load AWS configuration from environment variables, credentials file, or IAM roles
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	snsClient := sns.NewFromConfig(cfg)

	notifier := &Notifier{
		snsClient: snsClient,
		ctx:       ctx,
	}

	log.Println("SNS Notifier initialized successfully")
	return notifier, nil
}

// NewNotifierWithConfig creates a new SNS notifier with custom AWS configuration
func NewNotifierWithConfig(ctx context.Context, awsConfig aws.Config) *Notifier {
	snsClient := sns.NewFromConfig(awsConfig)

	notifier := &Notifier{
		snsClient: snsClient,
		ctx:       ctx,
	}

	log.Println("SNS Notifier initialized with custom config")
	return notifier
}

// PublishMessage sends a notification message to an SNS topic
func (n *Notifier) PublishMessage(msg NotificationMessage) error {
	publishInput := &sns.PublishInput{
		TopicArn: aws.String(msg.TopicArn),
		Message:  aws.String(msg.Message),
	}

	// Add optional subject
	if msg.Subject != "" {
		publishInput.Subject = aws.String(msg.Subject)
	}

	// Add FIFO-specific attributes
	if msg.GroupID != "" {
		publishInput.MessageGroupId = aws.String(msg.GroupID)
	}
	if msg.DedupeID != "" {
		publishInput.MessageDeduplicationId = aws.String(msg.DedupeID)
	}

	// Add message attributes for filtering
	if len(msg.Attributes) > 0 {
		messageAttributes := make(map[string]types.MessageAttributeValue)
		for key, value := range msg.Attributes {
			messageAttributes[key] = types.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(value),
			}
		}
		publishInput.MessageAttributes = messageAttributes
	}

	result, err := n.snsClient.Publish(n.ctx, publishInput)
	if err != nil {
		log.Printf("Failed to publish message to topic %s: %v", msg.TopicArn, err)
		return fmt.Errorf("failed to publish SNS message: %w", err)
	}

	log.Printf("Successfully published message to topic %s, MessageId: %s", msg.TopicArn, *result.MessageId)
	return nil
}

// PublishSimple is a convenience method for sending simple text messages
func (n *Notifier) PublishSimple(topicArn, message string) error {
	return n.PublishMessage(NotificationMessage{
		TopicArn: topicArn,
		Message:  message,
	})
}

// PublishWithSubject is a convenience method for sending messages with a subject
func (n *Notifier) PublishWithSubject(topicArn, subject, message string) error {
	return n.PublishMessage(NotificationMessage{
		TopicArn: topicArn,
		Message:  message,
		Subject:  subject,
	})
}

// PublishFIFO is a convenience method for sending messages to FIFO topics
func (n *Notifier) PublishFIFO(topicArn, message, groupID, dedupeID string) error {
	return n.PublishMessage(NotificationMessage{
		TopicArn: topicArn,
		Message:  message,
		GroupID:  groupID,
		DedupeID: dedupeID,
	})
}

// PublishWithFilter is a convenience method for sending messages with filtering attributes
func (n *Notifier) PublishWithFilter(topicArn, message string, attributes map[string]string) error {
	return n.PublishMessage(NotificationMessage{
		TopicArn:   topicArn,
		Message:    message,
		Attributes: attributes,
	})
}

// ListTopics returns all SNS topics in the account
func (n *Notifier) ListTopics() ([]string, error) {
	var topicArns []string

	paginator := sns.NewListTopicsPaginator(n.snsClient, &sns.ListTopicsInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(n.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list topics: %w", err)
		}

		for _, topic := range output.Topics {
			if topic.TopicArn != nil {
				topicArns = append(topicArns, *topic.TopicArn)
			}
		}
	}

	return topicArns, nil
}

// CreateTopic creates a new SNS topic and returns its ARN
func (n *Notifier) CreateTopic(topicName string, isFIFO bool) (string, error) {
	input := &sns.CreateTopicInput{
		Name: aws.String(topicName),
	}

	// Add FIFO attributes if needed
	if isFIFO {
		input.Attributes = map[string]string{
			"FifoTopic": "true",
		}
		// Ensure FIFO topic name ends with .fifo
		if len(topicName) < 5 || topicName[len(topicName)-5:] != ".fifo" {
			input.Name = aws.String(topicName + ".fifo")
		}
	}

	result, err := n.snsClient.CreateTopic(n.ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create topic %s: %w", topicName, err)
	}

	log.Printf("Successfully created topic: %s", *result.TopicArn)
	return *result.TopicArn, nil
}

// Close gracefully shuts down the notifier (implements similar pattern to MQTT subscriber's closeFn)
func (n *Notifier) Close() {
	log.Println("SNS Notifier closed gracefully")
	// SNS client doesn't require explicit closing, but this provides consistency with MQTT pattern
}

// GetTopicArnFromEnv is a helper function to get topic ARN from environment variable
func GetTopicArnFromEnv(envVarName string) (string, error) {
	topicArn := os.Getenv(envVarName)
	if topicArn == "" {
		return "", fmt.Errorf("environment variable %s is not set or empty", envVarName)
	}
	return topicArn, nil
}
