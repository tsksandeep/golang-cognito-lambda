package main

import (
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

var allowedMethods = map[string]bool{
	"GET":  true,
	"POST": true,
}

func isMethodAllowed(method string) bool {
	_, ok := allowedMethods[method]
	return ok
}

func checkRequest(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !isMethodAllowed(request.HTTPMethod) {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Body:       "Method not allowed",
		}, errors.New("request method not allowed")
	}

	return nil, nil
}

func checkRegister(userRegister UserRegister) error {
	if userRegister.Username == "" {
		return errors.New("username is not present")
	}
	if userRegister.Password == "" {
		return errors.New("password is not present")
	}
	if userRegister.Email == "" {
		return errors.New("email is not present")
	}
	if userRegister.PhoneNumber == "" {
		return errors.New("phone number is not present")
	}
	return nil
}
