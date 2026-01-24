# Documentation Website

The Drift documentation site is built with [Docusaurus](https://docusaurus.io/) and deployed to GitHub Pages.

## Structure

```
drift/
├── website/                    # Docusaurus site
│   ├── docusaurus.config.ts    # Site configuration
│   ├── sidebars.ts             # Sidebar navigation
│   ├── package.json            # Dependencies
│   ├── docs/                   # Generated docs (gitignored)
│   ├── versioned_docs/         # Versioned snapshots (gitignored)
│   └── src/                    # Custom pages and CSS
│
├── website-docs/               # Hand-written documentation
│   ├── intro.md                # Introduction page
│   └── guides/                 # Tutorial guides
│       ├── getting-started.md
│       ├── widgets.md
│       └── ...
│
└── cmd/docgen/main.go          # Documentation generator
```

## How It Works

1. **Hand-written guides** live in `website-docs/` at the repo root
2. **API docs** are generated from godoc comments using `gomarkdoc`
3. **cmd/docgen** copies guides and generates API docs into `website/docs/`
4. **Docusaurus** builds the static site from `website/docs/`

## Local Development

### Generate Documentation

```bash
# From repo root
go run ./cmd/docgen
```

This:
- Copies `website-docs/` → `website/docs/`
- Generates API docs for each package in `pkg/`
- Outputs to `website/docs/api/`

### Start Development Server

```bash
cd website
npm install   # First time only
npm start
```

Opens at http://localhost:3000/

### Build for Production

```bash
cd website
npm run build
```

Output is in `website/build/`.

## Versioning

Docusaurus supports versioned documentation. The current (unreleased) docs are "Next", and tagged releases become versioned snapshots.

**Important:** Versioned docs must be committed to the repository. The CI workflow only builds what's in the repo - it does not auto-generate versions.

### Creating a Version

When preparing a release, create the version snapshot locally and commit it:

```bash
# 1. Generate the latest docs
go run ./cmd/docgen

# 2. Create the version snapshot
cd website
npm run docusaurus docs:version v0.5.0

# 3. Commit the versioned docs
cd ..
git add website/versioned_docs website/versioned_sidebars website/versions.json
git commit -m "docs: add v0.5.0 documentation"
```

This creates:
- `versioned_docs/version-v0.5.0/` - Snapshot of docs
- `versioned_sidebars/version-v0.5.0-sidebars.json` - Sidebar config
- Updates `versions.json` with the new version

### Release Workflow

1. Finish all code changes for the release
2. Generate and version the docs (see above)
3. Commit the versioned docs
4. Tag the release (`git tag v0.5.0`)
5. Push both the commit and tag (`git push && git push --tags`)

The docs workflow will deploy the site with all committed versions.

### Version Configuration

Edit `docusaurus.config.ts` to control version display:

```typescript
docs: {
  lastVersion: 'current',  // or 'v0.4.0' to show latest release by default
  versions: {
    current: {
      label: 'Next',
      path: '',
    },
    'v0.4.0': {
      label: 'v0.4.0',
      path: 'v0.4.0',
    },
  },
},
```

## CI/CD

The GitHub Actions workflow (`.github/workflows/docs.yml`) automatically:

1. **On push to master**: Regenerates API docs and deploys the site
2. **Manual trigger**: Can be run manually via workflow_dispatch

The workflow builds and deploys whatever is in the repo, including any committed versioned docs in `website/versioned_docs/`. It does not auto-generate versions - see [Versioning](#versioning) for how to create them locally.

## Adding New Guides

1. Create a new `.md` file in `website-docs/guides/`
2. Add frontmatter:
   ```yaml
   ---
   id: my-guide
   title: My Guide
   sidebar_position: 10
   ---
   ```
3. Add to sidebar in `website/sidebars.ts`
4. Run `go run ./cmd/docgen` to copy to website

## Adding New Packages to API Docs

Edit `cmd/docgen/main.go` and add to the `packages` slice:

```go
var packages = []Package{
    // ... existing packages
    {Name: "newpkg", Path: "pkg/newpkg", Position: 15},
}
```

## Troubleshooting

### Build Errors with MDX

If you see MDX compilation errors about `<details>` or curly braces, the `processMarkdown` function in `cmd/docgen/main.go` handles these conversions. Check that the HTML tag filtering is working.

### Broken Anchor Warnings

Warnings about broken anchors (e.g., `#SomeType`) are from gomarkdoc's internal links not matching Docusaurus's anchor format. These are cosmetic and don't affect functionality.

### Missing Packages

Some packages have build constraints (e.g., `//go:build darwin`). The docgen uses `--tags darwin` to include these. If a package still fails, it may have complex constraints and will be skipped with a warning.
