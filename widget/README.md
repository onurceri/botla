# Botla chat widget

The embeddable chat widget: a Preact + TypeScript bundle that mounts into a Shadow DOM, so the host page's CSS can't affect it. It loads the bot's public configuration, then sends messages to the public chat API. Domain allowlists and optional signed embed tokens are enforced by the API.

```html
<script src="https://<your-widget-host>/widget.js" data-bot="<chatbot-id>"></script>
```

## Develop

Install dependencies once from the repository root (`npm ci`), then:

```bash
cp .env.example .env     # VITE_API_BASE_URL, VITE_MARKETING_URL
npm run dev              # Vite dev server; public/widget-test.html embeds the widget
```

## Scripts

| Script | What it does |
|---|---|
| `npm run ci` | `tsc --noEmit`, ESLint, Vitest with coverage |
| `npm run build` | Builds `dist/widget.js` |
| `npm run test:e2e` | Playwright against `vite preview`; run `npm run build` first |
| `npm run build:analyze` | Build and open a bundle-size report |

Deploy to Cloudflare Pages with `make widget-deploy` from the repository root.
