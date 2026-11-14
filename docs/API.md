# API

Base path: `/api`. Requests and responses are JSON.

Protected endpoints need the header `Authorization: Bearer <token>`. A token is valid for 24 hours.

Errors look like this:

```json
{ "error": "issue not found" }
```

| Status | Meaning |
| --- | --- |
| 400 | Bad input |
| 401 | No token or a bad token |
| 403 | Not an admin |
| 404 | Not found |
| 409 | Duplicate username or email |

## Health

### GET /api/health

```
curl http://localhost:8080/api/health
```

```json
{ "status": "ok" }
```

## Auth

### POST /api/auth/register

The password needs at least 8 characters. Returns 201.

```
curl -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"serafim","email":"serafim@example.com","password":"password123","languages":["Java","Go"]}'
```

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "6734a1c2e4b0a1b2c3d4e5f6",
    "username": "serafim",
    "email": "serafim@example.com",
    "role": "user",
    "bio": "",
    "languages": ["Java", "Go"],
    "interests": [],
    "createdAt": "2026-11-01T10:00:00Z"
  }
}
```

### POST /api/auth/login

```
curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@issuematch.local","password":"demo12345"}'
```

The response has the same shape as for register.

## Users

### GET /api/users/me

Needs a token.

```
curl http://localhost:8080/api/users/me -H "Authorization: Bearer $TOKEN"
```

```json
{
  "id": "6734a1c2e4b0a1b2c3d4e5f6",
  "username": "demo",
  "email": "demo@issuematch.local",
  "role": "user",
  "bio": "Backend developer",
  "languages": ["Go", "Java"],
  "interests": ["backend", "databases"],
  "createdAt": "2026-11-01T10:00:00Z"
}
```

### PUT /api/users/me

Needs a token. Returns the updated profile.

```
curl -X PUT http://localhost:8080/api/users/me \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"username":"demo","email":"demo@issuematch.local","bio":"Backend developer","languages":["Go"],"interests":["databases"]}'
```

The response has the same shape as `GET /api/users/me`.

## Issues

### GET /api/issues

| Parameter | Meaning |
| --- | --- |
| `q` | Text search in title and description |
| `category` | `bug`, `feature`, `documentation` or `refactoring` |
| `language` | For example `Go` |
| `difficulty` | `easy`, `medium` or `hard` |
| `label` | For example `good first issue` |
| `page` | Default 1 |
| `limit` | Default 20, at most 100 |

With `q` the result is sorted by text score. Without it the newest issues come first.

```
curl 'http://localhost:8080/api/issues?q=memory&category=bug&language=Go&limit=1'
```

```json
{
  "items": [
    {
      "id": "6734a1c2e4b0a1b2c3d4e5f7",
      "title": "Connection pool leaks memory when a request times out",
      "description": "The connection pool leaks memory when a request times out...",
      "category": "bug",
      "language": "Go",
      "difficulty": "medium",
      "labels": ["help wanted"],
      "project": { "name": "northwind/httpclient", "url": "https://github.com/northwind/httpclient", "stars": 4200 },
      "status": "open",
      "createdAt": "2026-10-20T08:30:00Z"
    }
  ],
  "total": 12,
  "page": 1,
  "limit": 1
}
```

Imported issues also have `externalId`, for example `github:golang/go#812`.

### GET /api/issues/:id

```
curl http://localhost:8080/api/issues/6734a1c2e4b0a1b2c3d4e5f7
```

Returns one issue in the shape shown above.

### POST /api/issues

Admin only. Returns 201 with the issue.

```
curl -X POST http://localhost:8080/api/issues \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d '{"title":"Add CSV export","description":"Reports need CSV.","category":"feature","language":"Python","difficulty":"easy","labels":["good first issue"],"project":{"name":"tidewater/reports","url":"https://github.com/tidewater/reports","stars":310},"status":"open"}'
```

```json
{
  "id": "6734a1c2e4b0a1b2c3d4e5f8",
  "title": "Add CSV export",
  "description": "Reports need CSV.",
  "category": "feature",
  "language": "Python",
  "difficulty": "easy",
  "labels": ["good first issue"],
  "project": { "name": "tidewater/reports", "url": "https://github.com/tidewater/reports", "stars": 310 },
  "status": "open",
  "createdAt": "2026-11-12T09:00:00Z"
}
```

### PUT /api/issues/:id

Admin only. Takes the same body as create and returns the updated issue.

