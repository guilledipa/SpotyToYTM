// Package main provides the entry point for the SpotyToYTM CLI.
package main

import (
	"context"
	"encoding/json"
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

	log.Println("To authenticate, please open the URL printed below in your web browser.")
	client, err := spotify.NewClient(ctx, clientID, clientSecret)
	if err != nil {
		log.Fatalf("Could not create Spotify client: %v", err)
	}

	log.Println("Authentication successful! Pulling playlists...")
	migratedPlaylists, err := client.PrepareMigration(ctx) // Use PrepareMigration
	if err != nil {
		log.Fatalf("Could not get migrated playlists: %v", err)
	}

	// Create a temporary file to store the JSON data.
	tempFile, err := os.CreateTemp("", "spotify-playlists-*.json")
	if err != nil {
		log.Fatalf("Failed to create temporary file: %v", err)
	}
	defer tempFile.Close()

	// Write the JSON data to the temporary file.
	if err := json.NewEncoder(tempFile).Encode(migratedPlaylists); err != nil {
		log.Fatalf("Failed to write JSON to temporary file: %v", err)
	}

	// Pretty print the playlists.
	spotify.PrettyPrintPlaylists(migratedPlaylists)

}
