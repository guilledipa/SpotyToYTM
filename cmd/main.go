// Package main provides the entry point for the SpotyToYTM CLI.
package main

import (
	"context"
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
	playlists, err := client.GetPlaylists(ctx)
	if err != nil {
		log.Fatalf("Could not get playlists: %v", err)
	}
	fmt.Println("Your Spotify Playlists:")
	for _, p := range playlists {
		fmt.Printf("- %s (ID: %s)\n", p.Name, p.ID)
	}

}
