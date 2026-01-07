package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
)

func  cleanBody(s string) string{
	profaneWords := []string {"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(s, " ")
	for i, word := range words{
		if slices.Contains(profaneWords,  strings.ToLower(word)){
			words[i] = "****"
		}
	}
	cleanedString := strings.Join(words, " ")
	return cleanedString 
}

func respondWithError(w http.ResponseWriter, statusCode int, err string){
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(err))
}

func respondWithJson[T any](w http.ResponseWriter, statusCode int, data *T){
	w.WriteHeader(statusCode)
	if s, ok := any(data).(string); ok{
		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(s))
		return
	}
	if data == nil {
		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(""))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if reflect.DeepEqual(*data, struct{}{}){
		return
	}
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		fmt.Println("Error: Trying to marshal given data")
		fmt.Printf("Data: %v\n", *data)
		fmt.Println(err)
		respondWithError(w, 500, "500: Internal Server Error")
	}
}
