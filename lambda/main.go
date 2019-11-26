package main

import (
	"bytes"
	"encoding/base64"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/gommon/log"
	"os"
)

var BUCKET string

func main() {
	BUCKET = os.Getenv(`IMAGE_BUCKET`)
	lambda.Start(serveFunc)
}

func serveFunc(request events.APIGatewayProxyRequest) (resp events.APIGatewayProxyResponse, err error) {

	sess, err := session.NewSession()
	if err != nil {
		log.Error(err)
		return
	}

	s3Sdk :=  s3.New(sess)
	path := request.Path

	obj  , err := s3Sdk.GetObject(&s3.GetObjectInput{
		Bucket:                     &BUCKET,
		Key:                        &path,
	})
	if err != nil {
		log.Error(err)
		return
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(obj.Body)
	if err != nil {
		return
	}

	sEnc := base64.StdEncoding.EncodeToString(buf.Bytes())

	resp = events.APIGatewayProxyResponse{
		StatusCode:        200,
		Headers:           map[string]string{
			`Content-Type` : `image/jpeg`,
		},
		//MultiValueHeaders: nil,
		Body:              sEnc,
		IsBase64Encoded:   true,
	}

	return resp , nil
}