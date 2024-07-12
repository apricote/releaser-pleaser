# Functionality

## Reconciling PRs

- Checkout Git Repository
- Get Commits since last tag
- If there are releasable commits (Conventional Commit & SemVer) continue
- Build Changelog from Commits
- Figure out next version based on ConvCommits & SemVer
- Create new branch
- Update local version references & Changelog with new proposed changes
- Commit these
- Push branch to Forge
- Open PR in Forge

## Reconciling Tags

- Check if any merged PRs exist that have no associated tags
- Create tag+release with the planned version & changelog
