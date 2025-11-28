package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/Mr-Rafael/chirpy/internal/auth"
	"github.com/Mr-Rafael/chirpy/internal/database"
	"github.com/google/uuid"
)

type chirpParams struct {
	Body string `json:"body"`
}

type chirpResponseOKParams struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (c *apiConfig) handlerChirpsPOST(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	reqParams := chirpParams{}
	err := decoder.Decode(&reqParams)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to decode the request body: %v", err), "Something went wrong", http.StatusBadRequest)
		return
	}

	bearerToken, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to get bearer from request: %v", err), "Unauthorized", http.StatusUnauthorized)
		return
	}
	jwt_user_id, err := auth.ValidateJWT(bearerToken, c.secret)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Error validating JWT: %v", err), "Unauthorized", http.StatusUnauthorized)
		return
	}
	if jwt_user_id == uuid.Nil {
		respondWithError(writer, fmt.Sprintf("Error validating JWT: %v", err), "Unauthorized", http.StatusUnauthorized)
		return
	}

	isValid := len(reqParams.Body) <= 140

	if !isValid {
		respondWithError(writer, "Error: chirp too long", "Chirp is too long", http.StatusBadRequest)
		return
	}

	createChirpParams := database.CreateChirpParams{
		Body:   sanitizeText(reqParams.Body),
		UserID: jwt_user_id,
	}
	queryResult, err := c.db.CreateChirp(context.Background(), createChirpParams)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Error saving chirp on the database: %v", err), "Something went wrong.", http.StatusInternalServerError)
		return
	}

	respBody := chirpResponseOKParams{
		ID:        queryResult.ID,
		CreatedAt: queryResult.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: queryResult.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Body:      queryResult.Body,
		UserID:    queryResult.UserID,
	}
	respondWithJSON(writer, respBody, http.StatusCreated)
}

func (c *apiConfig) handlerChirpsGET(writer http.ResponseWriter, request *http.Request) {
	authorID := request.URL.Query().Get("author_id")
	sortOrder := request.URL.Query().Get("sort")

	var queryResult []database.Chirp

	authorUUID, err := uuid.Parse(authorID)
	if err != nil || len(authorUUID) == 0 {
		queryResult, err = c.db.GetChirps(context.Background())
	} else {
		queryResult, err = c.db.GetChirpsByUser(context.Background(), authorUUID)
	}
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Error getting chirps from database: %v", err), "Something went wrong", http.StatusInternalServerError)
	}

	switch sortOrder {
	case "asc":
		sort.Slice(queryResult, func(i, j int) bool { return queryResult[i].CreatedAt.Before(queryResult[j].CreatedAt) })
	case "desc":
		sort.Slice(queryResult, func(i, j int) bool { return queryResult[i].CreatedAt.After(queryResult[j].CreatedAt) })
	}

	var responseData []chirpResponseOKParams
	for _, chirp := range queryResult {
		chirpData := chirpResponseOKParams{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: chirp.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
		CreatedAt: queryResult.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: queryResult.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Body:      queryResult.Body,
		UserID:    queryResult.UserID,
	}
	respondWithJSON(writer, responseData, http.StatusOK)
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

func (c *apiConfig) handlerChirpsDELETE(writer http.ResponseWriter, request *http.Request) {
	chirpID, err := uuid.Parse(request.PathValue("chirp_id"))
	if err != nil {
		respondWithError(writer, fmt.Sprintf("The Chirp ID was not specified: %v", err), "Invalid Chirp ID", http.StatusNotFound)
		return
	}

	bearerToken, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to get bearer from request: %v", err), "Unauthorized", http.StatusUnauthorized)
		return
	}
	jwt_user_id, err := auth.ValidateJWT(bearerToken, c.secret)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Error validating JWT: %v", err), "Unauthorized", http.StatusUnauthorized)
		return
	}
	if jwt_user_id == uuid.Nil {
		respondWithError(writer, fmt.Sprintf("Error validating JWT: %v", err), "Unauthorized", http.StatusUnauthorized)
		return
	}

	chirpData, err := c.db.GetChirp(context.Background(), chirpID)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Error fetching the Chirp from the database: %v", err), "Something went wrong.", http.StatusNotFound)
		return
	}
	if chirpData.UserID != jwt_user_id {
		respondWithError(writer, "The Chirp's User ID doesn't match the JWT User ID.", "Unauthorized.", http.StatusForbidden)
		return
	}

	err = c.db.DeleteChirp(context.Background(), chirpID)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Error saving chirp on the database: %v", err), "Something went wrong.", http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}
