package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	authorizeURL = "https://api.notion.com/v1/oauth/authorize"
	tokenURL     = "https://api.notion.com/v1/oauth/token"
	callbackPort = 9876
	callbackPath = "/callback"
)

// OAuthConfig holds the OAuth client credentials.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// NewOAuthConfig creates a config with the default redirect URI.
func NewOAuthConfig(clientID, clientSecret string) *OAuthConfig {
	return &OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  fmt.Sprintf("http://localhost:%d%s", callbackPort, callbackPath),
	}
}

// tokenResponse from Notion's OAuth token endpoint.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	BotID        string `json:"bot_id"`
	WorkspaceID  string `json:"workspace_id"`
}

// Login performs the OAuth flow:
// 1. Opens browser to Notion's authorize URL
// 2. Listens on localhost for the callback
// 3. Exchanges the code for a token
// 4. Saves the token to disk
func (c *OAuthConfig) Login(ctx context.Context) (*TokenData, error) {
	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Build authorize URL
	params := url.Values{
		"client_id":     {c.ClientID},
		"redirect_uri":  {c.RedirectURI},
		"response_type": {"code"},
		"owner":         {"user"},
		"state":         {state},
	}
	authURL := authorizeURL + "?" + params.Encode()

	// Start callback server
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", callbackPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", callbackPort, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("state mismatch")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errCh <- fmt.Errorf("authorization error: %s", errMsg)
			fmt.Fprintf(w, "<html><body><h2>Authorization failed: %s</h2><p>You can close this tab.</p></body></html>", errMsg)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}

		fmt.Fprint(w, "<html><body><h2>Login successful!</h2><p>You can close this tab and return to the terminal.</p></body></html>")
		codeCh <- code
	})

	server := &http.Server{Handler: mux}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	// Open browser
	fmt.Println("Opening browser for Notion authorization...")
	fmt.Printf("If the browser doesn't open, visit:\n%s\n\n", authURL)
	openBrowser(authURL)

	fmt.Println("Waiting for authorization...")

	// Wait for callback
	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authorization timed out after 5 minutes")
	}

	// Exchange code for token
	token, err := c.exchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	data := &TokenData{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		WorkspaceID:  token.WorkspaceID,
		BotID:        token.BotID,
	}

	if err := SaveToken(data); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return data, nil
}

// exchangeCode exchanges the authorization code for an access token.
func (c *OAuthConfig) exchangeCode(ctx context.Context, code string) (*tokenResponse, error) {
	body := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {c.RedirectURI},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.ClientID, c.ClientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var token tokenResponse
	if err := json.Unmarshal(respBody, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &token, nil
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}