```
curl -X PUT http://localhost:8080/api/issues/6734a1c2e4b0a1b2c3d4e5f8 \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d '{"title":"Add CSV export","description":"Reports need CSV.","category":"feature","language":"Python","difficulty":"medium","labels":[],"project":{"name":"tidewater/reports","url":"https://github.com/tidewater/reports","stars":310},"status":"closed"}'
```

### DELETE /api/issues/:id

Admin only. Returns 204 with no body.

```
curl -X DELETE http://localhost:8080/api/issues/6734a1c2e4b0a1b2c3d4e5f8 -H "Authorization: Bearer $ADMIN_TOKEN"
```

## Interactions

### POST /api/interactions

Needs a token. `type` is `view`, `bookmark`, `working` or `solved`. Returns 201. Recording the same interaction again updates `createdAt`.

```
curl -X POST http://localhost:8080/api/interactions \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"issueId":"6734a1c2e4b0a1b2c3d4e5f7","type":"bookmark"}'
```

```json
{
  "id": "6734a1c2e4b0a1b2c3d4e5f9",
  "userId": "6734a1c2e4b0a1b2c3d4e5f6",
  "issueId": "6734a1c2e4b0a1b2c3d4e5f7",
  "type": "bookmark",
  "createdAt": "2026-11-05T14:12:00Z"
}
```

### DELETE /api/interactions/:issueId/:type

Needs a token. Removes one interaction of the current user. Returns 204 with no body.

```
curl -X DELETE http://localhost:8080/api/interactions/6734a1c2e4b0a1b2c3d4e5f7/bookmark -H "Authorization: Bearer $TOKEN"
```

### GET /api/users/me/interactions

Needs a token. Returns the current user's history, newest first.

```
curl http://localhost:8080/api/users/me/interactions -H "Authorization: Bearer $TOKEN"
```

```json
[
  {
    "id": "6734a1c2e4b0a1b2c3d4e5f9",
    "issueId": "6734a1c2e4b0a1b2c3d4e5f7",
    "type": "bookmark",
    "createdAt": "2026-11-05T14:12:00Z",
    "title": "Connection pool leaks memory when a request times out"
  }
]
```

## Recommendations

### GET /api/recommendations

Needs a token. Returns issues picked for the current user by item-based collaborative filtering. A user with no history gets the most popular open issues.

| Parameter | Meaning |
| --- | --- |
| `limit` | Default 10, at most 50 |
| `strategy` | `precomputed` (default) reads stored similarities. `live` computes them during the request |
| `cache` | `on` (default) or `off` |

```
curl 'http://localhost:8080/api/recommendations?limit=1' -H "Authorization: Bearer $TOKEN"
```

```json
{
  "items": [
    {
      "issue": {
        "id": "6734a1c2e4b0a1b2c3d4e5f7",
        "title": "Connection pool leaks memory when a request times out",
        "description": "The connection pool leaks memory when a request times out...",
        "category": "bug",
        "language": "Go",
        "difficulty": "medium",
        "labels": ["help wanted"],
        "project": { "name": "northwind/httpclient", "url": "https://github.com/northwind/httpclient", "stars": 4200 },
        "status": "open",
        "createdAt": "2026-10-20T08:30:00Z"
      },
      "score": 2.41,
      "basedOn": "Scheduler deadlocks under heavy load"
    }
  ],
  "strategy": "precomputed",
  "cached": false
}
```

`basedOn` is the title of the issue from the user's history that added most to the score. It is empty for a user with no history.

An issue that no other user has touched has no neighbors yet. Open issues with the same language and category stand in for them, the same project first.

A rating loses half of its weight for every 7 days since the user last touched the issue. Recent activity shapes the result more.

### GET /api/recommendations/trending

Needs a token. Returns the most active open issues of the last 7 days. Issues in the given languages come first.

| Parameter | Meaning |
| --- | --- |
| `languages` | Comma-separated, for example `Go,Java` |
| `limit` | Default 5, at most 50 |

```
curl 'http://localhost:8080/api/recommendations/trending?languages=Go&limit=1' -H "Authorization: Bearer $TOKEN"
```

The response is `{ "items": [ ... ] }` with the same items as above and an empty `basedOn`.

### POST /api/admin/recompute

Admin only. Runs the similarity job now.

```
curl -X POST http://localhost:8080/api/admin/recompute -H "Authorization: Bearer $ADMIN_TOKEN"
```

```json
{ "issues": 1000, "durationMs": 165 }
```
