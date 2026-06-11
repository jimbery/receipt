package mail

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// ErrTokenExpired indicates the pilot OAuth access token is past its expiry.
var ErrTokenExpired = errors.New("pilot oauth access token expired")

// PilotToken is a pre-obtained OAuth access token stored locally for pilot runs.
// Volunteers obtain tokens outside CI via the flow in docs/development/pilot-mail-oauth.md.
type PilotToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	Expiry      time.Time `json:"expiry"`
}

// LoadPilotToken reads a pilot token JSON file from disk.
func LoadPilotToken(path string) (PilotToken, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PilotToken{}, fmt.Errorf("read pilot token: %w", err)
	}
	var tok PilotToken
	if unmarshalErr := json.Unmarshal(data, &tok); unmarshalErr != nil {
		return PilotToken{}, fmt.Errorf("parse pilot token: %w", unmarshalErr)
	}
	if tok.AccessToken == "" {
		return PilotToken{}, errors.New("pilot token missing access_token")
	}
	if tok.TokenType == "" {
		tok.TokenType = "Bearer"
	}
	return tok, nil
}

func (t PilotToken) expired(now time.Time) bool {
	return !t.Expiry.IsZero() && !now.Before(t.Expiry)
}

func (t PilotToken) authHeader() string {
	return t.TokenType + " " + t.AccessToken
}
