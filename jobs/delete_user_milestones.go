package main

import (
	// "database/sql"
	"fmt"
	// "log"
	"net/http"
	// "os"
	"time"
)

func main() {
	supabaseURL := "http://127.0.0.1:54321"// os.Getenv("SUPABASE_URL")
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0" // os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	loc, _ := time.LoadLocation("Australia/Melbourne")
	now := time.Now().In(loc)
	weekday := int(now.Weekday()) // Sunday = 0
	sunday := now.AddDate(0, 0, -weekday)
	sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 0, 0, 0, 0, loc)

	// Format to ISO for Supabase query
	weekCutoff := sunday.Format("2006-01-02") 

	deleteURL := fmt.Sprintf("%s/rest/v1/user_milestones?week_start=lt.%s", supabaseURL, weekCutoff)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Prefer", "return=representation") // Used to get count of deleted rows

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("✅ user_milestones reset completed for all rows older than %s\n", weekCutoff)
	} else {
		fmt.Printf("❌ Failed to reset user_milestones — status %d\n", resp.StatusCode)
	}
}