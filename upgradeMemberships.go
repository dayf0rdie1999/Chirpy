package main

import (
	"encoding/json"
	"net/http"

	"github.com/dayf0rdie1999/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (a *apiConfig) UpgradeMembership(w http.ResponseWriter, r *http.Request) {
	type DataBody struct {
		UserId uuid.UUID `json:"user_id"`
	}
	type parameters struct {
		Event string   `json:"event"`
		Data  DataBody `json:"data"`
	}

	apiToken, err := auth.GetApiToken(r.Header)
	if err != nil {
		ResponseError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if apiToken != a.polkaApiKey {
		ResponseError(w, "Forbidden to access the endpoint", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		ResponseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if params.Event != "user.upgraded" {
		ResponseError(w, "No event found", http.StatusNoContent)
		return
	}

	_, err = a.queries.UpgradeChirpMembershipByuserId(r.Context(), params.Data.UserId)
	if err != nil {
		ResponseError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
