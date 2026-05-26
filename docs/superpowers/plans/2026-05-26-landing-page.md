# Landing Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Place the production-ready Daco landing page at `landing/index.html` in the monorepo.

**Architecture:** Single self-contained static HTML/CSS/JS file extracted verbatim from the Claude Design export bundle. No build step, no framework, no external JS dependencies (only Google Fonts via CDN). Deploy by pointing Netlify/Vercel/GitHub Pages at the `landing/` directory.

**Tech Stack:** HTML5, CSS custom properties, vanilla JS, Inter + JetBrains Mono (Google Fonts CDN)

---

### Task 1: Create landing/index.html from the design bundle

**Files:**
- Create: `landing/index.html`

The design bundle is a gzip-compressed tar saved at:
```
/Users/gudmundurpalsson/.claude/projects/-Users-gudmundurpalsson-Documents-DacoLabs-cli/950014ff-e453-4cc5-bbd2-ec5d728e5362/tool-results/webfetch-1779779568350-jwtfo5.bin
```

The target file inside the tar is `daco-landing-page/project/Daco Landing.html`.

- [ ] **Step 1: Create the landing/ directory**

```bash
mkdir -p landing
```

- [ ] **Step 2: Extract the HTML file from the design bundle**

```bash
cat /Users/gudmundurpalsson/.claude/projects/-Users-gudmundurpalsson-Documents-DacoLabs-cli/950014ff-e453-4cc5-bbd2-ec5d728e5362/tool-results/webfetch-1779779568350-jwtfo5.bin \
  | gunzip \
  | tar -x --to-stdout "daco-landing-page/project/Daco Landing.html" \
  > landing/index.html
```

- [ ] **Step 3: Verify the file was created correctly**

```bash
wc -l landing/index.html
head -5 landing/index.html
```

Expected: ~1539 lines, first lines show `<!doctype html>` and `<title>Daco · Build data pipelines from your code</title>`

- [ ] **Step 4: Commit**

```bash
git add landing/index.html
git commit -m "feat: add Daco landing page

Self-contained HTML/CSS/JS landing page exported from Claude Design.
Deploy by pointing static host at the landing/ directory."
```
