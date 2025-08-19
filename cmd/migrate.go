package cmd

import (
	"context"
	"log"

	"github.com/guilledipa/SpotyToYTM/libs/migrator"
	"github.com/guilledipa/SpotyToYTM/libs/youtubemusic"
	"github.com/spf13/cobra"
)

var (
	clientSecretFile string
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrates playlists from local files to YouTube Music.",
	Long:  `Reads the playlist files from the 'playlists' directory, and for each playlist, it creates a new one in YouTube Music and adds the tracks.`,
	Run: func(cmd *cobra.Command, args []string) {
		migrate()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVarP(&clientSecretFile, "client-secret", "c", "client_secret.json", "Path to the client_secret.json file for YouTube Music API authentication.")
}

func migrate() {
	ctx := context.Background()
	log.Println("Starting migration to YouTube Music...")

	ytClient, err := youtubemusic.NewClient(ctx, clientSecretFile) // Use the flag value
	if err != nil {
		log.Fatalf("Could not create YouTube Music client: %v", err)
	}

	if err := migrator.Migrate(ctx, ytClient, "playlists"); err != nil {
		log.Fatalf("Could not migrate playlists: %v", err)
	}

	log.Println("Migration completed successfully.")
}