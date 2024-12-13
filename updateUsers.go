package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dayf0rdie1999/Chirpy/internal/auth"
	"github.com/dayf0rdie1999/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (a *apiConfig) UpdateUsers(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		NewPassword string `json:"password"`
		Email       string `json:"email"`
	}

	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ResponseError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		ResponseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	user_id, err := auth.ValidateJWT(bearerToken, a.authSecret)
	if err != nil {
		ResponseError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	newHashedPassword, err := auth.HashPassword(params.NewPassword)
	if err != nil {
		ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedUser, err := a.queries.UpdateUserPassword(req.Context(), database.UpdateUserPasswordParams{
		Email:          params.Email,
		HashedPassword: newHashedPassword,
		ID:             user_id,
	})
	if err != nil {
		ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type result struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	w.WriteHeader(http.StatusOK)
	data, err := json.Marshal(result{
		ID:          updatedUser.ID,
		CreatedAt:   updatedUser.CreatedAt,
		UpdatedAt:   updatedUser.UpdatedAt,
		Email:       updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
	})
	if err != nil {
		ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
