package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	authURL      = "https://login.yotoplay.com/authorize"
	tokenURL     = "https://login.yotoplay.com/oauth/token"
	audience     = "https://api.yotoplay.com"
	scope        = "offline_access"
	callbackPort = "8081"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
}

var (
	codeVerifier string
	authCode     chan string
)

func main() {
	clientID := os.Getenv("YOTO_CLIENT_ID")
	if clientID == "" {
		clientID = "qRdsgw6mmhaTWPvauY1VyE3Mkx64yaHU"
	}

	fmt.Println("🔐 Yoto Browser-Based Authentication with PKCE")
	fmt.Println("==============================================")
	fmt.Printf("Client ID: %s\n\n", clientID)

	// Generate PKCE challenge
	codeVerifier = generateCodeVerifier()
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Set up callback server
	authCode = make(chan string)
	redirectURI := fmt.Sprintf("http://localhost:%s/callback", callbackPort)
	
	// Start local server for callback
	go startCallbackServer()

	// Build authorization URL
	params := url.Values{}
	params.Set("audience", audience)
	params.Set("scope", scope)
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")
	params.Set("redirect_uri", redirectURI)

	authFullURL := fmt.Sprintf("%s?%s", authURL, params.Encode())

	fmt.Println("📱 Opening browser for authentication...")
	fmt.Printf("If browser doesn't open, visit:\n%s\n\n", authFullURL)

	// Open browser
	openBrowser(authFullURL)

	// Wait for callback
	fmt.Println("⏳ Waiting for authorization...")
	code := <-authCode

	if code == "" {
		log.Fatal("No authorization code received")
	}

	fmt.Printf("\n✅ Received authorization code: %s...\n", code[:10])

	// Exchange code for tokens
	tokens, err := exchangeCodeForTokens(clientID, code, codeVerifier, redirectURI)
	if err != nil {
		log.Fatalf("Failed to exchange code for tokens: %v", err)
	}

	// Display tokens
	fmt.Println("\n✅ Successfully obtained tokens!")
	fmt.Println("=====================================")
	fmt.Printf("\nAccess Token (first 50 chars):\n%s...\n", tokens.AccessToken[:min(50, len(tokens.AccessToken))])
	fmt.Printf("\nRefresh Token (first 50 chars):\n%s...\n", tokens.RefreshToken[:min(50, len(tokens.RefreshToken))])
	fmt.Printf("\nExpires in: %d seconds\n", tokens.ExpiresIn)

	// Update .env file
	updateEnvFile(tokens)

	// Generate Cloud Run commands
	generateCloudRunCommands(tokens)
}

func generateCodeVerifier() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateCodeChallenge(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func startCallbackServer() {
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}

		// Send success page to browser
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Successful</title>
    <style>
        body {
            font-family: system-ui, -apple-system, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
        }
        .container {
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            text-align: center;
            max-width: 500px;
        }
        h1 { color: #2d3748; }
        .success { color: #48bb78; }
    </style>
</head>
<body>
    <div class="container">
        <h1 class="success">✅ Authentication Successful!</h1>
        <p>You can close this window and return to your terminal.</p>
    </div>
</body>
</html>`)

		// Send code to main goroutine
		authCode <- code
	})

	fmt.Printf("Starting callback server on http://localhost:%s/callback\n", callbackPort)
	if err := http.ListenAndServe(":"+callbackPort, nil); err != nil {
		log.Fatal(err)
	}
}

func exchangeCodeForTokens(clientID, code, verifier, redirectURI string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("code_verifier", verifier)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("Error opening browser: %v\n", err)
	}
}

func updateEnvFile(tokens *TokenResponse) {
	fmt.Println("\n📝 Updating .env file...")
	
	// Read existing .env
	envContent, err := os.ReadFile(".env")
	if err != nil {
		fmt.Printf("Warning: Could not read .env file: %v\n", err)
		return
	}

	// Backup current .env
	backupName := fmt.Sprintf(".env.backup.%s", time.Now().Format("20060102_150405"))
	os.WriteFile(backupName, envContent, 0644)

	// Update tokens
	lines := strings.Split(string(envContent), "\n")
	var newLines []string
	foundAccess := false
	foundRefresh := false

	for _, line := range lines {
		if strings.HasPrefix(line, "YOTO_ACCESS_TOKEN=") {
			newLines = append(newLines, fmt.Sprintf("YOTO_ACCESS_TOKEN=%s", tokens.AccessToken))
			foundAccess = true
		} else if strings.HasPrefix(line, "YOTO_REFRESH_TOKEN=") {
			newLines = append(newLines, fmt.Sprintf("YOTO_REFRESH_TOKEN=%s", tokens.RefreshToken))
			foundRefresh = true
		} else {
			newLines = append(newLines, line)
		}
	}

	// Add tokens if not found
	if !foundAccess || !foundRefresh {
		newLines = append(newLines, "")
		newLines = append(newLines, fmt.Sprintf("# Yoto Tokens (updated %s)", time.Now().Format("Mon Jan 02 15:04:05 MST 2006")))
		if !foundAccess {
			newLines = append(newLines, fmt.Sprintf("YOTO_ACCESS_TOKEN=%s", tokens.AccessToken))
		}
		if !foundRefresh {
			newLines = append(newLines, fmt.Sprintf("YOTO_REFRESH_TOKEN=%s", tokens.RefreshToken))
		}
	}

	// Write updated .env
	err = os.WriteFile(".env", []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		fmt.Printf("Error updating .env file: %v\n", err)
	} else {
		fmt.Println("✅ Updated .env file")
	}
}

func generateCloudRunCommands(tokenResp *TokenResponse) {
	fmt.Println("\n🚀 Cloud Run Update Commands")
	fmt.Println("=====================================")
	fmt.Printf(`
# Update Cloud Run with tokens:
gcloud run services update bird-song-explorer \
  --region us-central1 \
  --update-env-vars \
    "YOTO_ACCESS_TOKEN=%s,\
YOTO_REFRESH_TOKEN=%s" \
  --quiet

`, tokenResp.AccessToken, tokenResp.RefreshToken)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}