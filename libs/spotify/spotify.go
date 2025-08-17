// Package spotify provides a client and functions for using the Spotify API.
package spotify

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// Client is a wrapper around the Spotify client.
type Client struct {
	client *spotify.Client
}

// MigratePlaylist is a map of playlist names, which will be used in YTM to
// preserve the playlists name, and a slice of playlist items.
type MigratePlaylist struct {
	Playlist spotify.SimplePlaylist
	Items    []spotify.PlaylistItem
}

// MigratePlaylists is a slice of MigratePlaylist.
type MigratePlaylists []MigratePlaylist

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

// PrepareMigration retrieves all playlists and their items into a convenient
// data structure for migration.
func (c *Client) PrepareMigration(ctx context.Context) (MigratePlaylists, error) {
	playlists, err := c.AllPlaylists(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get playlists: %w", err)
	}
	var migratePlaylists MigratePlaylists
	for _, p := range playlists {
		items, err := c.AllItemsFromPlaylist(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("could not get items for playlist %s: %w", p.Name, err)
		}
		migratePlaylists = append(migratePlaylists, MigratePlaylist{
			Playlist: p,
			Items:    items,
		})
	}
	return migratePlaylists, nil
}

// PrettyPrintPlaylists prints the playlists and their items in a human-readable
// format.
func PrettyPrintPlaylists(playlists MigratePlaylists) {
	for _, p := range playlists {
		fmt.Printf("Playlist: %s\n", p.Playlist.Name)
		for _, item := range p.Items {
			track := item.Track.Track
			if track != nil {
				artists := make([]string, len(track.Artists))
				for i, artist := range track.Artists {
					artists[i] = artist.Name
				}
				fmt.Printf("  - %s by %s\n", track.Name, strings.Join(artists, ", "))
			}
		}
	}
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
