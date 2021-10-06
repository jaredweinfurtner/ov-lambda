// SPDX-License-Identifier: Unlicense

package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ses"
)

var (
	db     *dynamodb.DynamoDB
	email  *ses.SES
	config *Config
)

type Config struct {
	FromEmail    string
	ToEmail      string
	EmailCharset string
	Theater      string
	DbTableName  string
}

type Theater struct {
	ID    string `json:"id"`
	Url   string `json:"url"`
	Movie string `json:"movie"`
}

func init() {
	awsSession := session.Must(session.NewSession())
	db = dynamodb.New(awsSession)
	email = ses.New(awsSession)
	config = &Config{
		FromEmail:    os.Getenv("FROM_EMAIL"),
		ToEmail:      os.Getenv("TO_EMAIL"),
		EmailCharset: os.Getenv("EMAIL_CHARSET"),
		Theater:      os.Getenv("THEATER"),
		DbTableName:  os.Getenv("DB_TABLE_NAME"),
	}
}

func main() {
	lambda.Start(handler)
}

func handler() error {
	// get the theater from the db
	theater, err := getTheater(config.Theater)
	if err != nil {
		return err
	}

	// get the current OV movie from the theater's URL
	currentMovie, err := getCurrentMovie(theater)
	if err != nil {
		return err
	}

	// compare (do nothing if they're the same)
	if !strings.EqualFold(currentMovie, theater.Movie) {

		// update db
		theater.Movie = currentMovie
		err = updateTheater(theater)
		if err != nil {
			return err
		}

		// send notification
		err = sendNotification(theater)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateTheater(theater *Theater) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(config.DbTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(theater.ID),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":m": {
				S: aws.String(theater.Movie),
			},
		},
		UpdateExpression: aws.String("set movie = :m"),
	}

	_, err := db.UpdateItem(input)
	return err
}

func getTheater(id string) (*Theater, error) {
	theaterQuery := &dynamodb.GetItemInput{
		TableName: aws.String(config.DbTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}

	result, err := db.GetItem(theaterQuery)
	if err != nil {
		return nil, err
	}

	theater := new(Theater)
	err = dynamodbattribute.UnmarshalMap(result.Item, theater)
	return theater, err
}

func getCurrentMovie(theater *Theater) (string, error) {
	res, err := http.Get(theater.Url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	// make sure it connects
	if res.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("Could not retrieve movie from theater [%s]: %v", theater.ID, res.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	return doc.Find(".contentbox-body h3.headline-blue a").First().Text(), nil
}

func sendNotification(theater *Theater) error {
	emailInput := &ses.SendEmailInput{
		Source: aws.String(config.FromEmail),
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(config.ToEmail),
			},
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String(config.EmailCharset),
				Data:    aws.String(fmt.Sprintf("New OV Movie in %s: %s", theater.ID, theater.Movie)),
			},
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String(config.EmailCharset),
					Data:    aws.String(fmt.Sprintf("%s \n%s", theater.Movie, theater.Url)),
				},
			},
		},
	}

	_, err := email.SendEmail(emailInput)
	return err
}
