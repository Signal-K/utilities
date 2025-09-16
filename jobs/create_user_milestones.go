package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	// "os"
	"strconv"
	"strings"
	"time"
)

func main() {
	supabaseURL := "http://127.0.0.1:54321"// os.Getenv("SUPABASE_URL")
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0" // os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	client := &http.Client{}

	loc, _ := time.LoadLocation("Australia/Sydney")
	now := time.Now().In(loc)
	weekday := int(now.Weekday()) // Sunday == 0
	weekStartDate := now.AddDate(0, 0, -weekday)
	weekStart := weekStartDate.Format("2006-01-02")

	// Find active users (1+ classifications in previous fortnight)
	twoWeeksAgo := now.AddDate(0, 0, -14).UTC().Format(time.RFC3339)
	// For 'inactive' users, they are assigned default milestones
	activeUsers := getActiveUserIDs(client, supabaseURL, supabaseKey, twoWeeksAgo)

	rand.Seed(time.Now().UnixNano())

	for _, userID := range activeUsers {
		// Check if user already has a milestone for this week
		if hasMilestoneForWeek(client, supabaseURL, supabaseKey, userID, weekStart) {
			continue
		}

		planetCount := getClassificationCount(client, supabaseURL, supabaseKey, userID, "planet", twoWeeksAgo)

		var classificationType, label string
		if planetCount >= 4 {
			classificationType = "telescope-minorPlanet"
			label = "asteroid"
		} else {
			classificationType = "planet"
			label = "planet"
		}

		target := rand.Intn(3) + 1
		name := fmt.Sprintf("Classify %d %s%s", target, label, plural(target))
		description := fmt.Sprintf("Head to the classification panel and identify %d %s%s.", target, label, plural(target))

		milestone := map[string]interface{}{
			"user_id":    userID,
			"week_start": weekStart,
			"milestone_data": map[string]interface{}{
				"name":                name,
				"structure":           "Telescope",
				"group":               "Astronomy",
				"icon":                "Telescope",
				"extendedDescription": description,
				"table":               "classifications",
				"field":               "classificationtype",
				"value":               classificationType,
				"requiredCount":       target,
			},
		}

		data, _ := json.Marshal(milestone)
		req, _ := http.NewRequest("POST", supabaseURL+"/rest/v1/user_milestones", bytes.NewBuffer(data))
		setHeaders(req, supabaseKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("❌ POST error for user %s: %v", userID, err)
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("✅ [%s] milestone created: %s\n", userID, name)
		} else {
			log.Printf("❌ Failed for user %s: %s", userID, string(body))
		}
	}
}

func getActiveUserIDs(client *http.Client, url, key, since string) []string {
	req, _ := http.NewRequest("GET", fmt.Sprintf(
		"%s/rest/v1/classifications?select=author&created_at=gte.%s&limit=10000",
		url, since), nil)

	setHeaders(req, key)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("❌ error fetching active users: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("❌ error reading response body: %v", err)
	}

	// Try decoding generically
	var raw []map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		log.Fatalf("❌ failed to decode active users JSON: %v\nRaw response: %s", err, string(body))
	}

	userSet := make(map[string]struct{})
	for _, row := range raw {
		if idRaw, ok := row["author"]; ok {
			if id, ok := idRaw.(string); ok {
				userSet[id] = struct{}{}
			}
		}
	}

	var ids []string
	for id := range userSet {
		ids = append(ids, id)
	}
	return ids
}

func hasMilestoneForWeek(client *http.Client, url, key, userID, weekStart string) bool {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/rest/v1/user_milestones?select=id&user_id=eq.%s&week_start=eq.%s", url, userID, weekStart), nil)
	setHeaders(req, key)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("⚠️ check milestone error: %v", err)
		return true
	}
	defer resp.Body.Close()

	var data []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	return len(data) > 0
}

// Count classifications by type since given date
func getClassificationCount(client *http.Client, url, key, userID, typ, since string) int {
	req, _ := http.NewRequest("GET", fmt.Sprintf(
		"%s/rest/v1/classifications?select=id&author=eq.%s&classificationtype=eq.%s&created_at=gte.%s&count=exact",
		url, userID, typ, since), nil)
	setHeaders(req, key)
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	cr := resp.Header.Get("Content-Range")
	parts := strings.Split(cr, "/")
	if len(parts) != 2 {
		return 0
	}
	count, _ := strconv.Atoi(parts[1])
	return count
}

func setHeaders(req *http.Request, key string) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Prefer", "plurality=plural")
	req.Header.Set("apikey", key)
	req.Header.Set("Authorization", "Bearer "+key)
}

func plural(n int) string {
if n == 1 {
	return ""
}
return "s"
}