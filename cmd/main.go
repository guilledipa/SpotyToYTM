// Package main provides the entry point for the SpotyToYTM CLI.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/guilledipa/SpotyToYTM/libs/spotify"
)

func main() {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET environment variables must be set.")
	}
	ctx := context.Background()
	client, err := spotify.NewClient(ctx, clientID, clientSecret)
	if err != nil {
		log.Fatalf("Could not create Spotify client: %v", err)
	}
	migratedPlaylists, err := client.PrepareMigration(ctx) // Use PrepareMigration
	if err != nil {
		log.Fatalf("Could not get migrated playlists: %v", err)
	}
	for _, mp := range migratedPlaylists {
		playlistJSON, err := json.MarshalIndent(mp.Playlist, "", "  ")
		if err != nil {
			log.Printf("Error marshaling playlist to JSON: %v", err)
			continue
		}
		fmt.Println("Playlist:")
		fmt.Println(string(playlistJSON))

		fmt.Println("\nItems:")
		for _, item := range mp.Items {
			itemJSON, err := json.MarshalIndent(item, "", "  ")
			if err != nil {
				log.Printf("Error marshaling item to JSON: %v", err)
				continue
			}
			fmt.Println(string(itemJSON))
		}
		fmt.Println("------------------") // Separator between playlists
	}

}
