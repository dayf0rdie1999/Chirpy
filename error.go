package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func ResponseError(res http.ResponseWriter, errorMessage string, statusRequest int) {
	res.WriteHeader(statusRequest)
	data, err := json.Marshal(errorResponse{
		Error: errorMessage,
	})
	if err != nil {
		log.Fatalln("Unexpected Error from marshal error")
	}
	res.Write(data)
}
