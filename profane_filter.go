package main

import "strings"

func profaneFilter(s string) string {
	words := strings.Split(s, " ")
	expWords := []string{}
	for _, word := range words {
		if strings.ToLower(word) == "kerfuffle" ||
			strings.ToLower(word) == "sharbert" ||
			strings.ToLower(word) == "fornax" {
			word = "****"
		}
		expWords = append(expWords, word)
	}

	return strings.Join(expWords, " ")
}
