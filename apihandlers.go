package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

func handlerHealthZ(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.Write([]byte("OK"))
}

func handlerValidateChirp(writer http.ResponseWriter, request *http.Request) {
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
		respondWithJSON(writer, respBody)
	} else {
		respondWithError(writer, "Error: chirp too long", "Chirp is too long", http.StatusBadRequest)
	}
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Error marshalling JSON: %v", err), "Something went wrong", http.StatusInternalServerError)
		return
	}
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

func respondWithJSON(writer http.ResponseWriter, data any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(data)
}

func sanitizeText(text string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	splitText := strings.Fields(text)

	for i, _ := range splitText {
		for _, word := range profaneWords {
			if strings.ToLower(splitText[i]) == word {
				splitText[i] = "****"
			}
		}
	}

	return strings.Join(splitText, " ")
}
