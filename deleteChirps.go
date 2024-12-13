package main

import (
	"net/http"

	"github.com/dayf0rdie1999/Chirpy/internal/auth"
	"github.com/dayf0rdie1999/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (a *apiConfig) DeleteChrips(w http.ResponseWriter, req *http.Request) {
	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ResponseError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	user_id, err := auth.ValidateJWT(bearerToken, a.authSecret)
	if err != nil {
		ResponseError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	chirp, err := a.queries.GetChirpById(req.Context(), uuid.MustParse(req.PathValue("chirpID")))
	if err != nil {
		ResponseError(w, err.Error(), http.StatusNotFound)
		return
	}

	if chirp.UserID != user_id {
		ResponseError(w, "Not Allowed to delete chirp", http.StatusForbidden)
		return
	}

	_, err = a.queries.DeleteChirp(req.Context(), database.DeleteChirpParams{
		ID:     uuid.MustParse(req.PathValue("chirpID")),
		UserID: user_id,
	})
	if err != nil {
		ResponseError(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
