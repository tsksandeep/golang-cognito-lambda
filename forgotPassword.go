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

type UserForgot struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

type UserConfirmForgot struct {
	ConfirmationCode string `json:"confirmationCode"`
	Email            string `json:"email"`
	PhoneNumber      string `json:"phoneNumber"`
	Password         string `json:"password"`
}

func ForgotPassword(ctx context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var userForgot UserForgot
	err := json.Unmarshal([]byte(request.Body), &userForgot)
	if err != nil {
		log.Error("User Forgot: ", err)
		return writeGatewayProxyResponse("", ErrBadRequest)
	}

	cognitoUsername := getCognitoUsername(userForgot.Email, userForgot.PhoneNumber)

	input := &cognito.ForgotPasswordInput{
		ClientId:   aws.String(appClientID),
		Username:   aws.String(cognitoUsername),
		SecretHash: aws.String(cognitoSecretHash(cognitoUsername)),
	}

	res, err := cognitoClient.ForgotPassword(input)
	if awsErr, ok := err.(awserr.Error); ok {
		log.Error("User Login: ", awsErr)

		switch awsErr.Code() {
		case cognito.ErrCodeUserNotFoundException:
			return writeGatewayProxyResponse("User not found", ErrBadRequest)
		default:
			return writeGatewayProxyResponse("", ErrInternalServerError)
		}
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		log.Error("User Forgot: ", err)
		return writeGatewayProxyResponse("", ErrInternalServerError)
	}

	return writeGatewayProxyResponse(string(resBytes), nil)
}

func ConfirmForgotPassword(ctx context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var userConfirmForgot UserConfirmForgot
	err := json.Unmarshal([]byte(request.Body), &userConfirmForgot)
	if err != nil {
		log.Error("User Confirm Forgot: ", err)
		return writeGatewayProxyResponse("", ErrBadRequest)
	}

	cognitoUsername := getCognitoUsername(userConfirmForgot.Email, userConfirmForgot.PhoneNumber)

	input := &cognito.ConfirmForgotPasswordInput{
		ClientId:         aws.String(appClientID),
		Username:         aws.String(cognitoUsername),
		Password:         aws.String(userConfirmForgot.Password),
		ConfirmationCode: aws.String(userConfirmForgot.ConfirmationCode),
		SecretHash:       aws.String(cognitoSecretHash(cognitoUsername)),
	}

	_, err = cognitoClient.ConfirmForgotPassword(input)
	if awsErr, ok := err.(awserr.Error); ok {
		log.Error("User Confirm Forgot: ", awsErr)

		switch awsErr.Code() {
		case cognito.ErrCodeUserNotFoundException:
			return writeGatewayProxyResponse("User not found", ErrBadRequest)
		case cognito.ErrCodeCodeMismatchException:
			return writeGatewayProxyResponse("Invalid OTP", ErrBadRequest)
		case cognito.ErrCodeLimitExceededException:
			return writeGatewayProxyResponse("Maximum OTP limit reached", ErrBadRequest)
		case cognito.ErrCodeExpiredCodeException:
			return writeGatewayProxyResponse("OTP expired", ErrBadRequest)
		default:
			return writeGatewayProxyResponse("", ErrInternalServerError)
		}
	}

	return writeGatewayProxyResponse("Password reset successful", nil)
}
