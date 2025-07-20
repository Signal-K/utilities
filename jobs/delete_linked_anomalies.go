package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	url := supabaseUrl + "/rest/v1/linked_anomalies?author=gt.00000000-0000-0000-0000-000000000000"

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Prefer", "return=representation")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("✅ linked_anomalies cleared successfully.")
	} else {
		fmt.Printf("❌ Failed with status: %d\n", resp.StatusCode)
	}
}