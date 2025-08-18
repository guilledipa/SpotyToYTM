// Package main provides the entry point for the SpotyToYTM CLI.
package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/guilledipa/SpotyToYTM/libs/spotify"
)

var outputDir = flag.String("output-dir", "./playlists", "Directory to save playlists")

func main() {
	flag.Parse()

	ctx := context.Background()
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET environment variables must be set.")
	}

	log.Println("To authenticate, please open the URL printed below in your web browser.")
	client, err := spotify.NewClient(ctx, clientID, clientSecret)
	if err != nil {
		log.Fatalf("Could not create Spotify client: %v", err)
	}

	log.Printf("Authentication successful! Pulling playlists and saving them to %q directory...", *outputDir)
	if err := client.PrepareMigration(ctx, *outputDir); err != nil {
		log.Fatalf("Could not prepare migration: %v", err)
	}

	log.Println("Playlists saved successfully.")
}
