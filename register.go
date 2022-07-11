package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	log "github.com/sirupsen/logrus"
)

type UserRegister struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phoneNumber"`
}

func Register(ctx context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var userRegister UserRegister
	err := json.Unmarshal([]byte(request.Body), &userRegister)
	if err != nil {
		log.Error("User Register: ", err)
		return writeGatewayProxyResponse("", ErrBadRequest)
	}

	err = checkRegister(userRegister)
	if err != nil {
		log.Error("User Register: ", err)
		return writeGatewayProxyResponse("", ErrBadRequest)
	}

	cognitoUsername := getCognitoUsername(userRegister.Email, userRegister.PhoneNumber)

	input := &cognito.SignUpInput{
		Username:   aws.String(cognitoUsername),
		Password:   aws.String(userRegister.Password),
		ClientId:   aws.String(appClientID),
		SecretHash: aws.String(cognitoSecretHash(cognitoUsername)),
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("name"),
				Value: aws.String(userRegister.Username),
			},
			{
				Name:  aws.String("phone_number"),
				Value: aws.String(userRegister.PhoneNumber),
			},
			{
				Name:  aws.String("email"),
				Value: aws.String(userRegister.Email),
			},
		},
	}

	res, err := cognitoClient.SignUp(input)
	if awsErr, ok := err.(awserr.Error); ok {
		log.Error("User Register: ", awsErr)

		switch awsErr.Code() {
		case cognito.ErrCodeUsernameExistsException:
			return writeGatewayProxyResponse("User already exists", ErrBadRequest)
		default:
			return writeGatewayProxyResponse("", ErrInternalServerError)
		}
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		log.Error("User Register: ", err)
		return writeGatewayProxyResponse("", ErrInternalServerError)
	}

	return writeGatewayProxyResponse(string(resBytes), nil)
}
