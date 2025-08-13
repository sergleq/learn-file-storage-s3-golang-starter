# Copilot Instructions for Tubely (learn-file-storage-s3-golang-starter)

## Project Overview
- This is a Go backend for "Tubely", a video/image upload and streaming service, used in a Boot.dev course on S3/CDN.
- The app exposes HTTP endpoints for user authentication, video/image upload, and metadata management.
- Media files are processed (e.g., for fast start streaming) and uploaded to AWS S3. Metadata is stored in SQLite (`tubely.db`).
- The `app/` directory contains a minimal frontend (HTML/CSS/JS) for interacting with the backend.

## Key Components
- `main.go`: Entry point, sets up HTTP routes and server config.
- `internal/`:
  - `auth/`: JWT auth, bearer token extraction, user validation.
  - `database/`: SQLite DB access for users, videos, refresh tokens.
- `handler_*.go`: HTTP handlers for login, upload, refresh, video meta, etc.
- `assets.go`: Asset management logic.
- `cache.go`: Caching layer (if used).
- `samples/` and `assets/`: Sample and uploaded media files.

## Developer Workflows
- **Install dependencies:** `go mod download`, ensure `ffmpeg`, `ffprobe`, and `sqlite3` are installed and in `PATH`.
- **Download samples:** `./samplesdownload.sh` populates `samples/`.
- **Run server:** `go run .` (creates `tubely.db` and `assets/` if missing).
- **Environment:** Copy `.env.example` to `.env` and fill in as needed (AWS keys, JWT secret, etc.).
- **Testing:** No formal test suite; manual testing via frontend or API tools (e.g., curl, Postman).

## Project-Specific Patterns
- **Video Upload:**
  - Handlers (e.g., `handler_upload_video.go`) validate JWT, check video ownership, process video with ffmpeg, and upload to S3.
  - Video aspect ratio is determined for S3 key prefixing.
  - Fast start processing is required for streaming (see `processVideoForFastStart`).
- **Error Handling:** Uses `respondWithError` for consistent API error responses.
- **UUIDs:** Used for user, video, and asset identification (see `github.com/google/uuid`).
- **S3 Integration:** Uses AWS SDK v2 for Go. Bucket/key naming conventions are enforced in handlers.

## Conventions & Integration
- All HTTP handlers are in root with `handler_*.go` naming.
- Internal logic is under `internal/` by concern (auth, db).
- Media processing relies on external tools (`ffmpeg`, `ffprobe`).
- S3 bucket name is hardcoded as `tubely-8531` (update as needed).
- Minimal/no test automation; manual/interactive testing is expected.

## Examples
- See `handler_upload_video.go` for the full upload/processing/S3 flow.
- See `internal/database/videos.go` for DB video model and queries.
- See `internal/auth/auth.go` for JWT validation and extraction.

---

For questions about project structure or workflows, see `README.md` or the Boot.dev course materials.
