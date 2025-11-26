package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type polkaWebhookRequestParams struct {
	Event string `json:"event"`
	Data  data   `json:"data"`
}

type data struct {
	UserID uuid.UUID `json:"user_id"`
}

func (c *apiConfig) handlerPolkaWebhook(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	reqParams := polkaWebhookRequestParams{}
	err := decoder.Decode(&reqParams)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to decode request: %v", err), "Something went wrong", http.StatusBadRequest)
		return
	}
	if reqParams.Event != "user.upgraded" {
		fmt.Println("Received an event that is not 'user.upgraded'. Ignoring the request.")
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	err = c.db.UpgradeUser(context.Background(), reqParams.Data.UserID)
	if err != nil {
		respondWithError(writer, fmt.Sprintf("Failed to upgrade the user on the database: %v", err), "Something went wrong", http.StatusNotFound)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
