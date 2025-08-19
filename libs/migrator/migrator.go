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

// FailedTracksMap stores tracks that failed to migrate, mapped by YouTube Music Playlist ID.
type FailedTracksMap map[string][]spotify.Track

// Migrate reads playlist files from a directory and migrates them to YouTube Music.
func Migrate(ctx context.Context, ytClient *youtubemusic.Client, playlistsDir string) error {
	files, err := os.ReadDir(playlistsDir)
	if err != nil {
		return fmt.Errorf("could not read playlists directory: %w", err)
	}

	failedTracks := make(FailedTracksMap)

	for _, file := range files {
		if file.IsDir() { // Skip directories
			continue
		}
		if filepath.Ext(file.Name()) != ".json" { // Only process .json files
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

		if playlist.Name == "" { // Validate playlist name
			log.Printf("Playlist name is empty in file %s. Skipping.", filePath)
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
				failedTracks[ytPlaylist.Id] = append(failedTracks[ytPlaylist.Id], track)
			} else {
				log.Printf("Successfully added track '%s' to playlist '%s'", query, ytPlaylist.Snippet.Title)
			}
		}
	}

	// Save failed tracks to a JSON file
	if len(failedTracks) > 0 {
		failedTracksFile, err := os.Create("failed_tracks.json")
		if err != nil {
			return fmt.Errorf("could not create failed_tracks.json: %w", err)
		}
		defer failedTracksFile.Close()

		encoder := json.NewEncoder(failedTracksFile)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(failedTracks); err != nil {
			return fmt.Errorf("could not encode failed tracks: %w", err)
		}
		log.Println("Failed tracks saved to failed_tracks.json")
	} else {
		log.Println("No tracks failed to migrate.")
	}

	return nil
}
