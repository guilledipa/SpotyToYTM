// Package spotify provides a client and functions for using the Spotify API.
package spotify

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type Client struct {
	client *spotify.Client
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
	go http.ListenAndServe(":8080", nil)
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	client := <-ch // Receive from channel after callback is complete
	return &Client{client: client}, nil
}

// GetPlaylists retrieves all private playlists for the current user.
func (c *Client) GetPlaylists(ctx context.Context) ([]spotify.SimplePlaylist, error) {
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
