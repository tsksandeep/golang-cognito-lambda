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

type UserOTP struct {
	OTP         string `json:"otp"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

func OTP(ctx context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var userOTP UserOTP
	err := json.Unmarshal([]byte(request.Body), &userOTP)
	if err != nil {
		log.Error("User OTP: ", err)
		return writeGatewayProxyResponse("", ErrBadRequest)
	}

	cognitoUsername := getCognitoUsername(userOTP.Email, userOTP.PhoneNumber)

	input := &cognito.ConfirmSignUpInput{
		ConfirmationCode: aws.String(userOTP.OTP),
		Username:         aws.String(cognitoUsername),
		ClientId:         aws.String(appClientID),
		SecretHash:       aws.String(cognitoSecretHash(cognitoUsername)),
	}

	_, err = cognitoClient.ConfirmSignUp(input)
	if awsErr, ok := err.(awserr.Error); ok {
		log.Error("User OTP: ", awsErr)

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

	return writeGatewayProxyResponse("OTP successful", nil)
}
