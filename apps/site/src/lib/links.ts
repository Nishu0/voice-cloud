// Outbound links. Override at build time with VITE_APP_URL / VITE_DOCS_URL.
// The console is self-hosted, so the default points at a local installation.
export const APP_URL: string = import.meta.env.VITE_APP_URL ?? 'http://localhost:3000';
export const REPO_URL = 'https://github.com/Nishu0/voice-cloud';
export const DOCS_URL: string = import.meta.env.VITE_DOCS_URL ?? `${REPO_URL}/tree/main/docs`;
