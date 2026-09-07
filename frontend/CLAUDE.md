# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
npm install
npm run dev      # Vite dev server on port 1001 (strictPort: fails if occupied)
npm run build
npm run preview
npm run lint     # eslint . --fix --cache  (auto-fixes in place) — CURRENTLY BROKEN, see below
```

`npm run lint` crashes before linting anything: `eslint.config.js` ends with a stray `module.exports = {...}` block, and since `package.json` sets `"type": "module"` that throws `ReferenceError: module is not defined in ES module scope`. Deleting that trailing block (lines 26–32) fixes it. Until then, lint results cannot be trusted as a signal.

There is no test framework configured — no test runner dependency, no test files, no `test` script. Verify changes by running the app.

## Backend dependency

The Go backend lives in the sibling repo `../oj-backend` and must be running on `http://localhost:9090` for the dev server to work. Vite proxies both `/api` (API calls) and `/uploads` (user-uploaded images/avatars) to it. `VITE_API_BASE_URL` is `/api` in both `.env` and `.env.production`.

## Architecture

### API + response contract

All HTTP goes through `src/utils/request.js` (axios, `withCredentials: true`, 5s timeout). `src/api/*.js` are thin per-domain wrappers returning that promise; components import named functions from there rather than calling axios.

The response interceptor enforces the backend's envelope `{ code, msg, data }`:

- `code === 200` → resolves with the **envelope**, so callers read `res.data`.
- `code` in `[20001, 20002, 20003]` (not logged in / token expired / token invalid) → toasts, calls `userStore.localLogout()` + `promptLogin()`, then rejects.
- any other `code`, or a network/non-2xx error → toasts `msg` and rejects.

**The interceptor already shows an `ElMessage` for every failure.** Do not add another error toast in a `.catch` — that produces duplicate messages. Catch only to reset local UI state (e.g. `loading = false`).

### Auth

Session is cookie-based; the frontend never holds a token. `src/stores/user.js` (setup-style Pinia store) caches the user object in `localStorage` purely so the UI can render logged-in state before the network settles. `main.js` restores it via `loadFromStorage()` and then revalidates against the backend with `userInfo()`, clearing local state if the cookie is dead.

Two distinct logouts: `logout()` is the active one (POSTs `/logout` to clear the backend cookie/Redis, then clears local); `localLogout()` only clears local state and is for passive cleanup on auth failure.

`isAdmin` is `user.role === 1`.

### Routing

`src/router/index.js` holds the root shell (`Layout` → children) and delegates feature trees to `src/router/modules/{contest,problem,submission,post,admin}.js`. All components are lazy `() => import(...)`.

Route `meta` drives cross-cutting behavior in the global `beforeEach`/`afterEach`:

- `meta.requiresAuth` → redirect to `/login?redirect=<fullPath>` when logged out.
- `meta.requiresAdmin` → redirect to `/` for non-admins (set once on the parent `adminRoutes`, inherited by children).
- `meta.title` → `document.title`; NProgress bar is started/stopped here (and in `router.onError`, so it can't get stuck).
- `/login` and `/404` are in `whiteList`; visiting `/login` while logged in bounces to `/`.

Guards are the *only* place these are enforced — adding a protected page means setting `meta`, not writing in-component checks.

### Layout + tab pages use provide/inject

Detail pages are a parent layout route that fetches the entity once and `provide()`s it to tab children, which `inject()` rather than re-fetching:

- `Problem/ProblemLayout.vue` provides `"problem"` → `ProblemDetail.vue`, `tabs/*`
- `User/index.vue` provides `"userProfile"` and `"userRating"` → `User/tabs/*`
- Contest tabs nest under `Contest/ContestDetail.vue` the same way.

When adding a tab, pull shared entity data from the injection instead of adding a request.

### Auto-import

`unplugin-auto-import` + `unplugin-vue-components` with `ElementPlusResolver`, so Element Plus components and APIs need no import. `unplugin-vue-components` also scans `src/components` by default — **everything in `src/components/` is globally registered and usable in templates without an import statement** (e.g. `<UserName>`, `<FTag>`, `<SubmissionStatus>`). Composition API helpers (`ref`, `computed`, …) are still explicitly imported throughout; match that.

### Constants mirror backend enums

`src/constants/index.js` is the single source of truth for values that must match Go-side `consts`: `CONTEST_TYPE` (1 ACM / 2 OI / 3 IOI / 4 CF), `JUDGE_STATUS` (0–10) with parallel `JUDGE_STATUS_TEXT` / `JUDGE_STATUS_CLASS` maps, post `categoryList`, CF-style `RATING_TIERS`, and `BALLOON_COLORS`. Changing a numeric code here requires a matching backend change. New judge statuses need an entry in *all three* status maps plus a `--judge-*` CSS variable.

### Styling system

Layered, applied in `main.js` in this order: `common.css` → `assets.css` → `table.css` → Element Plus CSS → `element-theme.css` → iconfont.

- `common.css` defines the design tokens as CSS custom properties on `:root`: colors, `--judge-*` status colors, radii, shadows, spacing, and the full `--table-*` scale. Prefer these variables over hardcoded values — the recent commit history is a sustained effort to unify cards, tables, and status colors this way.
- `table.css` styles native tables via two shared class names, `.f-table` and `.data-table`, so plain tables match Element Plus tables.
- `element-theme.css` holds Element Plus overrides.
- Tailwind 4 is enabled through `@tailwindcss/vite` and used for utility classes in templates.
- `variables.scss` is a small legacy SCSS partial; the CSS custom properties supersede it.

### Markdown and code editing

Two separate Markdown paths, intentionally:

- **Problem statements** use `renderMarkdown` from `src/utils/markdown.js` (markdown-it + texmath/KaTeX + highlight.js, sanitized with DOMPurify, `html: false`) rendered through `v-html`.
- **Posts and contest descriptions** use `src/components/MdEditor.vue`, a wrapper over `md-editor-v3` (`MdEditor` for editing, `MdPreview` for display) with image upload wired to `api/upload.js`. It passes `mdHeadingId` from `src/utils/md.js` to both preview and catalog — md-editor-v3 defaults these to *different* id generators, which silently breaks table-of-contents anchor jumps. Keep passing it.

`src/components/Markdown.vue`, `MarkdownView.vue`, and `MarkdownPreview.vue` are unused legacy variants — don't extend them; use one of the two paths above.

Code submission uses **CodeMirror 6** (`vue-codemirror` + `@codemirror/lang-cpp`) in `Problem/ProblemDetail.vue` and `Contest/ProblemDetail.vue`. `monaco-editor` and `vite-plugin-monaco-editor` are in `package.json` but are not imported anywhere and not wired into `vite.config.js`.

## Conventions

- Plain JavaScript, not TypeScript. `@/*` → `src/*` (declared in both `jsconfig.json` and `vite.config.js`).
- Vue 3 Composition API with `<script setup>` throughout.
- UI copy, comments, and commit messages are Chinese; commits follow conventional-commit prefixes, e.g. `style(frontend): 优化首页排行与题目提交交互`.
- The trailing `module.exports` block in `eslint.config.js` was an attempt to disable `vue/multi-word-component-names`. It never took effect (the rule comes from `pluginVue.configs["flat/essential"]`) and now breaks the whole lint run. To actually disable the rule, add a `rules` entry to the exported flat-config array instead.
