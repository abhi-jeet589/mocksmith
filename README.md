# mocksmith

A self-hosted mock API service used for local development purposes. Define HTTP endpoints from a web UI, and
mocksmith serves them back to your clients at the same path — no prefix, no
client-side awareness needed. Point any application at `http://localhost:8080`
and it can hit your mocks as if they were the real upstream.

Runs entirely inside Docker. The backend containers have no internet egress.

## Layout

```
mocksmith/
├── docker-compose.yml      orchestrates api + mongo + app
├── Makefile                dev shortcuts
├── api/                    Go service (chi + mongo-driver v2)
└── app/                    static frontend served by nginx
```

## Quickstart

```bash
cp .env.example .env
make rebuild
```

Open <http://localhost:8080/> and create a mock through the UI. The `app`
container is the only thing published to the host (loopback only, port 8080
by default).

## Using a mock

Whatever method + path you registered, your client uses verbatim:

```bash
curl -i http://localhost:8080/users/42
```

```python
import requests
r = requests.get("http://localhost:8080/users/42")
```

Match is exact: method, path, and trailing slash all count.

### Path parameters

A path may contain `{name}` tokens. For every token you must declare its
data type — the UI surfaces a row per token as you type. Supported types:

| Type     | Matches                                             |
| -------- | --------------------------------------------------- |
| `string` | one path segment, any characters except `/`         |
| `int`    | optional leading `-` followed by digits             |
| `uuid`   | canonical 8-4-4-4-12 hex UUID                       |

Example: registering `GET /users/{id}` with `{"id": "int"}` matches
`/users/42` but not `/users/abc`.

**Lookup precedence:** exact matches always win over pattern matches. Among
patterns the first registered wins — if your patterns can overlap, register
more specific ones first or rely on the order shown in the UI. Captured
values aren't yet exposed in the response body — that's the next iteration.

## Admin API

The UI is a thin client over these endpoints. They're proxied through nginx
at the same host port.

| Method | Path                | Purpose        |
| ------ | ------------------- | -------------- |
| GET    | `/health`           | Liveness check |
| GET    | `/admin/mocks`      | List all mocks |
| POST   | `/admin/mocks`      | Create a mock  |
| GET    | `/admin/mocks/{id}` | Get one        |
| PUT    | `/admin/mocks/{id}` | Replace one    |
| DELETE | `/admin/mocks/{id}` | Delete one     |

Mock payload shape:

```json
{
  "method": "GET",
  "path": "/users/42",
  "statusCode": 200,
  "contentType": "application/json",
  "headers": { "X-Custom": "value" },
  "body": "{\"id\":42,\"name\":\"Ada\"}"
}
```

`(method, path)` is unique — creating a duplicate returns 409.

## Reserved paths

mocksmith claims these paths for itself; you can't register mocks at them:

- `/` — the UI's landing page
- `/index.html`, `/app.js`, `/styles.css` — UI static assets
- `/admin/*` — admin API
- `/health` — health check

Everything else is fair game.

## Network model

Two Docker networks:

- `internal` (`internal: true`) — `mongo`, `api`, `app`. No internet egress.
- `public` (bridge) — `app` only, so it can be reached from the host.

Only the `app` container publishes a host port. `api` and `mongo` are not
reachable from outside the compose stack.

## Common operations

```bash
make up         # start in background
make logs       # tail api logs
make rebuild    # rebuild images and restart
make down       # stop containers, keep data
make clean      # stop and drop the mongo volume
```
