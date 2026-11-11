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

## What's in it

- `src/routes/+page.svelte`: the page: hero, feature cards, closing links, footer.
- `src/lib/components/Dither.svelte`: the dithered sky drawn behind the whole page.
- `src/lib/components/Wordmark.svelte`: the dot-matrix "Voice Cloud" in the hero.
- `src/lib/components/Navbar.svelte`: the floating navbar; it appears after the
  hero and names the section being read.
- `src/lib/components/Mark.svelte`, `static/logo.svg`, `static/favicon.svg`: the logo.

The data inside the console mock-ups on the cards (policy, model, and activity
names) is illustrative, not read from an installation.
