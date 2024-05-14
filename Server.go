package main

import (
	"NEWDailyRewardAPI/api"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/checkPlayer", api.CheckPlayer)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return
	}
}
