# Local voice cloning (your own voice)

Use voice cloning only for a voice you own or have explicit permission to model.

## Practical local stack

For high quality and local inference:

1. **XTTS v2 (Coqui)** for text-to-speech cloning
2. **RVC** (optional) for timbre conversion polish
3. **ffmpeg** for audio cleanup, loudness normalization, and final mix

## Minimal workflow

1. Record 3-10 minutes of clean dry speech (WAV, 24kHz+).
2. Trim silence/noise (`ffmpeg` or `audacity`).
3. Generate a speaker embedding/profile with XTTS tooling.
4. Synthesize narration from your script.
5. (Optional) Run through RVC model for style consistency.
6. Normalize and export:
   - `ffmpeg -i in.wav -af loudnorm out.wav`

## Good local model options

- **coqui-ai/TTS (XTTS v2)**: strong multilingual cloning, straightforward Python setup.
- **OpenVoice**: lightweight voice transfer/control.
- **Piper**: very fast local TTS (less expressive cloning than XTTS).

## Hardware guidance

- CPU-only works for short clips but is slow.
- NVIDIA GPU (8GB+ VRAM) is strongly recommended for fast iteration.

## Safety and labeling

- Keep source clips and consent records.
- Label generated voice as synthetic when sharing publicly.
- Avoid generating deceptive or impersonation content.

