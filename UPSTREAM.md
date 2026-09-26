# Upstream Cheesy Arena Checkpoint

Source repo: `../cheesy-arena`

Last reviewed commit: `2c54d16a49f1c9d1169d2be1d1438eaf1757393c`

Updated: 2026-09-26

Notes:
- This checkpoint records the full Cheesy Arena commit last reviewed for portable, game-agnostic changes.
- Future porting passes should compare this commit against the sibling repo's current `HEAD`, port only changes that fit cheesy-arena-lite, and then update this file to the newest reviewed upstream commit.

## 2026-09-26 porting pass

Reviewed all 13 commits after `a3b623d395ed0e1284a171115c4bc24e11c5ffa5` through the checkpoint above, using only the sibling checkout.

Ported:
- `e86332e`: Updated the eight-alliance double-elimination bracket's field breaks to five and ten minutes. This scheduling change is independent of scoring and field hardware.
- `d98a55e`: Read the driver-station mode byte at the correct TCP log packet offset.
- `9013a90`: Improved FMS field monitor station directions, radio icons, faded states, and alliance-side shortcut.
- `75a540e`: Track playoff cards and carried yellow cards by alliance, adapting the referee panel, match review, and displays to Lite's generic scores. Excluded the TBA publishing changes.
- `2090b5b`: Preserve display configuration drafts and acknowledge saves using configuration revisions.
- `d475c12`: Add audience-overlay zoom configuration.
- `5a465ed`: Add YouTube stream displays and stream caption options.
- `2c54d16`: Update Go to 1.26.7, x/crypto to v0.56.0, and associated dependencies.

Excluded:
- `375bb9f`, `f726dd7`: TBA match and award publishing changes.
- `820b7d1`: Game-specific traversal bonus display changes.
- `069a11e`: Game-specific hub motor clearing time.
- `f2f38f4`: Game-specific LED/hub testing labels.

Compatibility: As upstream does, playoff cards now use explicit alliance fields; historical per-team playoff card maps are not migrated into those fields. Qualification cards remain per-team.

Validation: `go fmt ./...`, `go generate ./...`, `go test ./...`, `go build`, and `node --test static/js/stream_displays_test.js` passed. Browser checks covered display save acknowledgements, playoff card controls and score previews, and the updated display pages.
