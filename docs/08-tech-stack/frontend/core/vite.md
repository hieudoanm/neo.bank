# NEO Bank — Vite Best Practices

_Frontend · Core · part of the [NEO Bank tech stack](../../../08-tech-stack.md) decision record · see the [frontend index](../README.md) and the [core index](./README.md)._

Vite builds and serves the React + TypeScript SPA. It owns the dev server, the production bundle, environment variables, path aliases, and the shared transform pipeline that Vitest reuses.

## 1. Configuration

One `vite.config.ts` at the frontend root, exported with `defineConfig`.

```ts
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  server: {
    port: 3000,
    open: true,
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    chunkSizeWarningLimit: 600,
    rollupOptions: {
      output: {
        manualChunks: {
          react: ['react', 'react-dom'],
          router: ['react-router-dom'],
          query: ['@tanstack/react-query'],
        },
      },
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
  },
});
```

Rules:

- One config file; keep it declarative. Build tweaks live here, not in scripts.
- `@` aliases to `src`; keep the same alias in `tsconfig.json` (`paths`).
- Proxy `/api` to the Go API in development so the browser stays same-origin and CORS is a non-issue locally.
- Vendor chunks are split manually (`react`, `router`, `query`) to improve caching.
- `test` block shares config with Vitest — aliases and plugins work identically in tests.

## 2. Environment Variables

Only variables prefixed `VITE_` are exposed to the client; everything else stays server-side.

```env
// .env.example
VITE_API_URL=http://localhost:8080/api/v1
VITE_APP_ENV=development
```

```ts
// src/lib/env.ts
export const env = {
  apiUrl: import.meta.env.VITE_API_URL,
  appEnv: import.meta.env.VITE_APP_ENV,
} as const;

export function requireEnv() {
  if (!env.apiUrl) {
    throw new Error('VITE_API_URL is required');
  }
}
```

Rules:

- Never put secrets in a `VITE_` variable — they are bundled into the client and visible to everyone.
- Read env through a single `env` module (`import.meta.env`), never scattered across components.
- `.env.example` holds placeholders only; `.env*` real files are git-ignored.
- Fail fast at startup if a required `VITE_` variable is missing (call `requireEnv()` in `main.tsx`).

## 3. Dev Server

- `pnpm dev` serves with HMR on port 3000; the proxy forwards `/api` to the API.
- HMR preserves component state where possible; avoid module-scope side effects that break fast refresh.
- React Fast Refresh requires every file to export only components (plus constants); do not mix component and non-component exports in one file.
- The dev server restarts on config changes; keep `vite.config.ts` minimal.

## 4. Build

```bash
pnpm build      # tsc --noEmit && vite build
pnpm preview    # serve the production build locally
```

Rules:

- Production build runs `tsc --noEmit` first; a type error fails the build.
- Emit `dist/` with hashed asset filenames; the server serves them with long-lived cache headers.
- Sourcemaps are generated for error tracking; keep them out of the public bundle in production (or upload to an error tracking service).
- Never edit `dist/`; it is generated and git-ignored.
- `chunkSizeWarningLimit` is 600KB — a warning means a route is too heavy, not that the limit should rise.

## 5. Code Splitting & Chunking

Route-level lazy loading is the primary split (see [`react-router.md`](./react-router.md)).

```tsx
const TransactionsPage = lazy(() => import('@/pages/TransactionsPage'));
const SettingsPage = lazy(() => import('@/pages/SettingsPage'));
```

Rules:

- Split at route boundaries with `React.lazy`; keep the auth entry eager (login/register are on the critical path).
- Let Vite hash and split vendor code; do not hand-manage shared chunk lists unless profiling shows a need.
- Wrap lazy components in `<Suspense fallback={<Loading />}>`.
- Watch the `chunkSizeWarningLimit`; a warning means a route is too heavy.

## 6. Path Aliases & TypeScript

- `@/*` maps to `src/*` in both `vite.config.ts` and `tsconfig.json`.
- `tsc --noEmit` is the type gate; Vite only strips types and does not typecheck.
- Keep `tsconfig.json` strict (see [`typescript.md`](./typescript.md)).
- Import from concrete paths, not barrel files (`import { Button } from '@/components/ui/Button'` not `import { Button } from '@/components/ui'`).

## 7. Testing Integration

Vitest reuses the Vite pipeline and `resolve.alias`, so `@/` imports work in tests unchanged:

```ts
// vitest.config.ts (reuses vite config)
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'node:path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
  },
});
```

## 8. Performance

- Prefer native ESM and dynamic `import()` for large, rarely used modules (charts, editors, maps).
- Analyse the bundle in CI (`rollup-plugin-visualizer`) and track size over time.
- Avoid barrel files that pull the whole module graph; import from the concrete path.
- Keep the initial JS payload small; defer anything below the fold.
- Preload critical assets (`<link rel="modulepreload">` for the main chunk).

## 9. Code Style & Rules

- Arrow functions only in project code (`export const X = () => {}`); no `function` declarations.
- One `vite.config.ts`; declarative, no imperative build scripts.
- Only `VITE_` variables reach the client; secrets never do.
- `@` alias defined in both Vite and TypeScript configs.
- `tsc --noEmit` gates the build; never bypass it.
- Route-level code splitting; `dist/` is generated and never edited.
- Related: [`react-router.md`](./react-router.md) for lazy routes, [`vitest.md`](../testing/vitest.md) for the shared test pipeline, [`typescript.md`](./typescript.md) for the alias and strict config.
