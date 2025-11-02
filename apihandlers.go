package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type validateRequestParams struct {
	Body string `json:"body"`
}

type validateResponseErrorParams struct {
	Error string `json:"error"`
}

type validateResponseOKParams struct {
	Valid bool `json:"valid"`
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
		internalServerError(writer, fmt.Sprintf("Failed to decode the request body: %v", err), "Something went wrong")
		return
	}

	isValid := len(reqParams.Body) <= 140

	var respData []byte
	var respCode int
	if isValid {
		respBody := validateResponseOKParams{
			Valid: true,
		}
		respData, err = json.Marshal(respBody)
		respCode = http.StatusOK
	} else {
		respBody := validateResponseErrorParams{
			Error: "Chirp is too long",
		}
		respData, err = json.Marshal(respBody)
		respCode = http.StatusBadRequest
	}
	if err != nil {
		internalServerError(writer, fmt.Sprintf("Error marshalling JSON: %v", err), "Something went wrong")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(respCode)
	writer.Write(respData)
}

func internalServerError(writer http.ResponseWriter, logMessage string, apiErrorMessage string) {
	fmt.Println(logMessage)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusInternalServerError)
	respBody := validateResponseErrorParams{
		Error: apiErrorMessage,
	}
	json.NewEncoder(writer).Encode(respBody)
}
