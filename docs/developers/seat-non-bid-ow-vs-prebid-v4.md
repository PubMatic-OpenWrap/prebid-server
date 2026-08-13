# Seat Non-Bid: OpenWrap vs Prebid v4

Comparison and merge guidance for the IAB Seat Non-Bid extension when upgrading the PubMatic-OpenWrap fork to Prebid Server v4.

Related: [Prebid v4 upgrade notes](./prebid-upgrade-v4.7.0.md)

During merge review, `--ours` = upstream v4.x, `--theirs` = OW `master`.

---

## Background

Both implement the [IAB Seat Non-Bid extension](https://github.com/InteractiveAdvertisingBureau/openrtb/blob/master/extensions/community_extensions/seat-non-bid.md) — recording why bids were rejected and attaching them to `bidresponse.ext.prebid.seatnonbid`.

They are the **same feature**, but Prebid v4 and OpenWrap implement it in **different packages** with **different APIs and data models**.

| | Prebid v4.8.0 (upstream) | OpenWrap |
|---|---|---|
| Builder file | `exchange/seat_non_bids.go` | `openrtb_ext/seat_non_bids.go` |
| Builder type | `exchange.SeatNonBidBuilder` | `openrtb_ext.SeatNonBidBuilder` |
| Tests | `exchange/seat_non_bids_test.go` | `openrtb_ext/seat_non_bids_test.go` |
| Present in upstream v4.8.0 tag? | **Yes** | **No** (OW-only) |

OW removed `exchange/seat_non_bids.go` in **OTT-1168** (May 2024, commit `d8123665d`) and moved the builder into `openrtb_ext` so **hooks and exchange** share one type.

When merging upstream v4.8.0, Git reintroduces `exchange/seat_non_bids.go`. **Delete it** — it is upstream code, duplicates OW’s implementation, and **does not compile** against OW’s `openrtb_ext` types.

---

## Merge resolution

**Delete (from upstream v4 merge):**

- `exchange/seat_non_bids.go`
- `exchange/seat_non_bids_test.go`

**Keep (OW code — do not delete):**

- `openrtb_ext/seat_non_bids.go`
- `openrtb_ext/seat_non_bids_test.go`
- `exchange/non_bid_reason.go` — reason-code constants only; not the builder

---

## Why upstream `exchange/seat_non_bids.go` fails in the OW fork

Upstream’s exchange builder references types OW renamed or removed:

| Referenced by upstream `exchange/seat_non_bids.go` | In OW fork? |
|----------------------------------------------------|-------------|
| `openrtb_ext.NonBidExt` | No → OW uses `ExtNonBid` |
| `openrtb_ext.NonBidObject` | No → OW uses `ExtNonBidPrebidBid` |
| `openrtb_ext.ExtResponseNonBidPrebid` | No → OW uses `ExtNonBidPrebid` |
| `exchange.NonBidReason` | Yes, in `exchange/non_bid_reason.go` (enum only) |

Build errors if kept:

```
undefined: openrtb_ext.NonBidExt
undefined: openrtb_ext.ExtResponseNonBidPrebid
undefined: openrtb_ext.NonBidObject
```

---

## Builder API comparison

Both use the same underlying structure: `map[string][]NonBid` keyed by seat.

| Operation | Prebid v4 (`exchange`) | OpenWrap (`openrtb_ext`) |
|-----------|------------------------|---------------------------|
| Add rejected bid | `rejectBid(bid, reason, seat)` — **unexported**, builds `NonBid` inline | `AddBid(nonBid, seat)` — **exported**; caller builds `NonBid` via `NewNonBid()` |
| Add rejected imps | `rejectImps(impIds, exchange.NonBidReason, seat)` | `RejectImps(impIds, openrtb3.NoBidReason, seat)` |
| Merge builders | `append(...)` — **unexported** | `Append(...)` — **exported** |
| Convert to response slice | `Slice()` | `Get()` |
| Factory for `NonBid` | None (inline in `rejectBid`) | `NewNonBid(NonBidParams)` |

### Prebid v4 — minimal, exchange-internal

```go
// exchange/seat_non_bids.go (upstream)
func (b SeatNonBidBuilder) rejectBid(bid *entities.PbsOrtbBid, nonBidReason int, seat string) {
    nonBid := openrtb_ext.NonBid{
        ImpId: bid.Bid.ImpID, StatusCode: nonBidReason,
        Ext: &openrtb_ext.NonBidExt{
            Prebid: openrtb_ext.ExtResponseNonBidPrebid{
                Bid: openrtb_ext.NonBidObject{
                    Price: bid.Bid.Price, OriginalBidCPM: bid.OriginalBidCPM, ...
                },
            },
        },
    }
    b[seat] = append(b[seat], nonBid)
}
```

### OpenWrap — exported API + factory

```go
// openrtb_ext/seat_non_bids.go (OW)
func NewNonBid(bidParams NonBidParams) NonBid { ... }

func (snb *SeatNonBidBuilder) AddBid(nonBid NonBid, seat string) { ... }
func (snb *SeatNonBidBuilder) Append(nonBids ...SeatNonBidBuilder) { ... }
func (snb *SeatNonBidBuilder) Get() []SeatNonBid { ... }
func (snb *SeatNonBidBuilder) RejectImps(impIds []string, reason openrtb3.NoBidReason, seat string) { ... }
```

---

## NonBid data model comparison

JSON output shape is similar; Go types differ.

### Prebid v4 (`openrtb_ext/response.go` on upstream tag)

```go
type NonBidObject struct { /* basic bid fields + origbidcpm/cur */ }
type NonBidExt struct { Prebid ExtResponseNonBidPrebid }
type NonBid struct {
    ImpId, StatusCode int
    Ext *NonBidExt    // pointer, omitempty
}
```

### OpenWrap (`openrtb_ext/response.go`)

```go
type ExtNonBidPrebidBid struct {
    /* same core fields as upstream NonBidObject, plus OW fields: */
    ID, DealPriority, DealTierSatisfied, Meta, Targeting, Type,
    Video, BidId, Floors, OriginalBidCPMUSD, Bundle
}
type ExtNonBid struct {
    Prebid  ExtNonBidPrebid
    IsAdPod *bool   // OW-only internal flag
}
type NonBid struct {
    ImpId, StatusCode int
    Ext ExtNonBid     // value, not pointer
}
```

| Field in non-bid bid snapshot | Prebid v4 | OpenWrap |
|-------------------------------|-----------|----------|
| Core ORTB bid fields (price, w, h, dealid, …) | Yes | Yes |
| `origbidcpm` / `origbidcur` | Yes | Yes |
| Deal priority / tier | No | Yes |
| Prebid meta, targeting, type | No | Yes |
| Video, floors, bidid | No | Yes |
| Generated bid UUID | No | Yes |
| Ad-pod flag | No | Yes (`IsAdPod`, internal) |

OW’s model is a **superset** of upstream’s for analytics and hook integration.

---

## Who calls what in each codebase

### Prebid v4

- `exchange/exchange.go` owns `exchange.SeatNonBidBuilder`
- Calls `rejectBid()` / `rejectImps()` inside exchange
- Final attach via `setSeatNonBid()` → `seatNonBidBuilder.Slice()`
- Hooks do not own the builder type

### OpenWrap (this fork)

- `exchange/exchange.go`, `exchange/exchange_ow.go`, `exchange/bidder.go` use **`openrtb_ext.SeatNonBidBuilder`**
- Pattern: `seatNonBidBuilder.AddBid(openrtb_ext.NewNonBid(nonBidParams), seat)`
- OW hooks also emit seat non-bids (`beforevalidationhook`, `hook_raw_bidder_response`, etc.)
- `endpoints/openrtb2/auction_ow.go` merges hook + exchange non-bids via `seatNonBid.Get()`

---

## Architecture

```
Prebid v4.8.0                         OpenWrap fork
─────────────────                     ─────────────
exchange/seat_non_bids.go             openrtb_ext/seat_non_bids.go
  SeatNonBidBuilder                     SeatNonBidBuilder + NewNonBid()
       │                                      │
       ▼                                      ▼
openrtb_ext/response.go               openrtb_ext/response.go
  NonBid, NonBidExt,                  NonBid, ExtNonBid,
  NonBidObject                        ExtNonBidPrebidBid
```

---

## Merge checklist

1. After merging upstream v4.x, confirm `exchange/seat_non_bids.go` appeared from upstream → **delete both** `exchange/seat_non_bids.go` and `exchange/seat_non_bids_test.go`.
2. Confirm `openrtb_ext/seat_non_bids.go` and `openrtb_ext/seat_non_bids_test.go` are **present and unchanged** (OW code).
3. Run `go build ./...` — should pass with only `openrtb_ext` builder in use.
4. Grep for `exchange.SeatNonBidBuilder` or `.rejectBid(` — should find **no** references in OW fork (OW uses `openrtb_ext` API).

---

## FAQ

**Q: Are we losing Prebid 4 seat-non-bid functionality by deleting `exchange/seat_non_bids.go`?**

No. OW implements the same IAB extension in `openrtb_ext/` with a richer API and hook support. Deleting the upstream file removes a **duplicate, incompatible** copy.

**Q: Is `exchange/seat_non_bids.go` OW code?**

No. OW **deleted** it in 2024. What shows as deleted in git diff after a v4 merge is **upstream Prebid v4 code** being removed again.

**Q: Where is OW’s seat-non-bid code?**

`openrtb_ext/seat_non_bids.go` and `openrtb_ext/seat_non_bids_test.go`.
