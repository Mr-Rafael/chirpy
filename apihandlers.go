package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Mr-Rafael/chirpy/internal/auth"
	"github.com/Mr-Rafael/chirpy/internal/database"
	"github.com/google/uuid"
)

type chirpParams struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type validateResponseErrorParams struct {
	Error string `json:"error"`
}

type chirpResponseOKParams struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type usersRequestParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequestParams struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds *int   `json:"expires_in_seconds,omitempty"`
}

type usersResponseParams struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type loginResponseParams struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func handlerHealthZ(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.Write([]byte("OK"))
}

func (c *apiConfig) handlerChirpsPOST(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	reqParams := chirpParams{}
	err := decoder.Decode(&reqParams)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to decode the request body: %v", err), "Something went wrong", http.StatusBadRequest)
		return
	}

	isValid := len(reqParams.Body) <= 140

	if !isValid {
		respondWithError(writer, "Error: chirp too long", "Chirp is too long", http.StatusBadRequest)
		return
	}

	createChirpParams := database.CreateChirpParams{
		Body:   sanitizeText(reqParams.Body),
		UserID: reqParams.UserID,
	}
	queryResult, err := c.db.CreateChirp(context.Background(), createChirpParams)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Error saving chirp on the database: %v", err), "Something went wrong.", http.StatusInternalServerError)
		return
	}
	respBody := chirpResponseOKParams{
		ID:        queryResult.ID,
		CreatedAt: queryResult.CreatedAt.Format("2021-01-01T00:00:00Z"),
		UpdatedAt: queryResult.UpdatedAt.Format("2021-01-01T00:00:00Z"),
		Body:      queryResult.Body,
		UserID:    queryResult.UserID,
	}
	respondWithJSON(writer, respBody, http.StatusCreated)
}

func (c *apiConfig) handlerChirpsGET(writer http.ResponseWriter, request *http.Request) {
	queryResult, err := c.db.GetChirps(context.Background())
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Error getting chirps from database: %v", err), "Something went wrong", http.StatusInternalServerError)
	}
	var responseData []chirpResponseOKParams
	for _, chirp := range queryResult {
		chirpData := chirpResponseOKParams{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt.Format("2021-01-01T00:00:00Z"),
			UpdatedAt: chirp.UpdatedAt.Format("2021-01-01T00:00:00Z"),
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		responseData = append(responseData, chirpData)
	}
	respondWithJSON(writer, responseData, http.StatusOK)
}

func (c *apiConfig) handlerChirpsGETID(writer http.ResponseWriter, request *http.Request) {
	chirpID, err := uuid.Parse(request.PathValue("chirp_id"))
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to parse the chirp id: %v", err), "Invalid Chirp ID", http.StatusNotFound)
		return
	}

	queryResult, err := c.db.GetChirp(context.Background(), chirpID)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Error getting chirps from database: %v", err), "Chirp not found", http.StatusNotFound)
		return
	}
	responseData := chirpResponseOKParams{
		ID:        queryResult.ID,
		CreatedAt: queryResult.CreatedAt.Format("2021-01-01T00:00:00Z"),
		UpdatedAt: queryResult.UpdatedAt.Format("2021-01-01T00:00:00Z"),
		Body:      queryResult.Body,
		UserID:    queryResult.UserID,
	}
	respondWithJSON(writer, responseData, http.StatusOK)
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
	var expires_in_seconds time.Duration
	if reqParams.ExpiresInSeconds == nil {
		expires_in_seconds = 1 * time.Hour
	} else {
		expires_in_seconds = time.Duration(*reqParams.ExpiresInSeconds) * time.Second
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

	return_jwt, err := auth.MakeJWT(userData.ID, c.secret, expires_in_seconds)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to generate JWT for user: %v"), "Something went wrong.", http.StatusInternalServerError)
	}

	responseParams := loginResponseParams{
		ID:        userData.ID,
		Email:     userData.Email,
		CreatedAt: userData.CreatedAt,
		UpdatedAt: userData.UpdatedAt,
		Token:     return_jwt,
	}
	respondWithJSON(writer, responseParams, http.StatusOK)
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
