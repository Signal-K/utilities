package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LinkedAnomaly struct {
	ID        int64     `json:"id"`
	Author    string    `json:"author"`
	Unlocked  bool      `json:"unlocked"`
	Automaton string    `json:"automaton"`
	Date      time.Time `json:"date"`
}

func main() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	req, err := http.NewRequest("GET", supabaseURL+"/rest/v1/linked_anomalies?select=id,author,unlocked,automaton,date&automaton=eq.Telescope&order=date.asc", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var allEntries []LinkedAnomaly
	if err := json.NewDecoder(resp.Body).Decode(&allEntries); err != nil {
		panic(err)
	}

	// Time reference: Sunday 00:00 AEST
	location, _ := time.LoadLocation("Australia/Melbourne")
	now := time.Now().In(location)

	// Find end of the current week (Sunday 00:00 AEST)
	weekday := int(now.Weekday())
	daysUntilSunday := (7 - weekday) % 7
	endOfWeek := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).AddDate(0, 0, daysUntilSunday)

	timeRemaining := endOfWeek.Sub(now).Hours()

	// Group locked entries by user
	userLocked := make(map[string][]LinkedAnomaly)
	userUnlockedCount := make(map[string]int)

	for _, entry := range allEntries {
		if entry.Automaton != "Telescope" {
			continue
		}
		if entry.Unlocked {
			userUnlockedCount[entry.Author]++
		} else {
			userLocked[entry.Author] = append(userLocked[entry.Author], entry)
		}
	}

	for author, locked := range userLocked {
		unlocked := userUnlockedCount[author]
		lockedCount := len(locked)

		// Basic rule: don't unlock if it's too soon based on what's already been unlocked
		if unlocked >= 4 || lockedCount == 0 {
			continue
		}

		// Estimate pacing window: X anomalies across the time window
		expectedPerUnlock := 168.0 / 4 // hours per unlock in the week (168 = 7d * 24h)
		minHoursPerUnlock := expectedPerUnlock * float64(unlocked+1)

		if timeRemaining < minHoursPerUnlock {
			// Enough time has passed to unlock the next
			next := locked[0]
			patchBody := map[string]any{
				"unlocked": true,
			}
			jsonBody, _ := json.Marshal(patchBody)

			url := fmt.Sprintf("%s/rest/v1/linked_anomalies?id=eq.%d", supabaseURL, next.ID)
			patchReq, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonBody))
			patchReq.Header.Set("apikey", supabaseKey)
			patchReq.Header.Set("Authorization", "Bearer "+supabaseKey)
			patchReq.Header.Set("Content-Type", "application/json")
			patchReq.Header.Set("Prefer", "return=representation")

			res, err := client.Do(patchReq)
			if err != nil {
				fmt.Printf("❌ Failed to unlock anomaly %d for user %s\n", next.ID, next.Author)
				continue
			}
			defer res.Body.Close()

			if res.StatusCode == 200 || res.StatusCode == 204 {
				fmt.Printf("✅ Unlocked anomaly %d for user %s (timeRemaining=%.1f hrs)\n", next.ID, next.Author, timeRemaining)
			} else {
				fmt.Printf("❌ Error unlocking anomaly %d for user %s, status %d\n", next.ID, next.Author, res.StatusCode)
			}
		} else {
			fmt.Printf("⏳ Skipping %s – %.1f hrs still remaining, too early to unlock next\n", author, timeRemaining)
		}
	}
}