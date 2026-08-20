# Prebid Server v4 Upgrade — Conflict Resolution Notes

Notes for merging OW `master` onto upstream Prebid v4 (branches `prebid_v4.7.0-review` / `prebid_v4.8.0-review`).

During merge review, `--ours` = upstream v4.x, `--theirs` = OW `master`.

---

## Seat Non-Bid

See **[seat-non-bid-ow-vs-prebid-v4.md](./seat-non-bid-ow-vs-prebid-v4.md)** for the full comparison of OW `openrtb_ext/seat_non_bids*` vs upstream `exchange/seat_non_bids*`, merge resolution (delete upstream exchange files, keep OW openrtb_ext files), API/data-model differences, and checklist.

---

## `static/bidder-params/nativo.json`

### Background

Nativo `placementId` support was added independently in two places:

| Source | PR / ticket | Merged | Schema key | Type | Required |
|--------|-------------|--------|------------|------|----------|
| **OW fork** | [PubMatic-OpenWrap/prebid-server#1108](https://github.com/PubMatic-OpenWrap/prebid-server/pull/1108) (Isha, UOE-12550) | Jun 2025 | Originally `placementid`; OW `master` later uses `placementID` | `string` | yes |
| **Upstream OSS** | [prebid/prebid-server#4679](https://github.com/prebid/prebid-server/pull/4679) (v4.2.0) | Upstream v4.2.0+ | `placementId` | `integer` \| `string` | no (optional) |

PR #1108 also changed OW-specific files that are **not** duplicated upstream:

- `adapters/nativo/nativo.go` — copied `imp.ext.bidder.placementid` → `imp.ext.nativo.placementid` before sending the request
- `openrtb_ext/imp_nativo.go` — OW struct used `placementid` / `placementID` json tags
- `modules/pubmatic/openwrap/adapters/bidders.go` — `builderNativo` maps OW profile `placementId` → bidder params

Upstream v4.2.0+ keeps `placementId` under `imp.ext.bidder` and does **not** copy it into `imp.ext.nativo`.

### Resolution: keep upstream v4.7.0

**Take `--ours` (v4.7.0) for:**

- `static/bidder-params/nativo.json`
- `adapters/nativo/nativo.go`
- `openrtb_ext/imp_nativo.go`
- `adapters/nativo/nativotest/exemplary/*.json` (upstream test fixtures)

Upstream schema:

```json
"placementId": {
    "type": ["integer", "string"],
    "description": "Placement ID"
}
```

No `required` array — param is optional per upstream.

**Keep OW-only pieces from `--theirs` (master):**

- `modules/pubmatic/openwrap/adapters/bidders.go` — `builderNativo`
- `modules/pubmatic/openwrap/adapters/builder.go` — Nativo builder registration
- `modules/pubmatic/openwrap/adapters/bidders_test.go` — `TestBuilderNativo`

### Follow-up after resolving the conflict

Update `builderNativo` so its JSON output uses the upstream key **`placementId`** (camelCase), not `placementID` or `placementid`:

```go
// Before (OW master)
fmt.Fprintf(jsonStr, `{"placementID":"%d"}`, pid)

// After (align with upstream v4.7.0 schema)
fmt.Fprintf(jsonStr, `{"placementId":%d}`, pid)
```

Also update `TestBuilderNativo` expected JSON accordingly.

OW profile config can continue to use `placementId` in `FieldMap`; only the emitted bidder-param key must match upstream.

### Why not keep OW master’s `nativo.json`?

1. Upstream v4.2.0 is the canonical Prebid schema — same feature, standardized naming (`placementId`).
2. OW’s adapter-side copy logic in PR #1108 is superseded by upstream’s simpler approach (params stay in `imp.ext.bidder`).
3. Keeping OW’s `placementID` / required-string schema would diverge from upstream validation and Prebid.js param conventions.

---

## ModuleContext API (v4 hook migration)

Upstream v4 changed hook `ModuleContext` from `map[string]any` to `*hookstage.ModuleContext` with `Get` / `Set` / `GetAll`.

OW hooks in `modules/pubmatic/openwrap/` were migrated to use helpers in `module_context.go`:

```go
// Old (OW v3)
moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
len(moduleCtx.ModuleContext) == 0
moduleCtx.ModuleContext["rctx"] = rCtx

// New (v4)
getRequestCtx(moduleCtx.ModuleContext)
!hasModuleContext(moduleCtx.ModuleContext)
setRequestCtx(moduleCtx.ModuleContext, rCtx)
```

Reference: `hooks/hookstage/invocation.go` (`NewModuleContext`, `Get`, `Set`).

---

## Other conflicts

_Add resolution notes for remaining merge conflicts here as they are reviewed._
