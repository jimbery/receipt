# Pilot mailbox OAuth (Gmail + Microsoft Graph)

Phase 1 uses **pilot-mode** mail sources: volunteers obtain OAuth tokens locally; CI never calls live APIs.

## Token file format

Store a JSON file outside the repository (e.g. `~/receipt-pilot/gmail-token.json`):

```json
{
  "access_token": "<from OAuth flow>",
  "token_type": "Bearer",
  "expiry": "2026-06-11T12:00:00Z"
}
```

Example committed without secrets: `test/testdata/mail/pilot-token.example.json`.

`internal/mail` loads this file via `NewGmailSource(mailboxID, tokenPath)` or `NewGraphSource(mailboxID, tokenPath)`. An empty `tokenPath` returns `ErrNotConfigured`. An expired token returns `ErrTokenExpired` — refresh the token before fetching.

## Gmail

1. Create a Google Cloud OAuth client (Desktop app) with Gmail read-only scope: `https://www.googleapis.com/auth/gmail.readonly`
2. Run the OAuth consent flow and save the access token (and refresh token on your machine only).
3. Point `GmailSource` at the token file and fetch with `Fetch(ctx, since)`.

API calls: `users.messages.list` (optional `q=after:YYYY/MM/DD`) then `users.messages.get?format=raw` (base64url RFC822).

## Microsoft Graph

1. Register an app in Azure AD with delegated `Mail.Read`.
2. Complete the OAuth flow and save the access token JSON as above.
3. Point `GraphSource` at the token file.

API calls: `GET /v1.0/me/messages` with optional `$filter=receivedDateTime ge …`, then `GET /v1.0/me/messages/{id}/$value` for raw MIME.

## Security

- Never commit token files — they are listed in `.gitignore` (`pilot-token*.json`, `~/receipt-pilot/`).
- Issue tokens only after the volunteer has reviewed [volunteer-data-handling.md](../consent/volunteer-data-handling.md).
