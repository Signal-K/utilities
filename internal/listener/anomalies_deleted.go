package listener

import (
	"context"
	"log"

	"github.com/supabase-community/realtime-go"
)

func ListenForAnomalyDeletions(ctx context.Context, supabaseUrl, supabaseKey string) error {
	client := realtime.NewClient(supabaseUrl, supabaseKey)

	if err := client.Connect(); err != nil {
		return err
	}
	log.Println("Connected to Supabase realtime")
}
