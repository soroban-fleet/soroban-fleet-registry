# Soroban Fleet Registry — Documentation Site

Public documentation for [`soroban-fleet-registry`](https://github.com/soroban-fleet/soroban-fleet-registry) v1.0.0, built with [Docusaurus 3](https://docusaurus.io/).

This site is separate from the repository `README.md` and from the engineering notes in [`docs/`](../docs/). It serves two audiences:

- ecosystem reviewers and non-technical readers (Introduction, Concepts);
- developers, maintainers, and contributors (Architecture, Guides, API, Development, Operations, Contributing, Reference).

## Run locally

```bash
cd docs-site
npm ci
npm run start
```

The site is served at `http://localhost:3000` with hot reload.

## Build

```bash
cd docs-site
npm ci
npm run build
```

`npm run build` fails on broken internal links, broken anchors, and broken Markdown links (`onBrokenLinks`, `onBrokenAnchors`, and `markdown.hooks.onBrokenMarkdownLinks` are all set to `throw`).

Preview the production build:

```bash
npm run serve
```

## Search

Local full-text search is provided by [`@easyops-cn/docusaurus-search-local`](https://github.com/easyops-cn/docusaurus-search-local) and indexes every page under `docs/` at build time. No external search service is required.

## Structure

```text
docs-site/
├── docs/                    # Documentation pages (source of truth for site content)
├── src/                     # Landing page and styles
├── static/                  # Static assets served from the site root
├── docusaurus.config.ts     # Site configuration, navbar, footer, search
├── sidebars.ts              # Navigation order
├── package.json
├── tsconfig.json
└── README.md
```

Repository-level engineering documentation lives in [`../docs/`](../docs/) and is not replaced by this site.

## Editing content

1. Edit or add a Markdown file under `docs-site/docs/`.
2. Add new pages to `sidebars.ts`; unlisted pages are still built but are not reachable from navigation.
3. Run `npm run build` before opening a pull request.
4. Follow the writing rules in the repository: direct language, real commands, real schemas, no marketing claims, no invented numbers.

## Version

The content of this site describes **v1.0.0** of the project. The version is displayed in the navbar and links to the [v1.0.0 release](https://github.com/soroban-fleet/soroban-fleet-registry/releases/tag/v1.0.0). Do not document unreleased features as present.

## License

Apache License 2.0, same as the repository.
