package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"strconv"
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
	leaderboardEndpoint := os.Getenv("LEADERBOARD_URL")
	leaderboardPayload := os.Getenv("LEADERBOARD_PAYLOAD")

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

	var leaderboardPos = -1
	if r.URL.RawQuery == "bot=true" {
		var postString = []byte(leaderboardPayload)
		req2, err := http.NewRequest("POST", leaderboardEndpoint, bytes.NewBuffer(postString))
		req2.Header.Set("Content-Type", "application/json")
		if err == nil {
			client2 := &http.Client{
				Timeout: 30 * time.Second,
			}
			resp2, err := client2.Do(req2)
			if err == nil {
				defer resp2.Body.Close()
				var body2 map[string]interface{}
				if err = json.NewDecoder(resp2.Body).Decode(&body2); err == nil {
					switch body2["content"].(type) {
						case string: {
							data := body2["content"].(string)
							data = strings.ReplaceAll(data, "&lt;\\/td&gt;&lt;td&gt;&lt;span class=\\\"mc_font\\\"&gt;&lt;a href=\\\"\\/hypixel\\/player\\/stats\\/", "")
							splits := strings.SplitN(data, os.Getenv("PLAYER_NAME"), 2)
							if len(splits) == 2 {
								splits := strings.Split(splits[0], ";")
								leaderboardPos, err = strconv.Atoi(splits[len(splits)-1])
								if err != nil {
									leaderboardPos = -7
								}
							} else {
								leaderboardPos = -6
							}
						}
						default: leaderboardPos = -5
					}
				} else {
					leaderboardPos = -4
				}
			} else {
				leaderboardPos = -3
			}
		} else {
			leaderboardPos = -2
		}
	}

	currentTimestamp := time.Now()

	lastRewardTimestamp := int64(body["player"].(map[string]interface{})["lastClaimedReward"].(float64) / 1000)
	// convert to time with Rome timezone
	lastRewardTime := time.Unix(lastRewardTimestamp, 0).In(time.FixedZone("Rome", 1*60*60))
	currentTime := currentTimestamp.In(time.FixedZone("Rome", 1*60*60))

	// check if date is today
	DateEqual := func(date1, date2 time.Time) bool {
		y1, m1, d1 := date1.Date()
		y2, m2, d2 := date2.Date()
		return y1 == y2 && m1 == m2 && d1 == d2
	}

	// write response
	var response = map[string]interface{}{
		"lastRewardTimestamp": lastRewardTimestamp,
		"currentTimestamp":    currentTimestamp.Unix(),
		"lastReward":          lastRewardTime.Format("2006-01-02"),
		"currentDate":         currentTime.Format("2006-01-02"),
		"rewardStreak":        body["player"].(map[string]interface{})["rewardScore"],
		"rewardLeaderboard":   leaderboardPos,
		"result":              true,
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
	SendDiscordWebhook := func() {
		discordWebhook := os.Getenv("DISCORD_WEBHOOK")
		discordUsername := os.Getenv("DISCORD_USERNAME")
		discordAvatarURL := os.Getenv("DISCORD_AVATAR")
		discordContent := os.Getenv("DISCORD_MESSAGE")
		discordTTS := os.Getenv("DISCORD_TTS") == "true"

		var jsonBody = map[string]interface{}{
			"username":   discordUsername,
			"avatar_url": discordAvatarURL,
			"content":    strings.ReplaceAll(fmt.Sprintf(discordContent, fmt.Sprint(lastRewardTimestamp)), "\\n", "\n"),
			"tts":        discordTTS,
		}
		jsonStr, err := json.Marshal(jsonBody)
		if err != nil {
			return
		}
		req, _ = http.NewRequest("POST", discordWebhook, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		_, _ = client.Do(req)
	}
	SendDiscordWebhook()
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
