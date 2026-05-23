# Headless demo video pipeline

This repository includes a reproducible, unattended demo recorder for the MCP server.

## What it shows

The scenario runs these MCP natural-language queries in order:

1. bushfires around Australia
2. water hazards around Australia
3. earthquakes around Australia
4. storms around Australia

For each query, it calls `atlas_answer_query`, parses returned extent bboxes, and animates them on a 1280x720 map canvas.

## Local run

```bash
npm install
npx playwright install chromium
npm run demo:record
```

Outputs are written to:

- `demo/output/hazards-demo.webm`
- `demo/output/hazards-demo.mp4` (if `ffmpeg` is installed)

## Use a prebuilt MCP binary instead of `go run`

```bash
MCP_COMMAND=/absolute/path/to/daa-mcp npm run demo:record
```

Optional:

```bash
DIGITAL_ATLAS_BASE_URL=https://digital.atlas.gov.au/api/search/v1 npm run demo:record
```

## CI / unattended recording

Use GitHub Actions workflow:

- `.github/workflows/demo-video.yml`

Trigger via **Run workflow** in Actions. It uploads `demo/output/*` as an artifact.

