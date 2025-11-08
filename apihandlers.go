package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type validateRequestParams struct {
	Body string `json:"body"`
}

type validateResponseErrorParams struct {
	Error string `json:"error"`
}

type validateResponseOKParams struct {
	CleanedBody string `json:"cleaned_body"`
}

type usersRequestParams struct {
	Email string `json:"email"`
}

type usersResponseParams struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func handlerHealthZ(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.Write([]byte("OK"))
}

func (c *apiConfig) handlerValidateChirp(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	reqParams := validateRequestParams{}
	err := decoder.Decode(&reqParams)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to decode the request body: %v", err), "Something went wrong", http.StatusInternalServerError)
		return
	}

	isValid := len(reqParams.Body) <= 140

	if isValid {
		respBody := validateResponseOKParams{
			CleanedBody: sanitizeText(reqParams.Body),
		}
		respondWithJSON(writer, respBody, http.StatusOK)
	} else {
		respondWithError(writer, "Error: chirp too long", "Chirp is too long", http.StatusBadRequest)
	}
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
		respondWithError(writer, fmt.Sprintf("The username came empty: %v", err), "Something went wrong", http.StatusBadRequest)
		return
	}

	queryResult, err := c.db.CreateUser(context.Background(), reqParams.Email)
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

func respondWithError(writer http.ResponseWriter, logMessage string, apiErrorMessage string, statusCode int) {
	fmt.Printf("[Error]: %v\n", logMessage)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	respBody := validateResponseErrorParams{
		Error: apiErrorMessage,
	}
	json.NewEncoder(writer).Encode(respBody)
}

func respondWithJSON(writer http.ResponseWriter, data any, statusCode int) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	json.NewEncoder(writer).Encode(data)
}

func sanitizeText(text string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	splitText := strings.Fields(text)

	for i := range splitText {
		for _, word := range profaneWords {
			if strings.ToLower(splitText[i]) == word {
				splitText[i] = "****"
			}
		}
	}

	return strings.Join(splitText, " ")
}
