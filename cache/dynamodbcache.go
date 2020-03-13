package cache

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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

func (d *DynamoDbCache) SetupRequired() bool {
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
			marshalMap, err := dynamodbattribute.MarshalMap(word)
			if err != nil {
				panic(err)
			}
			input := &dynamodb.PutItemInput{
				TableName: aws.String("Words"),
				Item:      marshalMap,
			}
			_, err = d.svc.PutItem(input)
			if err != nil {
				panic(err)
			}
		}
	}()

	return wordChannel
}

func (d *DynamoDbCache) LoadWordsFromCache() model.Words {
	var words model.Words
	input := &dynamodb.ScanInput{
		TableName: aws.String("Words")}
	output, err := d.svc.Scan(input)
	if err != nil {
		panic(err)
	}

	for _, item := range output.Items {
		word := model.Word{}
		err := dynamodbattribute.UnmarshalMap(item, &word)
		if err != nil {
			panic(err)
		}
		words = append(words, word)
	}

	return words
}

func (d *DynamoDbCache) createTable() {
	createTableInput := &dynamodb.CreateTableInput{
		TableName: aws.String("Words"),
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("Word"),
				KeyType:       aws.String("HASH"),
			},
		},
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("Word"),
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
