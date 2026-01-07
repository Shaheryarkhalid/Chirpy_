# Chirpy API (Go)


## Routes

### GET /api/healthz
Input: None
Output (200):
"OK"

### GET /api/crash
Input: None
Behavior: Panics and crashes server

### POST /api/users
Input (JSON):
{ "email": "string", "password": "string" }
Output (201):
{
  "id": "uuid",
  "email": "string",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "is_chirpy_red": false
}

### POST /api/login
Input (JSON):
{ "email": "string", "password": "string" }
Output (200):
{
  "id": "uuid",
  "email": "string",
  "is_chirpy_red": "bool",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "token": "jwt (1h expiry)",
  "refresh_token": "string"
}

### PUT /api/users
Auth: Bearer JWT required
Input (JSON):
{ "email": "string", "password": "string" }
Output (200):
{
  "id": "uuid",
  "email": "string",
  "is_chirpy_red": "bool",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
Errors:
401 → Invalid/missing token
400 → Invalid JSON/missing fields
500 → DB failure

### POST /api/refresh
Auth: Bearer refresh token required
Input: Authorization header token
Output (200):
{ "token": "new jwt (1h expiry)" }
Errors:
401 → Not found/revoked/expired token
500 → DB failure

### POST /api/revoke
Auth: Bearer refresh token required
Input: Authorization header token
Output (204):
{}
Errors:
401 → Invalid/missing/already revoked token
500 → DB failure

### POST /api/polka/webhooks
Auth: API Key required (must match POLKA_KEY secret)
Input (JSON):
{
  "event": "string",
  "data": { "user_id": "uuid string" }
}
Behavior:
If event != "user.upgraded" → 204 {}
If valid → upgrades user to red
Output (204):
{}
Errors:
401 → Invalid API key
400 → Invalid JSON/UUID
404 → User not found
500 → DB failure

### POST /api/chirps
Auth: Bearer JWT required
Input (JSON):
{ "body": "string (max 140 chars)" }
Output (201):
{
  "id": "uuid",
  "body": "string",
  "user_id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
Errors:
401 → Invalid/missing JWT
400 → Body > 140 or invalid JSON
500 → DB failure

### GET /api/chirps
Query Params (optional):
author_id=uuid → filter by author
sort=asc|desc → sort by created_at
Output (200):
[
  {
    "id": "uuid",
    "body": "string",
    "user_id": "uuid",
    "created_at": "timestamp",
    "updated_at": "timestamp"
  }
]
Sorting:
Default → ascending created_at
sort=desc → descending created_at
Filtering:
If author_id valid and not nil → returns chirps for that author
Else → returns all chirps
Errors:
500 → DB failure

### GET /api/chirps/{chirpID}
Input: chirpID path param (UUID)
Output (200):
{
  "id": "uuid",
  "body": "string",
  "user_id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
Errors:
400 → Invalid/missing UUID
404 → Chirp not found
500 → DB failure

### DELETE /api/chirps/{chirpID}
Auth: Bearer JWT required
Input: chirpID path param (UUID)
Output (204):
{}
Errors:
401 → Invalid JWT or UUID
403 → Chirp not owned by user
404 → ChirpID missing
500 → DB failure

### GET /admin/metrics
Input: None
Output (200):
"<html><body>Chirpy has been visited X times!</body></html>"

### POST /admin/reset
Input: None
Behavior:
Resets hit counter to 0
Clears users only if PLATFORM == "dev"
Output (200):
"Hits: 0"
Errors:
403 → Not dev platform
500 → DB failure

