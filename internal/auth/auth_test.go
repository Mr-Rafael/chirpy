package auth

import (
	"testing"

	"time"

	"github.com/google/uuid"
)

func TestValidJWT(t *testing.T) {
	test_uuid := uuid.New()
	secret_key, err := GenerateSecretKeyHS256()
	if err != nil {
		t.Errorf("Failed to generate a secret HS256 key for signing the jwt: %v", err)
	}
	duration := 1 * time.Hour

	jwt_token, err := MakeJWT(test_uuid, secret_key, duration)
	if err != nil {
		t.Errorf("Failed to generate a JWT: %v", err)
	}

	result, err := ValidateJWT(jwt_token, secret_key)
	if err != nil {
		t.Errorf("Failed to validate the JWT: %v", err)
	}
	if result == uuid.Nil {
		t.Errorf("Expected the UUID to be valid, but was invalid.")
	}
}

func TestWrongKeyJWT(t *testing.T) {
	test_uuid := uuid.New()
	correct_key, err := GenerateSecretKeyHS256()
	if err != nil {
		t.Errorf("Failed to generate a secret HS256 key for signing the jwt: %v", err)
	}
	incorrect_key, err := GenerateSecretKeyHS256()
	if err != nil {
		t.Errorf("Failed to generate a secret HS256 key for signing the jwt: %v", err)
	}
	duration := 1 * time.Hour

	jwt_token, err := MakeJWT(test_uuid, incorrect_key, duration)
	if err != nil {
		t.Errorf("failed to generate a JWT: %v", err)
	}

	_, err = ValidateJWT(jwt_token, correct_key)
	if err == nil {
		t.Errorf("Expected the JWT to be invalid, but was valid.")
	}
}

func TestExpiredJWT(t *testing.T) {
	test_uuid := uuid.New()
	correct_key, err := GenerateSecretKeyHS256()
	if err != nil {
		t.Errorf("Failed to generate a secret HS256 key for signing the jwt: %v", err)
	}
	incorrect_key, err := GenerateSecretKeyHS256()
	if err != nil {
		t.Errorf("Failed to generate a secret HS256 key for signing the jwt: %v", err)
	}
	duration := 1 * time.Second
	time.Sleep(3 * time.Second)

	jwt_token, err := MakeJWT(test_uuid, incorrect_key, duration)
	if err != nil {
		t.Errorf("failed to generate a JWT: %v", err)
	}

	_, err = ValidateJWT(jwt_token, correct_key)
	if err == nil {
		t.Errorf("Expected the JWT to be invalid, but was valid.")
	}
}
