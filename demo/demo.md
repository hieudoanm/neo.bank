# NEO Bank — End-to-End Demo

This folder holds the scripted end-to-end demo of NEO Bank and its recording.

- **`record-demo.sh`** — boots the whole stack, drives the UI with Playwright, and records **`demo.mp4`**.
- **`demo.mp4`** — the recorded walkthrough (generated; see below).

The demo proves the money path works end to end across the real services: a
customer registers (which triggers account provisioning over Kafka), sends funds
to another account, the Transfer Service saga runs asynchronously, and both
balances plus the transaction histories update.

## The scenario

Playwright drives a single browser session through the React app (served by the
frontend container at `http://localhost:3000`), which talks to the API gateway
at `http://localhost:8080` — the same path a real user takes.

1. Register a **recipient**; the Account Service provisions its account from the
   `auth.user.created` event.
2. Register a **sender**; registration signs them in and provisions an account
   funded with the configured opening balance (`$1,000.00` locally).
3. As the sender, transfer **$25.00** to the recipient's account ID.
4. Wait for the saga to complete — the status page polls until the transfer is
   `completed`.
5. Assert the sender's balance is now **`$975.00`** and the debit appears in the
   sender's transaction history.
6. Sign in as the recipient, assert the balance is **`$1,025.00`**, and confirm
   the credit in their history.

The spec lives in
[`packages/app/e2e/demo.spec.ts`](../packages/app/e2e/demo.spec.ts).

## Prerequisites

- Docker with the Compose v2 plugin (`docker compose`).
- Node and pnpm (the repo pins `pnpm` and Node in `packages/app/package.json`).
- `ffmpeg` on `PATH` (or Playwright's bundled copy, which the script falls back to).
- A local `.env` with a JWT key pair — the services refuse to start without it:

  ```bash
  cp .env.example .env
  # Generate a throwaway dev RSA key pair, then paste it into .env
  # (see the instructions at the top of .env.example).
  openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out private.pem
  openssl rsa -in private.pem -pubout -out public.pem
  ```

  The demo also relies on the opening balance, which `docker-compose.yaml`
  sets to `INITIAL_BALANCE_CENTS=100000` (`$1,000.00`) for the account
  consumer. Override it in `.env` if you like.

## Record the demo

From the repository root:

```bash
demo/record-demo.sh
```

The script:

1. runs `docker compose up -d --build`;
2. waits until the backend really works (it polls `POST /api/v1/register` until
   it returns `201`/`409` rather than a boot-time `5xx`);
3. installs Playwright's Chromium if needed and runs the demo spec, recording
   the browser (the video is written under
   `packages/app/test-results/`);
4. converts the recording to **`demo/demo.mp4`** with `ffmpeg`;
5. tears the stack down.

Useful knobs:

| Variable          | Default                 | Purpose                                            |
| ----------------- | ----------------------- | -------------------------------------------------- |
| `KEEP_STACK=1`    | unset (stack is removed) | Leave the stack running after recording            |
| `DEMO_SLOWMO`     | `300`                   | Milliseconds of slow-motion per action (bigger = clearer video) |
| `DEMO_BASE_URL`   | `http://localhost:3000` | Frontend URL Playwright drives                     |
| `DEMO_GATEWAY_URL`| `http://localhost:8080` | Gateway URL used by the readiness probe            |

Examples:

```bash
KEEP_STACK=1 DEMO_SLOWMO=600 demo/record-demo.sh
```

## Run just the E2E test

With a stack already running (e.g. start it yourself and pass `KEEP_STACK=1`):

```bash
cd packages/app
pnpm e2e:install   # first time only
pnpm e2e           # or: pnpm exec playwright test
```

Playwright records every run, so `packages/app/test-results/` always
holds the latest `.webm`. Convert it by hand if you want:

```bash
ffmpeg -y -i packages/app/test-results/*/video.webm \
  -c:v libx264 -pix_fmt yuv420p -movflags +faststart demo/demo.mp4
```

## Notes

- The E2E spec is **not** part of `pnpm test` (Vitest) or the coverage gate; it
  needs the full stack, so it is run explicitly via `pnpm e2e`.
- The frontend image never downloads Playwright browsers
  (`PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1` in the `Dockerfile`); the browser runs on
  the host.
- Registration generates a unique email each run, so repeated demos don't
  collide on the persisted Cassandra volume.
