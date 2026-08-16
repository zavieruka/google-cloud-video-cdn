# Video Platform web app

An intentionally small Next.js 16 example consumer for the Go Video Platform API.
It demonstrates the real browser flow: request a signed upload URL, upload
directly to Cloud Storage, confirm the upload, watch processing status, and play
the returned HLS manifest.

## Local setup

1. Install dependencies:

   ```bash
   npm ci
   ```

2. Copy the environment example and point it at the running Go API:

   ```bash
   cp .env.example .env.local
   ```

   ```dotenv
   NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
   ```

3. Configure the API and processed bucket for this browser origin:

   ```dotenv
   CORS_ALLOWED_ORIGINS=http://localhost:3000
   ```

   Apply the same origin to the processed bucket with
   `backend/scripts/setup-buckets.sh`. HLS playlists come from the API, while
   the browser fetches signed HLS segments and uploads custom thumbnails
   directly to Cloud Storage.

4. Start the app:

   ```bash
   npm run dev
   ```

## Quality checks

```bash
npm run check
npm run build
```

`npm run check` runs ESLint, TypeScript, and the unit/component suite. Node.js
20.9 or later is required by Next.js 16.

## Browser smoke test (manual review)

This remains unverified until it is run against a configured API, processor, and
GCP buckets.

1. Open `http://localhost:3000` and confirm the list loads.
2. Upload a small MP4 with a title. Confirm the upload progress reaches 100% and
   the detail page opens.
3. Observe `uploaded`/`processing` while the processor runs, then `ready`.
4. In Chrome, play the HLS video and verify DevTools shows playlists from the API
   and signed segments from Cloud Storage with no CORS failures.
5. Repeat playback in Safari to exercise native HLS.
6. Delete the ready video, confirm it leaves the list, and verify the API reports
   an error rather than deleting while a video is still `processing`.
7. Upload a JPEG, PNG, or WebP custom thumbnail (at least 1280×720, at most 10
   MB), confirm it displays clearly, then select a generated candidate to switch
   back.
