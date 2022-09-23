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

type UserLogin struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	PhoneNumber  string `json:"phoneNumber"`
	UserId       string `json:"userId"`
	RefreshToken string `json:"refreshToken"`
}

type LoginResponse struct {
	UserId             string                      `json:"userId"`
	AuthenticationInfo *cognito.InitiateAuthOutput `json:"authenticationInfo"`
}

func Login(ctx context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	var userLogin UserLogin
	err := json.Unmarshal([]byte(request.Body), &userLogin)
	if err != nil {
		log.Error("User Login: ", err)
		return writeGatewayProxyResponse("", ErrBadRequest)
	}

	cognitoUsername := getCognitoUsername(userLogin.Email, userLogin.PhoneNumber)

	input := &cognito.InitiateAuthInput{
		ClientId: aws.String(appClientID),
	}

	if userLogin.UserId != "" && userLogin.RefreshToken != "" {
		input.AuthFlow = aws.String("REFRESH_TOKEN_AUTH")
		input.AuthParameters = map[string]*string{
			"REFRESH_TOKEN": aws.String(userLogin.RefreshToken),
			"SECRET_HASH":   aws.String(cognitoSecretHash(userLogin.UserId)),
		}
	} else {
		input.AuthFlow = aws.String("USER_PASSWORD_AUTH")
		input.AuthParameters = map[string]*string{
			"USERNAME":    aws.String(cognitoUsername),
			"PASSWORD":    aws.String(userLogin.Password),
			"SECRET_HASH": aws.String(cognitoSecretHash(cognitoUsername)),
		}

	}

	authOutput, err := cognitoClient.InitiateAuth(input)
	if awsErr, ok := err.(awserr.Error); ok {
		log.Error("User Login: ", awsErr)

		switch awsErr.Code() {
		case cognito.ErrCodeNotAuthorizedException:
			return writeGatewayProxyResponse("Incorrect username/password or refreshToken", ErrUnauthorized)
		case cognito.ErrCodeUserNotFoundException:
			return writeGatewayProxyResponse("User not found", ErrBadRequest)
		case cognito.ErrCodeUserNotConfirmedException:
			return writeGatewayProxyResponse("User not confirmed", ErrUnauthorized)
		default:
			return writeGatewayProxyResponse("", ErrInternalServerError)
		}
	}

	userInput := &cognito.GetUserInput{
		AccessToken: authOutput.AuthenticationResult.AccessToken,
	}

	userOutput, err := cognitoClient.GetUser(userInput)
	if err != nil {
		log.Error("User Login: ", err)
		return writeGatewayProxyResponse("", ErrInternalServerError)
	}

	loginResponse := LoginResponse{
		UserId:             *userOutput.Username,
		AuthenticationInfo: authOutput,
	}

	resBytes, err := json.Marshal(loginResponse)
	if err != nil {
		log.Error("User Login: ", err)
		return writeGatewayProxyResponse("", ErrInternalServerError)
	}

	return writeGatewayProxyResponse(string(resBytes), nil)
}
