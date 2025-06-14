package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

// makeOauthClient creates a new http.Client with oauth2 set up from the
// given config.
func makeOauthClient(config *oauth2.Config, tokenPath string) *http.Client {
	tok, err := loadCachedToken(tokenPath)
	if err != nil || !tok.Valid() {
		tok = getTokenFromWeb(config)
		saveCachedToken(tokenPath, tok)
	}
	return config.Client(context.Background(), tok)
}

// authenticateUser launches a web browser to authenticate the user vs. Google's
// auth server and returns the auth code that can then be exchanged for tokens.
func authenticateUser(config *oauth2.Config) string {
	const redirectPath = "/redirect"
	// We spin up a goroutine with a web server listening on the redirect route,
	// which the auth server will redirect the user's browser to after
	// authentication.
	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	// When the web server receives redirection, it sends the code to codeChan.
	codeChan := make(chan string)
	var srv http.Server

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc(redirectPath, func(w http.ResponseWriter, req *http.Request) {
			codeChan <- req.URL.Query().Get("code")
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintln(w, "<h3>Authenticated! You can now close this page.</h3>")
		})
		srv.Handler = mux
		if err := srv.Serve(listener); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	config.RedirectURL = fmt.Sprintf("http://localhost:%d%s", port, redirectPath)
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Click this link to authenticate:\n", authURL)

	// Receive code from the web server and shut it down.
	authCode := <-codeChan
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}

	return authCode
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authCode := authenticateUser(config)
	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("unable to retrieve token from web: %v", err)
	}
	return tok
}

// loadCachedToken tries to load a cached token from a local file.
func loadCachedToken(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveCachedToken saves an oauth2 token to a local file.
func saveCachedToken(path string, token *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("unable to cache OAuth token: %v", err)
	}
	defer f.Close()

	json.NewEncoder(f).Encode(token)
}

func deleteCachedToken(path string) {
	err := os.Remove(path)
	if err != nil {
		log.Fatalf("unable to delete cached OAuth token: %v", err)
	}
}
