# site

The Voice Cloud landing page: a static SvelteKit site (adapter-static).

```sh
bun install
bun run dev      # http://localhost:5174
bun run build    # static output in build/
bun run check
```

The "Open app" and "Docs" links come from `src/lib/links.ts`. Override them at
build time with `VITE_APP_URL` and `VITE_DOCS_URL`; by default the app link
points at a local console (`http://localhost:3000`) and the docs link at the
repository's `docs/` folder.
