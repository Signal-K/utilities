package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Route struct {
	ID                 int64     `json:"id"`
	Author             string    `json:"author"`
	RouteConfiguration *json.RawMessage `json:"routeConfiguration"`
	Timestamp          time.Time `json:"timestamp"`
	Location           int64     `json:"location"`
}

func main() {
	supabaseURL := "http://127.0.0.1:54321" // os.Getenv("SUPABASE_URL")
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0" // os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	// Calculate cutoff time: 24 hours and 1 minute ago
	cutoffTime := time.Now().Add(-24*time.Hour - 1*time.Minute)
	cutoffTimeISO := cutoffTime.Format(time.RFC3339)

	// Step 1: Fetch routes older than 24h 1m
	filterQuery := url.QueryEscape(fmt.Sprintf("lt.%s", cutoffTimeISO))
	req, err := http.NewRequest("GET", supabaseURL+"/rest/v1/routes?select=*&timestamp="+filterQuery, nil)
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

	if resp.StatusCode != 200 {
		fmt.Printf("❌ Failed to fetch routes, status: %d\n", resp.StatusCode)
		return
	}

	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		panic(err)
	}

	// Step 2: Write to logs/routes_<timestamp>.json for backup
	if len(routes) > 0 {
		timestamp := time.Now().Format("2006-01-02T15-04-05")
		filename := fmt.Sprintf("logs/routes_%s.json", timestamp)

		// Create logs dir if it doesn't exist
		if _, err := os.Stat("logs"); os.IsNotExist(err) {
			os.Mkdir("logs", 0755)
		}

		data, err := json.MarshalIndent(routes, "", "  ")
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			panic(err)
		}
		fmt.Printf("✅ Backup log written to %s (%d routes)\n", filename, len(routes))
	} else {
		fmt.Println("ℹ️ No old routes to delete.")
		return
	}

	// Step 3: Delete routes older than 24h 1m
	deleteFilterQuery := url.QueryEscape(fmt.Sprintf("lt.%s", cutoffTimeISO))
	deleteReq, err := http.NewRequest("DELETE", supabaseURL+"/rest/v1/routes?timestamp="+deleteFilterQuery, nil)
	if err != nil {
		panic(err)
	}
	deleteReq.Header.Set("apikey", supabaseKey)
	deleteReq.Header.Set("Authorization", "Bearer "+supabaseKey)
	deleteReq.Header.Set("Prefer", "return=minimal")

	deleteResp, err := client.Do(deleteReq)
	if err != nil {
		panic(err)
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode >= 200 && deleteResp.StatusCode < 300 {
		fmt.Printf("✅ Successfully deleted %d old routes (older than %s).\n", len(routes), cutoffTime.Format(time.RFC3339))
	} else {
		fmt.Printf("❌ Deletion failed with status: %d\n", deleteResp.StatusCode)
		// Read error response for debugging
		if body, err := ioutil.ReadAll(deleteResp.Body); err == nil {
			fmt.Printf("Error details: %s\n", string(body))
		}
	}
}
