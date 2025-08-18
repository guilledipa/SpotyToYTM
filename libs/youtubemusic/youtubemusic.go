// Package youtubemusic provides a client for interacting with the YouTube Music API.
package youtubemusic

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Client is a wrapper around the YouTube Music client.
type Client struct {
	service *youtube.Service
}

// NewClient creates a new YouTube Music client.
func NewClient(ctx context.Context, clientSecretFile string) (*Client, error) {
	b, err := os.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, youtube.YoutubeScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve YouTube client: %v", err)
	}

	return &Client{service: service}, nil
}

// getClient retrieves a token, saves the token, then returns the generated client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb requests a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the " +
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// tokenFromFile retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// CreatePlaylist creates a new playlist on YouTube.
func (c *Client) CreatePlaylist(ctx context.Context, title, description string) (*youtube.Playlist, error) {
	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       title,
			Description: description,
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "private",
		},
	}

	call := c.service.Playlists.Insert([]string{"snippet", "status"}, playlist)
	newPlaylist, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("could not create playlist: %w", err)
	}
	return newPlaylist, nil
}

// SearchAndAddSong searches for a song and adds it to a playlist.
func (c *Client) SearchAndAddSong(ctx context.Context, playlistID, query string) (*youtube.PlaylistItem, error) {
	// Search for the song
	searchCall := c.service.Search.List([]string{"snippet"}).Q(query).MaxResults(1).Type("video")
	searchResponse, err := searchCall.Do()
	if err != nil {
		return nil, fmt.Errorf("could not search for song: %w", err)
	}

	if len(searchResponse.Items) == 0 {
		return nil, fmt.Errorf("no results found for query: %s", query)
	}

	videoID := searchResponse.Items[0].Id.VideoId

	// Add the song to the playlist
	playlistItem := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistID,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoID,
			},
		},
	}

	insertCall := c.service.PlaylistItems.Insert([]string{"snippet"}, playlistItem)
	insertedItem, err := insertCall.Do()
	if err != nil {
		return nil, fmt.Errorf("could not add song to playlist: %w", err)
	}

	return insertedItem, nil
}