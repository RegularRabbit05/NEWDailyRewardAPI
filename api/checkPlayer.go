package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func CheckPlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	playerUUID := os.Getenv("PLAYER_UUID")
	apiKey := os.Getenv("API_KEY")

	// create request
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.hypixel.net/player?uuid=%s", playerUUID), nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	// add headers for api key
	req.Header.Add("API-Key", apiKey)
	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	// check error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	defer func() { _ = resp.Body.Close() }()
	// check status code
	if resp.StatusCode != 200 {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Response is not 200 (" + resp.Status + ")"))
		return
	}

	//read body
	var body map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	lastRewardTimestamp := int64(body["player"].(map[string]interface{})["lastClaimedReward"].(float64) / 1000)
	// convert to time with Rome timezone
	lastRewardTime := time.Unix(lastRewardTimestamp, 0).In(time.FixedZone("Rome", 1*60*60))
	currentTime := time.Now().In(time.FixedZone("Rome", 1*60*60))

	// check if date is today
	DateEqual := func(date1, date2 time.Time) bool {
		y1, m1, d1 := date1.Date()
		y2, m2, d2 := date2.Date()
		return y1 == y2 && m1 == m2 && d1 == d2
	}

	// write response
	var response = map[string]interface{}{
		"lastReward":  lastRewardTime.Format("2006-01-02"),
		"currentDate": currentTime.Format("2006-01-02"),
		"result":      true,
	}
	if DateEqual(lastRewardTime, currentTime) {
		response["result"] = true
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(response); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		return
	}
	response["result"] = false
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
}
