package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dayf0rdie1999/Chirpy/internal/auth"
	"github.com/dayf0rdie1999/Chirpy/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	queries        *database.Queries
	authSecret     string
	polkaApiKey    string
}

func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (a *apiConfig) resetMetric(res http.ResponseWriter, req *http.Request) {
	a.fileserverHits.Store(0)
	environment := os.Getenv("PLATFORM")
	if environment != "dev" {
		ResponseError(res, "Not allowed to delete users", http.StatusForbidden)
		return
	}
	a.queries.DeleteUsers(req.Context())
	res.WriteHeader(http.StatusOK)
}

func (a *apiConfig) renderAdminMetrics(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
	res.Header().Add("Content-Type", "text/html")
	htmlTemplate := fmt.Sprintf(
		`<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %v times!</p>
			</body>
		</html>`,
		a.fileserverHits.Load(),
	)

	res.Write([]byte(htmlTemplate))
}

func (a *apiConfig) handleCreateUser(res http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		ResponseError(res, "Unexpected Failure Parsing Data", http.StatusInternalServerError)
		return
	}

	hashPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		ResponseError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := a.queries.CreateUser(req.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashPassword,
	})
	if err != nil {
		ResponseError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusCreated)
	res.Header().Set("Content-Type", "application/json")
	type result struct {
		ID             uuid.UUID `json:"id"`
		Email          string    `json:"email"`
		HashedPassword string    `json:"hashed_password"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		IsChirpyRed    bool      `json:"is_chirpy_red"`
	}
	data, err := json.Marshal(result{
		ID:             user.ID,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		IsChirpyRed:    user.IsChirpyRed,
	})
	if err != nil {
		ResponseError(res, err.Error(), http.StatusInternalServerError)
	}
	res.Write(data)
}

func (a *apiConfig) handleInsertChirps(res http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	// Validate JWT
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ResponseError(res, err.Error(), http.StatusUnauthorized)
		return
	}

	userId, err := auth.ValidateJWT(token, a.authSecret)
	if err != nil {
		ResponseError(res, err.Error(), http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		ResponseError(res, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if len(params.Body) > 140 {
		ResponseError(res, "Chirp is too long", http.StatusBadRequest)
		return
	}

	text := strings.ReplaceAll(params.Body, "kerfuffle", "****")
	text = strings.ReplaceAll(text, "sharbert", "****")
	text = strings.ReplaceAll(text, "fornax", "****")
	text = strings.ReplaceAll(text, "Fornax", "****")

	data, err := a.queries.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   text,
		UserID: userId,
	})
	if err != nil {
		ResponseError(res, "Can't insert chirp", http.StatusInternalServerError)
		return
	}

	type responseBody struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	content, err := json.Marshal(responseBody{
		ID:        data.ID,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
		Body:      data.Body,
		UserID:    userId,
	})
	if err != nil {
		ResponseError(res, "Can't parse result", http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusCreated)
	res.Write(content)
}

func (a *apiConfig) getChirps(res http.ResponseWriter, req *http.Request) {
	author_id := req.URL.Query().Get("author_id")

	sortDirection := "asc"
	sortDirectionParam := req.URL.Query().Get("sort")
	if sortDirectionParam != "" {
		sortDirection = sortDirectionParam
	}
	var data []database.Chirp
	var err error
	if author_id != "" {
		data, err = a.queries.GetAllChirpsByUserId(req.Context(), uuid.MustParse(author_id))
	} else {
		data, err = a.queries.GetAllChirps(req.Context())
	}

	if err != nil {
		ResponseError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	sort.Slice(data, func(i, j int) bool {
		if sortDirection == "desc" {
			return data[i].CreatedAt.After(data[j].CreatedAt)
		}
		return data[i].CreatedAt.Before(data[j].CreatedAt)
	})

	type ChirpResponse struct {
		Id   uuid.UUID `json:"id"`
		Body string    `json:"body"`
	}
	responseBody := []ChirpResponse{}
	for _, value := range data {
		responseBody = append(responseBody, ChirpResponse{
			Id:   value.ID,
			Body: value.Body,
		})
	}
	result, err := json.Marshal(responseBody)
	if err != nil {
		ResponseError(res, "Failed to serialize data", http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write(result)
}

func (a *apiConfig) getChirpById(res http.ResponseWriter, req *http.Request) {
	data, err := a.queries.GetChirpById(req.Context(), uuid.MustParse(req.PathValue("chirpID")))
	if err != nil {
		ResponseError(res, err.Error(), http.StatusNotFound)
		return
	}

	result, err := json.Marshal(data)
	if err != nil {
		ResponseError(res, "Failed to serialize data", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write(result)
}

func (a *apiConfig) handleUserLogin(res http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password        string `json:"password"`
		Email           string `json:"email"`
		ExpireInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		ResponseError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := a.queries.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		ResponseError(res, err.Error(), http.StatusNotFound)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		ResponseError(res, "Incorrect email or password", http.StatusUnauthorized)
		return
	}
	if params.ExpireInSeconds == 0 {
		params.ExpireInSeconds = 3600
	}
	token, err := auth.MakeJWT(user.ID, a.authSecret)

	if err != nil {
		ResponseError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		ResponseError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	refreshTokenEntity, err := a.queries.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})
	if err != nil {
		ResponseError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	type result struct {
		ID           string    `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}
	res.WriteHeader(http.StatusOK)
	marshalContent, err := json.Marshal(result{
		ID:           user.ID.String(),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshTokenEntity.Token,
		IsChirpyRed:  user.IsChirpyRed,
	})
	if err != nil {
		ResponseError(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Write(marshalContent)
}

func (a *apiConfig) refreshAccessToken(res http.ResponseWriter, req *http.Request) {
	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ResponseError(res, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := a.queries.GetUserFromRefreshToken(req.Context(), bearerToken)
	if err != nil {
		ResponseError(res, err.Error(), http.StatusUnauthorized)
		return
	}

	newAccessToken, err := auth.MakeJWT(user.ID, a.authSecret)
	if err != nil {
		ResponseError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	type ResponseBody struct {
		Token string `json:"token"`
	}
	res.WriteHeader(http.StatusOK)
	marshalContent, err := json.Marshal(ResponseBody{
		Token: newAccessToken,
	})
	if err != nil {
		ResponseError(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Write(marshalContent)
}

func (a *apiConfig) RevokeRefreshToken(res http.ResponseWriter, req *http.Request) {
	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ResponseError(res, err.Error(), http.StatusUnauthorized)
		return
	}

	err = a.queries.RevokeRefreshToken(req.Context(), bearerToken)
	if err != nil {
		ResponseError(res, err.Error(), http.StatusNotFound)
		return
	}
	res.WriteHeader(http.StatusNoContent)
}
