package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

const apiPrefix = "/api/v1"

func getMethodPathString(method, path string) string {
	return method + " " + apiPrefix + path
}

var (
	registerUser              = getMethodPathString("POST", "/auth/register")
	loginUser                 = getMethodPathString("POST", "/auth/login")
	otpUser                   = getMethodPathString("POST", "/auth/otp")
	forgotPasswordUser        = getMethodPathString("POST", "/auth/forgot")
	confirmForgotPasswordUser = getMethodPathString("POST", "/auth/confirmforgot")
)

func router(ctx context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	methodAndPathString := request.HTTPMethod + " " + request.Path

	switch methodAndPathString {
	case registerUser:
		return Register(ctx, request)

	case loginUser:
		return Login(ctx, request)

	case otpUser:
		return OTP(ctx, request)

	case forgotPasswordUser:
		return ForgotPassword(ctx, request)

	case confirmForgotPasswordUser:
		return ConfirmForgotPassword(ctx, request)

	default:
		return events.APIGatewayProxyResponse{
			Headers: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "'POST,OPTIONS'",
			},
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid method and path: " + methodAndPathString,
		}
	}
}
