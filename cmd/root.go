package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spotytoytm",
	Short: "A tool to migrate playlists from Spotify to YouTube Music",
	Long:  `A tool to migrate playlists from Spotify to YouTube Music.
It allows you to first prepare the migration by fetching all your Spotify playlists
and saving them locally. Then you can migrate them to YouTube Music.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no subcommand is provided
		fmt.Println("Please use one of the available subcommands: prepare, migrate")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
