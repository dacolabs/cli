---
name: landing-page
description: Daco landing page implementation in landing/index.html — static HTML/CSS/JS, deployed to Netlify/Vercel/GitHub Pages
metadata:
  type: project
---

# Daco Landing Page

## Overview

Add a production-ready landing page at `landing/index.html` in the monorepo. It is a self-contained static HTML/CSS/JS file — no build step, no framework, no external JS dependencies — deployable by pointing any static host at the `landing/` directory.

The design was created in Claude Design and is exported as a gzip bundle. The file to implement is `daco-landing-page/project/Daco Landing.html` from that bundle.

## Placement

```
landing/
  index.html    ← the landing page
```

Single file. Nothing else required unless a deployment config is added later (netlify.toml, vercel.json).

## Design System

| Token | Value |
|---|---|
| Yellow accent | `#f4cf4a` |
| Dark background | `#0c0c0c` |
| Light body bg | `#ffffff` / `#faf9f5` |
| Body font | Inter (Google Fonts) |
| Code font | JetBrains Mono (Google Fonts) |
| Max width | 1200px |
| Gutter | 32px |

## Sections (top to bottom)

1. **Nav** — sticky dark nav, yellow Daco logo, links (Features / OpenDPI / Ecosystem / Blog), single "Book a demo" CTA button
2. **Hero** — full-width dark section, headline `"Your pipelines should run on your laptop."` with `locally` in yellow, subtitle, install snippet with copy-to-clipboard, animated code morph panel (YAML → PySpark / dbt / Pydantic / Go / TypeScript tabs)
3. **Trust strip** — "Works with the stack you already run" — Snowflake, Databricks, BigQuery, dbt, Postgres, Kafka
4. **Problem** — "Data engineers deserve a real local development loop" — no warehouse connection required to test, no deploy to verify a schema change
5. **Value props** — 3-column grid: True Local Development · Agent-Direct AI Integration · Metadata-Driven Security
6. **OpenDPI showcase** — live YAML code block, verified bullet points explaining the spec, link to spec docs
7. **Quote** — founder quote from Orri & Giuseppe, Founders, Daco
8. **Ecosystem** — 3 cards: Daco CLI (blinking terminal glyph) · Daco Service & SDKs (dark stacked-rack glyph) · Daco Studio (dark card, yellow "Premium" tag, 4-quadrant marketplace glyph)
9. **Blog** — 3 cards: CLI v0.2.0 changelog · OpenDPI standard post · Welcome to Daco
10. **Final CTA** — dark section with radial yellow glow, "See Daco on your stack", single demo CTA

## Interactivity

- Code morph panel: auto-cycles through language tabs every ~3s; pauses when user clicks a tab or panel scrolls out of view; resumes on re-entry
- Install snippet: copy-to-clipboard button, short success flash
- Light hover states on all buttons and cards
- Nav scrolls to sections via anchor links

## Implementation approach

Take the `Daco Landing.html` file from the Claude Design bundle exactly as-is and write it to `landing/index.html`. The file is already production-quality (1539 lines, fully self-contained). No conversion to a framework is needed — the monorepo is a Go CLI tool with no existing frontend framework, and the static file integrates cleanly with any static host.

## Out of scope

- Deployment configuration files (netlify.toml, vercel.json) — add later when deployment target is confirmed
- Analytics or tracking scripts
- CMS integration for blog posts
- Conversion of the page to a JS framework
