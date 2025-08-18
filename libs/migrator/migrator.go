// Package migrator provides functions to migrate playlists from Spotify to
// YouTube Music.
package migrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/guilledipa/SpotyToYTM/libs/spotify"
	"github.com/guilledipa/SpotyToYTM/libs/youtubemusic"
)

// Migrate reads playlist files from a directory and migrates them to YouTube Music.
func Migrate(ctx context.Context, ytClient *youtubemusic.Client, playlistsDir string) error {
	files, err := os.ReadDir(playlistsDir)
	if err != nil {
		return fmt.Errorf("could not read playlists directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		log.Printf("Migrating playlist from file: %s", file.Name())

		filePath := filepath.Join(playlistsDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("could not read file %s: %v. Skipping.", filePath, err)
			continue
		}

		var playlist spotify.Playlist
		if err := json.Unmarshal(data, &playlist); err != nil {
			log.Printf("could not unmarshal playlist from file %s: %v. Skipping.", filePath, err)
			continue
		}

		// Create a new playlist on YouTube Music.
		ytPlaylist, err := ytClient.CreatePlaylist(ctx, playlist.Name, "Migrated from Spotify")
		if err != nil {
			log.Printf("could not create YouTube Music playlist for %s: %v. Skipping.", playlist.Name, err)
			continue
		}
		log.Printf("Created YouTube Music playlist: %s (ID: %s)", ytPlaylist.Snippet.Title, ytPlaylist.Id)

		// Add tracks to the new playlist.
		for _, track := range playlist.Tracks {
			var artists []string
			for _, artist := range track.Artists {
				artists = append(artists, artist.Name)
			}
			query := fmt.Sprintf("%s %s", track.Name, strings.Join(artists, " "))

			log.Printf("Searching for track: %s", query)
			_, err := ytClient.SearchAndAddSong(ctx, ytPlaylist.Id, query)
		if err != nil {
				log.Printf("could not add track '%s' to playlist '%s': %v", query, ytPlaylist.Snippet.Title, err)
				// Continue to the next track
			} else {
				log.Printf("Successfully added track '%s' to playlist '%s'", query, ytPlaylist.Snippet.Title)
			}
		}
	}

	return nil
}