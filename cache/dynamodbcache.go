package cache

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/ksanta/wordofthedaygame/model"
	"log"
)

type DynamoDbCache struct {
	svc *dynamodb.DynamoDB
}

func NewDynamoDbCache() Cache {
	config := &aws.Config{
		Region:   aws.String("ap-southeast-2"),
		Endpoint: aws.String("http://localhost:8000"),
	}
	sess := session.Must(session.NewSession(config))
	svc := dynamodb.New(sess)
	return &DynamoDbCache{svc}
}

func (d *DynamoDbCache) DoesNotExist() bool {
	input := &dynamodb.DescribeTableInput{TableName: aws.String("Words")}
	_, err := d.svc.DescribeTable(input)
	if err != nil {
		if err, ok := err.(awserr.Error); ok {
			if err.Code() == dynamodb.ErrCodeResourceNotFoundException {
				return true
			}
		}
		log.Fatal(err)
	}
	return false
}

func (d *DynamoDbCache) CreateCacheWriter() chan model.Word {
	wordChannel := make(chan model.Word)

	go func() {
		d.createTable()

		for word := range wordChannel {
			fmt.Println(word)
		}
	}()

	return wordChannel
}

func (d *DynamoDbCache) LoadWordsFromCache() model.Words {
	panic("implement me")
}

func (d *DynamoDbCache) createTable() {
	createTableInput := &dynamodb.CreateTableInput{
		TableName: aws.String("Words"),
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("word"),
				KeyType:       aws.String("HASH"),
			},
		},
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("word"),
				AttributeType: aws.String("S"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	}
	_, err := d.svc.CreateTable(createTableInput)
	if err != nil {
		panic(err)
	}

	waitUntilExistsInput := &dynamodb.DescribeTableInput{TableName: aws.String("Words")}
	err = d.svc.WaitUntilTableExists(waitUntilExistsInput)
	if err != nil {
		panic(err)
	}
}
