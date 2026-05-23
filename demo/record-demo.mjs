import { spawn } from "node:child_process";
import { mkdirSync, copyFileSync, rmSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { chromium } from "playwright";

const __dirname = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(__dirname, "..");
const outputDir = join(repoRoot, "demo", "output");
const tmpVideoDir = join(outputDir, "tmp");
mkdirSync(outputDir, { recursive: true });
rmSync(tmpVideoDir, { recursive: true, force: true });
mkdirSync(tmpVideoDir, { recursive: true });

const scenario = [
  { label: "Bushfires", query: "Where are the bushfires around Australia?", color: "#ff643d" },
  { label: "Water hazards", query: "Where are the water hazards around Australia?", color: "#2bd0ff" },
  { label: "Earthquakes", query: "Where are the earthquakes around Australia?", color: "#c973ff" },
  { label: "Storms", query: "Where are the storms around Australia?", color: "#f8d542" }
];

function delay(ms) {
  return new Promise((r) => setTimeout(r, ms));
}

class MCPClient {
  constructor() {
    const cmd = process.env.MCP_COMMAND;
    const args = process.env.MCP_ARGS ? process.env.MCP_ARGS.split(" ").filter(Boolean) : [];

    if (cmd) {
      this.proc = spawn(cmd, args, { cwd: repoRoot, stdio: ["pipe", "pipe", "pipe"] });
    } else {
      this.proc = spawn("go", ["run", "./cmd/daa-mcp"], { cwd: repoRoot, stdio: ["pipe", "pipe", "pipe"] });
    }

    this.proc.stderr.on("data", (d) => {
      const s = d.toString().trim();
      if (s) process.stderr.write(`[mcp] ${s}\n`);
    });

    this.buf = Buffer.alloc(0);
    this.nextId = 1;
    this.pending = new Map();
    this.proc.stdout.on("data", (chunk) => this.onData(chunk));
  }

  onData(chunk) {
    this.buf = Buffer.concat([this.buf, chunk]);
    while (true) {
      const sep = this.buf.indexOf("\r\n\r\n");
      if (sep < 0) return;
      const header = this.buf.slice(0, sep).toString("utf8");
      const m = header.match(/content-length:\s*(\d+)/i);
      if (!m) throw new Error("Missing Content-Length in MCP response");
      const len = Number(m[1]);
      const bodyStart = sep + 4;
      if (this.buf.length < bodyStart + len) return;
      const body = this.buf.slice(bodyStart, bodyStart + len).toString("utf8");
      this.buf = this.buf.slice(bodyStart + len);
      const msg = JSON.parse(body);
      if (msg.id !== undefined && this.pending.has(msg.id)) {
        this.pending.get(msg.id).resolve(msg);
        this.pending.delete(msg.id);
      }
    }
  }

  request(method, params = {}) {
    const id = this.nextId++;
    const payload = { jsonrpc: "2.0", id, method, params };
    const body = Buffer.from(JSON.stringify(payload), "utf8");
    const header = Buffer.from(`Content-Length: ${body.length}\r\n\r\n`, "utf8");
    this.proc.stdin.write(Buffer.concat([header, body]));
    return new Promise((resolve, reject) => {
      this.pending.set(id, { resolve, reject });
      setTimeout(() => {
        if (this.pending.has(id)) {
          this.pending.delete(id);
          reject(new Error(`MCP timeout for ${method}`));
        }
      }, 30000);
    });
  }

  async initialize() {
    const resp = await this.request("initialize", {});
    if (resp.error) throw new Error(resp.error.message || "initialize failed");
    return resp.result;
  }

  async callTool(name, args) {
    const resp = await this.request("tools/call", { name, arguments: args });
    if (resp.error) throw new Error(resp.error.message || "tool call failed");
    const content = resp.result?.content || [];
    const text = content.find((c) => c.type === "text")?.text;
    return text || "";
  }

  close() {
    if (!this.proc.killed) this.proc.kill("SIGTERM");
  }
}

function parseHazards(answer) {
  const out = [];
  const lines = answer.split("\n");
  const re = /^\d+\.\s(.+?)\s—\s.+?\s—\sapprox bbox\s(-?\d+(?:\.\d+)?),(-?\d+(?:\.\d+)?) to (-?\d+(?:\.\d+)?),(-?\d+(?:\.\d+)?)/;
  for (const line of lines) {
    const m = line.trim().match(re);
    if (!m) continue;
    out.push({
      title: m[1],
      bbox: {
        minLon: Number(m[2]),
        minLat: Number(m[3]),
        maxLon: Number(m[4]),
        maxLat: Number(m[5])
      }
    });
  }
  return out;
}

async function maybeConvertToMP4(inWebmPath, outMp4Path) {
  const ff = spawn("ffmpeg", [
    "-y",
    "-i",
    inWebmPath,
    "-c:v",
    "libx264",
    "-pix_fmt",
    "yuv420p",
    "-movflags",
    "+faststart",
    outMp4Path
  ]);
  return new Promise((resolve) => {
    ff.on("error", () => resolve(false));
    ff.on("exit", (code) => resolve(code === 0));
  });
}

async function run() {
  const mcp = new MCPClient();
  await mcp.initialize();

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
    recordVideo: { dir: tmpVideoDir, size: { width: 1280, height: 720 } }
  });
  const page = await context.newPage();
  await page.goto(`file://${join(repoRoot, "demo", "visualizer.html")}`);
  await page.evaluate(() => window.DemoAPI.reset());
  await delay(1500);

  for (const step of scenario) {
    const answer = await mcp.callTool("atlas_answer_query", { query: step.query, limit: 5 });
    const hazards = parseHazards(answer);
    await page.evaluate(
      ({ query, status, hazards, color }) => {
        window.DemoAPI.setQuery(query);
        window.DemoAPI.setStatus(status);
        window.DemoAPI.addHazards(hazards, color);
      },
      {
        query: step.query,
        status: `${step.label}: ${hazards.length} extents added`,
        hazards,
        color: step.color
      }
    );
    await delay(3000);
  }

  await page.evaluate(() => window.DemoAPI.setStatus("Scenario complete: fire, water, earthquakes, storms."));
  await delay(2000);

  const videoObj = page.video();
  await context.close();
  await browser.close();
  mcp.close();

  const generated = await videoObj.path();
  const webmOut = join(outputDir, "hazards-demo.webm");
  copyFileSync(generated, webmOut);
  rmSync(tmpVideoDir, { recursive: true, force: true });

  const mp4Out = join(outputDir, "hazards-demo.mp4");
  const converted = await maybeConvertToMP4(webmOut, mp4Out);
  if (converted) {
    console.log(`Demo video written: ${mp4Out}`);
  } else {
    console.log(`Demo video written: ${webmOut}`);
    console.log("ffmpeg not available, skipped mp4 conversion.");
  }
}

run().catch((err) => {
  console.error(err);
  process.exit(1);
});
