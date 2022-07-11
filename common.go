package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func cognitoSecretHash(username string) string {
	mac := hmac.New(sha256.New, []byte(appClientSecret))
	mac.Write([]byte(username + appClientID))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func getCognitoUsername(email, phoneNumber string) string {
	if email != "" {
		return email
	}
	return phoneNumber
}
