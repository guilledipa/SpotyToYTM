// Package spotify provides a client and functions for using the Spotify API.
package spotify

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// Client is a wrapper around the Spotify client.
type Client struct {
	client *spotify.Client
}

// Artist represents a single artist.
type Artist struct {
	Name string `json:"name"`
}

// Album represents a single album.
type Album struct {
	Name    string   `json:"name"`
	Artists []Artist `json:"artists"`
}

// Track represents a single track.
type Track struct {
	Name    string   `json:"name"`
	Artists []Artist `json:"artists"`
	Album   Album    `json:"album"`
}

// Playlist represents the data to be saved for each playlist.
type Playlist struct {
	Name   string  `json:"name"`
	Tracks []Track `json:"tracks"`
}

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://localhost:8080/callback"

// NewClient creates a new Spotify client.
func NewClient(ctx context.Context, clientID, clientSecret string) (*Client, error) {
	// The redirect URL must be an exact match of a URL you've registered for
	// your application. Scopes determine which permissions the user is prompted
	// to authorize.
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopePlaylistReadPrivate),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
	)

	state := generateRandomState()
	ch := make(chan *spotify.Client)
	// Closure to capture state and auth.
	// By using a closure, the state variable is captured and available within
	// the callback function's scope.
	callbackHandler := func(w http.ResponseWriter, r *http.Request, auth *spotifyauth.Authenticator, state string) {
		t, err := auth.Token(r.Context(), state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		client := spotify.New(auth.Client(r.Context(), t))
		fmt.Fprintf(w, "Login Completed!")
		ch <- client // Send the client through the channel
	}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		callbackHandler(w, r, auth, state) // Call the closure handler.
	})
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	client := <-ch // Receive from channel after callback is complete
	return &Client{client: client}, nil
}

// AllPlaylists retrieves all private playlists for the current user.
func (c *Client) AllPlaylists(ctx context.Context) ([]spotify.SimplePlaylist, error) {
	playlists, err := c.client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return nil, err
	}
	var allPlaylists []spotify.SimplePlaylist
	for {
		allPlaylists = append(allPlaylists, playlists.Playlists...)
		err = c.client.NextPage(ctx, playlists)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return allPlaylists, nil
}

// AllItemsFromPlaylist retrieves all items (tracks and episodes) from a given playlist ID.
func (c *Client) AllItemsFromPlaylist(ctx context.Context, playlistID spotify.ID) ([]spotify.PlaylistItem, error) {
	var allItems []spotify.PlaylistItem
	limit := 50 // Maximum limit for playlist item retrieval
	for offset := 0; ; offset += limit {
		itemPage, err := c.client.GetPlaylistItems(ctx, playlistID, spotify.Offset(offset), spotify.Limit(limit))
		if err != nil {
			return nil, fmt.Errorf("could not get playlist items: %w", err)
		}
		allItems = append(allItems, itemPage.Items...)
		if itemPage.Next == "" {
			break
		}
	}
	return allItems, nil
}

// PrepareMigration retrieves all playlists and their items and saves them to disk.
func (c *Client) PrepareMigration(ctx context.Context, outputDir string) error {
	playlists, err := c.AllPlaylists(ctx)
	if err != nil {
		return fmt.Errorf("could not get playlists: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("could not create output directory: %w", err)
	}

	for _, p := range playlists {
		items, err := c.AllItemsFromPlaylist(ctx, p.ID)
		if err != nil {
			log.Printf("could not get items for playlist %s: %v. Skipping playlist.", p.Name, err)
			continue
		}

		var tracks []Track
		for _, item := range items {
			if item.Track.Track != nil {
				var artists []Artist
				for _, artist := range item.Track.Track.Artists {
					artists = append(artists, Artist{Name: artist.Name})
				}

				var albumArtists []Artist
				for _, artist := range item.Track.Track.Album.Artists {
					albumArtists = append(albumArtists, Artist{Name: artist.Name})
				}

				tracks = append(tracks, Track{
					Name:    item.Track.Track.Name,
					Artists: artists,
					Album: Album{
						Name:    item.Track.Track.Album.Name,
						Artists: albumArtists,
					},
				})
			}
		}

		playlist := Playlist{
			Name:   p.Name,
			Tracks: tracks,
		}

		// Sanitize playlist name for filename
		safePlaylistName := strings.ReplaceAll(p.Name, "/", "_")
		filePath := fmt.Sprintf("%s/%s.json", outputDir, safePlaylistName)

		file, err := os.Create(filePath)
		if err != nil {
			log.Printf("could not create file for playlist %s: %v", p.Name, err)
			continue
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(playlist); err != nil {
			log.Printf("could not encode playlist %s: %v", p.Name, err)
			continue
		}
	}
	return nil
}

// generateRandomState is a helper function to generate a random state value
// for OAuth.
func generateRandomState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// This should never happen.
		log.Fatalf("Failed to generate random state: %v", err)
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
