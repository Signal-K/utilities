package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type LinkedAnomaly struct {
	ID              int64     `json:"id"`
	Author          string    `json:"author"`
	AnomalyID       int64     `json:"anomaly_id"`
	ClassificationID *int64   `json:"classification_id"`
	Date            time.Time `json:"date"`
	Automaton       *string   `json:"automaton"`
}

func main() {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	// Step 1: Fetch existing linked anomalies
	req, err := http.NewRequest("GET", supabaseUrl+"/rest/v1/linked_anomalies?select=*", nil)
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
		fmt.Printf("❌ Failed to fetch records, status: %d\n", resp.StatusCode)
		return
	}

	var anomalies []LinkedAnomaly
	if err := json.NewDecoder(resp.Body).Decode(&anomalies); err != nil {
		panic(err)
	}

	// Step 2: Write to logs/linked_anomalies_<timestamp>.json
	if len(anomalies) > 0 {
		timestamp := time.Now().Format("2006-01-02T15-04-05")
		filename := fmt.Sprintf("logs/linked_anomalies_%s.json", timestamp)

		// Create logs dir if it doesn't exist
		if _, err := os.Stat("logs"); os.IsNotExist(err) {
			os.Mkdir("logs", 0755)
		}

		data, err := json.MarshalIndent(anomalies, "", "  ")
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			panic(err)
		}
		fmt.Println("✅ Backup log written to", filename)
	} else {
		fmt.Println("ℹ️ No linked anomalies to back up.")
	}

	// Step 3: Delete all linked anomalies
	deleteReq, err := http.NewRequest("DELETE", supabaseUrl+"/rest/v1/linked_anomalies?id=gt.0", nil)
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
		fmt.Println("✅ linked_anomalies cleared successfully.")
	} else {
		fmt.Printf("❌ Deletion failed with status: %d\n", deleteResp.StatusCode)
	}
}