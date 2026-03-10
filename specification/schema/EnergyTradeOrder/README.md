# EnergyTradeOrder

Order attributes for P2P energy trading — identifies BAP/BPP participants and total contracted energy quantity.

**Canonical IRI:** `https://schema.beckn.io/EnergyTradeOrder/v2.0`

**Namespace prefix:** `deg:` → `https://schema.beckn.io/deg/EnergyTradeOrder/v2.0/`

**Tags:** `energy-trade` · `p2p-trading` · `order`

---

## Versions

| Version | Status | Notes |
|---------|--------|-------|
| [v2.0](./v2.0/) | Current | Initial JSON Schema release, split from combined EnergyTrade schema |
| [v0.3](./v0.3/) | Deprecated | Original definition as a component in `EnergyTrade/v0.3/attributes.yaml` |

---

## Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `bap_id` | `string` | ✅ | Beckn Application Platform subscriber ID (buyer side) |
| `bpp_id` | `string` | ✅ | Beckn Provider Platform subscriber ID (seller side) |
| `total_quantity` | [Quantity](https://schema.beckn.io/Quantity/v2.0) | | Total energy quantity for the order (kWh) |

---

## Linked Data

| Term | IRI |
|------|-----|
| `EnergyTradeOrder` | `deg:EnergyTradeOrder` |
| `bap_id` | `deg:bap_id` |
| `bpp_id` | `deg:bpp_id` |
| `total_quantity` | `deg:total_quantity` |

---

## Usage

`EnergyTradeOrder` is attached to `Order.orderAttributes` in P2P energy trading beckn flows.
For inter-utility (inter-DISCOM) trades, buyer/seller utility IDs are captured separately in
`EnergyCustomer.utilityId` on the buyer and provider sides.
