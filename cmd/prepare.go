package cmd

import (
	"context"
	"log"
	"os"

	"github.com/guilledipa/SpotyToYTM/libs/spotify"
	"github.com/spf13/cobra"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Fetches playlists from Spotify and saves them locally.",
	Long:  `Authenticates with the Spotify API, fetches all of the user's playlists, and saves them as individual JSON files in the 'playlists' directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		prepare()
	},
}

func init() {
	rootCmd.AddCommand(prepareCmd)
}

func prepare() {
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

	log.Println("Authentication successful! Pulling playlists and saving them to './playlists' directory...")
	if err := client.PrepareMigration(ctx, "playlists"); err != nil {
		log.Fatalf("Could not prepare migration: %v", err)
	}

	log.Println("Playlists saved successfully.")
}
