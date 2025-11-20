package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Mr-Rafael/chirpy/internal/auth"
	"github.com/Mr-Rafael/chirpy/internal/database"
	"github.com/google/uuid"
)

type usersRequestParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequestParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequestParams struct {
	Token string `json:"token"`
}

type usersResponseParams struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type loginResponseParams struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (c *apiConfig) handlerUsers(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	reqParams := usersRequestParams{}
	err := decoder.Decode(&reqParams)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to create user: %v", err), "Something went wrong", http.StatusBadRequest)
		return
	}
	if len(reqParams.Email) <= 0 {
		respondWithError(writer, fmt.Sprintf("The username came empty: %v", err), "Missing param: email", http.StatusBadRequest)
		return
	}
	if len(reqParams.Password) <= 0 {
		respondWithError(writer, fmt.Sprintf("The password came empty: %v", err), "Missing param: password", http.StatusBadRequest)
		return
	}
	hashedPassword, err := auth.HashPassword(reqParams.Password)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("failed to hash the password: %v", err), "Something went wrong", http.StatusInternalServerError)
		return
	}

	queryParams := database.CreateUserParams{
		Email:          reqParams.Email,
		HashedPassword: hashedPassword,
	}
	queryResult, err := c.db.CreateUser(context.Background(), queryParams)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to save the user to database: %v", err), "Something went wrong", http.StatusInternalServerError)
		return
	}

	responseBody := usersResponseParams{
		ID:        queryResult.ID,
		CreatedAt: queryResult.CreatedAt,
		UpdatedAt: queryResult.UpdatedAt,
		Email:     queryResult.Email,
	}
	respondWithJSON(writer, responseBody, http.StatusCreated)
}

func (c *apiConfig) handlerLogin(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	reqParams := loginRequestParams{}
	err := decoder.Decode(&reqParams)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to create user: %v", err), "Something went wrong", http.StatusBadRequest)
		return
	}
	if len(reqParams.Email) <= 0 {
		respondWithError(writer, "The username came empty.", "Missing param: email", http.StatusBadRequest)
		return
	}
	if len(reqParams.Password) <= 0 {
		respondWithError(writer, "The password came empty.", "Missing param: password", http.StatusBadRequest)
		return
	}

	userData, err := c.db.GetUser(context.Background(), reqParams.Email)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to get user data: %v", err), "User not found", http.StatusNotFound)
		return
	}

	correctPassword, err := auth.CheckPasswordHash(reqParams.Password, userData.HashedPassword)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to compare password and hash: %v", err), "Something went wrong", http.StatusInternalServerError)
		return
	}
	if !correctPassword {
		respondWithError(writer, "Login attempt with incorrect password.", "Incorrect email or password", http.StatusUnauthorized)
		return
	}

	return_jwt, err := auth.MakeJWT(userData.ID, c.secret, 1*time.Hour)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to generate JWT for user: %v", err), "Something went wrong.", http.StatusInternalServerError)
		return
	}
	refresh_token, err := auth.GenerateSecretKeyHS256()
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to generate refresh token for user: %v", err), "Something went wrong.", http.StatusInternalServerError)
		return
	}
	create_refresh_token_params := database.CreateRefreshTokenParams{
		Token:     refresh_token,
		UserID:    userData.ID,
		ExpiresAt: time.Now().Add(time.Duration(24 * 60 * time.Hour)),
	}
	_, err = c.db.CreateRefreshToken(context.Background(), create_refresh_token_params)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to save the refresh token to the database: %v", err), "Something went wrong.", http.StatusInternalServerError)
		return
	}

	responseParams := loginResponseParams{
		ID:           userData.ID,
		Email:        userData.Email,
		CreatedAt:    userData.CreatedAt,
		UpdatedAt:    userData.UpdatedAt,
		Token:        return_jwt,
		RefreshToken: refresh_token,
	}
	respondWithJSON(writer, responseParams, http.StatusOK)
}
