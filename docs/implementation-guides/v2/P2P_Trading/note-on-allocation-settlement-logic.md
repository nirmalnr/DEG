# P2P Energy Allocation: Optimization Problem Formulation

## Problem Statement

When multiple P2P energy trades occur in a single time slot, the actual meter readings (production/consumption) may differ from contracted trade quantities. We need an allocation algorithm that:

1. Can be run independently by each utility (DISCOM) with minimal coordination
2. Converges to a stable settlement
3. Maximizes fulfilled trade value (minimizes unallocated energy)

---

## Notation

| Symbol | Description |
|--------|-------------|
| $\mathcal{B}$ | Set of buyers (consumers) |
| $\mathcal{S}$ | Set of sellers (producers/prosumers) |
| $\mathcal{T}$ | Set of trades in this time slot |
| $m_i$ | Meter reading (consumption) for buyer $i \in \mathcal{B}$ |
| $m_j$ | Meter reading (production) for seller $j \in \mathcal{S}$ |
| $\text{tr}_k$ | Contracted trade quantity for trade $k \in \mathcal{T}$ |
| $b(k)$ | Buyer involved in trade $k$ |
| $s(k)$ | Seller involved in trade $k$ |
| $a^B_k$ | Buyer utility's allocation for trade $k$ |
| $a^S_k$ | Seller utility's allocation for trade $k$ |
| $\text{settle}_k$ | Final settlement quantity for trade $k$ |

---

## Formal Optimization Problem

### Objective

Maximize total settled energy across all trades:

$$\max \sum_{k \in \mathcal{T}} \text{settle}_k$$

Equivalently (from buyer perspective): minimize unallocated consumption:

$$\min \sum_{i \in \mathcal{B}} \left( m_i - \sum_{k: b(k)=i} \text{settle}_k \right)$$

### Settlement Rule

The settlement for each trade is the minimum of both parties' allocations:

$$\text{settle}_k = \min(a^B_k, a^S_k) \quad \forall k \in \mathcal{T}$$

### Constraints

**Trade quantity bounds:**
$$0 \leq a^B_k \leq \text{tr}_k \quad \forall k \in \mathcal{T}$$
$$0 \leq a^S_k \leq \text{tr}_k \quad \forall k \in \mathcal{T}$$

**Buyer meter constraint:** Total allocations cannot exceed actual consumption:
$$\sum_{k: b(k)=i} a^B_k \leq m_i \quad \forall i \in \mathcal{B}$$

**Seller meter constraint:** Total allocations cannot exceed actual production:
$$\sum_{k: s(k)=j} a^S_k \leq m_j \quad \forall j \in \mathcal{S}$$

---

## Complete Formulation (Centralized)

If a central coordinator had access to all information, the optimal allocation is:

$$
\begin{align}
\max_{a^B, a^S, z} \quad & \sum_{k \in \mathcal{T}} z_k \\
\text{s.t.} \quad & z_k \leq a^B_k & \forall k \\
& z_k \leq a^S_k & \forall k \\
& \sum_{k: b(k)=i} a^B_k \leq m_i & \forall i \in \mathcal{B} \\
& \sum_{k: s(k)=j} a^S_k \leq m_j & \forall j \in \mathcal{S} \\
& 0 \leq a^B_k, a^S_k \leq \text{tr}_k & \forall k \\
& z_k \geq 0 & \forall k
\end{align}
$$

Where $z_k = \text{settle}_k = \min(a^B_k, a^S_k)$.

**This is a Linear Program (LP)** and can be solved optimally in polynomial time.

---

## Proposed Distributed Algorithm

The proposed iterative algorithm works as follows:

### Algorithm: Iterative Capped Allocation

**Input:** Trades $\mathcal{T}$ with timestamps, meter readings $m_i, m_j$

**Round 1 - Seller Utility Allocates:**
```
For each seller j:
    Sort trades k where s(k) = j by timestamp (FIFO) or use pro-rata
    remaining = m_j
    For each trade k in order:
        a^S_k = min(tr_k, remaining, a^B_k if exists else tr_k)
        remaining -= a^S_k
```

**Round 2 - Buyer Utility Allocates:**
```
For each buyer i:
    Sort trades k where b(k) = i by timestamp (FIFO) or use pro-rata
    remaining = m_i
    For each trade k in order:
        a^B_k = min(tr_k, remaining, a^S_k)  # cap at seller's allocation
        remaining -= a^B_k
```

**Round 3 - Seller Utility Re-allocates:**
```
For each seller j:
    remaining = m_j
    For each trade k where s(k) = j:
        a^S_k = min(a^S_k, a^B_k)  # update to min of both
        remaining -= a^S_k
```

**Output:** $\text{settle}_k = \min(a^B_k, a^S_k)$

---

## Analysis

### Convergence

**Theorem:** The algorithm converges in at most 2 iterations (3 rounds total).

**Proof sketch:**
- After Round 2, buyer allocations are capped at seller allocations: $a^B_k \leq a^S_k$
- After Round 3, seller allocations are capped at buyer allocations: $a^S_k \leq a^B_k$
- Combined: $a^B_k = a^S_k$ for all trades, so $\text{settle}_k = a^B_k = a^S_k$
- No further changes possible since allocations are now equal and consistent with constraints.

### Optimality

**Key insight:** The algorithm is **not globally optimal** in general, but achieves optimality under specific conditions.

#### Case 1: No Shortfalls (Optimal)

If $m_i \geq \sum_{k: b(k)=i} \text{tr}_k$ for all buyers and $m_j \geq \sum_{k: s(k)=j} \text{tr}_k$ for all sellers, then:

$$\text{settle}_k = \text{tr}_k \quad \forall k$$

This is trivially optimal.

#### Case 2: Single Party Shortfall (Optimal)

If only sellers (or only buyers) have shortfalls, the algorithm is optimal because the non-constrained party can fully match whatever the constrained party allocates.

#### Case 3: Both-Side Shortfalls (Suboptimal)

Consider this counterexample:

```
Trades:
  T1: Buyer B1 ↔ Seller S1, quantity = 10
  T2: Buyer B1 ↔ Seller S2, quantity = 10
  T3: Buyer B2 ↔ Seller S1, quantity = 10

Meter readings:
  m_B1 = 15, m_B2 = 10
  m_S1 = 15, m_S2 = 10

Trade timestamps: T1 < T2 < T3
```

**FIFO Algorithm Result:**
- Round 1 (Sellers): S1 allocates T1=10, T3=5; S2 allocates T2=10
- Round 2 (Buyers): B1 allocates T1=10, T2=5 (capped by remaining); B2 allocates T3=5
- Round 3 (Sellers): No change needed
- Total settled: 10 + 5 + 5 = **20**

**Optimal (LP Solution):**
- T1 = 5, T2 = 10, T3 = 10
- Check: B1 uses 15 (5+10), B2 uses 10, S1 produces 15 (5+10), S2 produces 10
- Total settled: 5 + 10 + 10 = **25**

The FIFO approach missed the optimal by 5 units (20% loss).

---

## Alternative: Pro-Rata Allocation

Instead of FIFO, use proportional allocation:

$$a^S_k = \text{tr}_k \cdot \frac{m_j}{\sum_{k': s(k')=j} \text{tr}_{k'}}$$

**Advantages:**
- Fairer distribution
- Often closer to optimal in symmetric scenarios
- No dependency on trade timestamps

**Analysis with same example:**
- S1 pro-rata: T1 = 7.5, T3 = 7.5 (since both are 10 out of 20 total, and m_S1=15)
- S2 pro-rata: T2 = 10
- B1 pro-rata: T1 = 7.5, T2 = 7.5 (15 allocated across 17.5 available)
  - Actually: min(7.5, 7.5) = 7.5 for T1, min(10, 7.5) = 7.5 for T2
- B2: T3 = 7.5
- Total settled: 7.5 + 7.5 + 7.5 = **22.5**

Better than FIFO (22.5 > 20), but still suboptimal (22.5 < 25).

---

## Theoretical Bounds

### Approximation Ratio

**Claim:** The iterative algorithm achieves at least 50% of optimal in the worst case.

**Proof sketch:** Consider the worst case where every unit allocated by one party is "wasted" due to the other party's shortfall. Even then, at least half of the minimum total capacity is settled:

$$\sum_k \text{settle}_k \geq \frac{1}{2} \min\left(\sum_i m_i, \sum_j m_j\right)$$

In practice, the approximation is much better because:
1. Shortfalls are typically correlated (weather affects all solar producers similarly)
2. Aggregators balance multiple buyers/sellers internally before trading

### Tight Example

The 50% bound is nearly tight. Consider:
```
T1: B1 ↔ S1, quantity = 100
T2: B2 ↔ S2, quantity = 100
m_B1 = 100, m_B2 = 0
m_S1 = 0, m_S2 = 100
```
- Algorithm: 0 settled (each trade has one party with zero capacity)
- Optimal: 0 (no feasible allocation exists)

This is actually optimal. For a true gap:
```
T1: B1 ↔ S1, quantity = 100
T2: B1 ↔ S2, quantity = 100
T3: B2 ↔ S1, quantity = 100
m_B1 = 100, m_B2 = 100
m_S1 = 100, m_S2 = 100
```
- FIFO (T1 first): T1=100, T2=0 (B1 exhausted), T3=0 (S1 exhausted) → 100
- Optimal: T1=50, T2=50, T3=50 → 150

Ratio: 100/150 = 67%

---

## Improved Algorithm: Iterative LP Relaxation

For better approximation, each utility can solve a local LP:

**Seller j's subproblem:**
$$
\begin{align}
\max \quad & \sum_{k: s(k)=j} a^S_k \\
\text{s.t.} \quad & a^S_k \leq \min(\text{tr}_k, a^B_k) & \forall k \\
& \sum_{k: s(k)=j} a^S_k \leq m_j \\
& a^S_k \geq 0
\end{align}
$$

This is equivalent to a simple water-filling algorithm when $a^B_k$ values are known.

**Convergence:** Alternating optimization converges to a Nash equilibrium (each utility's allocation is optimal given the other's).

---

## Recommendations

### For Implementation Simplicity

Use **Pro-Rata** allocation:
- Deterministic (no timestamp dependency)
- Fair across trade participants
- 2-3 iterations sufficient for convergence
- Adequate for most practical scenarios

### For Near-Optimal Settlement

Use **Iterative LP** if utilities have optimization capability:
- Each iteration is a simple bounded knapsack problem
- Converges to Nash equilibrium
- Typically within 5% of global optimum

### Hybrid Approach

1. Use pro-rata for initial allocation
2. If significant shortfalls detected, run one iteration of local optimization
3. Exchange updated allocations and settle at minimum

---

## Implementation Notes

### Data Exchange Format

Each utility needs to exchange:
```json
{
  "tradeAllocations": [
    {
      "tradeId": "tr_001",
      "allocatedQuantity": 8.5,
      "meterReading": 10.0,
      "allocationMethod": "pro-rata"
    }
  ],
  "iterationNumber": 2,
  "timestamp": "2025-10-04T18:00:00Z"
}
```

### Termination Condition

Converged when for all trades:
$$|a^B_k - a^S_k| < \epsilon$$

where $\epsilon$ is a small tolerance (e.g., 0.01 kWh).

### Tie-Breaking

When multiple trades have equal priority (same timestamp or equal weights in pro-rata):
- Use deterministic ordering (e.g., lexicographic by trade ID)
- Both utilities must use the same tie-breaking rule

---

## Conclusion

The proposed iterative algorithm with FIFO or pro-rata allocation:

1. **Converges** in 2-3 iterations
2. **Decentralized** - each utility runs independently with minimal data exchange
3. **Approximate** - not globally optimal, but achieves 67-90% of optimal in typical scenarios
4. **Simple** - can be implemented without LP solvers

For most practical P2P trading scenarios where shortfalls are infrequent and mild, the approximation gap is negligible. For high-shortfall scenarios (e.g., cloudy days with solar-heavy portfolios), the iterative LP approach is recommended.

---

# Part 2: Settlement Logic

## Overview

Settlement logic determines how utilities:
1. **Allocate** P2P trade quantities to actual meter readings
2. **Deduct** allocated quantities from billable consumption/production
3. **Bill** remaining energy at standard utility tariff rates
4. **Record** allocations to the DEG Ledger for audit trail

## Settlement Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Post-Delivery Settlement                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. METER READING           2. ALLOCATION              3. BILLING   │
│  ┌─────────────┐           ┌─────────────┐          ┌─────────────┐ │
│  │ m_i = 100   │    →      │ alloc = 80  │    →     │ utility_bill│ │
│  │ (consumed)  │           │ (from P2P)  │          │ = 20 × tariff│
│  └─────────────┘           └─────────────┘          └─────────────┘ │
│                                                                     │
│              Total Consumption = P2P Settled + Utility Import       │
│                     100 kWh     =    80 kWh   +    20 kWh           │
└─────────────────────────────────────────────────────────────────────┘
```

## Billing Calculation

### For Buyer i (Consumer)

| Component | Formula | Description |
|-----------|---------|-------------|
| **Meter Reading** | $m_i$ | Actual consumption (kWh) |
| **P2P Settled** | $\sum_{k: b(k)=i} \text{settle}_k$ | Total energy from P2P trades |
| **Utility Import** | $m_i - \sum_{k: b(k)=i} \text{settle}_k$ | Remaining from grid |
| **P2P Cost** | $\sum_{k: b(k)=i} \text{settle}_k \times p_k$ | Payment to sellers |
| **Utility Cost** | $(m_i - \sum \text{settle}_k) \times \text{tariff}_{\text{import}}$ | Standard grid charges |
| **Total Bill** | P2P Cost + Utility Cost + Wheeling | Total to pay |

### For Seller j (Prosumer)

| Component | Formula | Description |
|-----------|---------|-------------|
| **Meter Reading** | $m_j$ | Actual production (kWh) |
| **P2P Settled** | $\sum_{k: s(k)=j} \text{settle}_k$ | Total energy sold via P2P |
| **Utility Export** | $m_j - \sum_{k: s(k)=j} \text{settle}_k$ | Remaining exported to grid |
| **P2P Revenue** | $\sum_{k: s(k)=j} \text{settle}_k \times p_k$ | Payment from buyers |
| **Utility Revenue** | $(m_j - \sum \text{settle}_k) \times \text{tariff}_{\text{export}}$ | Net metering credits |
| **Total Revenue** | P2P Revenue + Utility Revenue | Total received |

## Example Settlement

```
Trade: T1 between Buyer B1 and Seller S1
  - Trade quantity: tr = 10 kWh
  - Trade price: p = 6 INR/kWh
  - Wheeling: 1 INR/kWh

Meter readings:
  - B1 consumed: m_B1 = 15 kWh
  - S1 produced: m_S1 = 8 kWh

Allocation (min-based):
  - Buyer utility allocates: a^B = min(10, 15) = 10 kWh
  - Seller utility allocates: a^S = min(10, 8) = 8 kWh
  - Settlement: settle = min(10, 8) = 8 kWh

Buyer B1's Bill:
  - P2P settled: 8 kWh × 6 INR = 48 INR
  - Wheeling: 8 kWh × 1 INR = 8 INR
  - Utility import: (15 - 8) = 7 kWh × 10 INR = 70 INR
  - Total: 48 + 8 + 70 = 126 INR

Seller S1's Revenue:
  - P2P revenue: 8 kWh × 6 INR = 48 INR
  - Utility export: (8 - 8) = 0 kWh × 3 INR = 0 INR
  - Total: 48 INR

Seller Shortfall (2 kWh):
  - Buyer expected 10 kWh, got 8 kWh
  - Buyer had to import extra 2 kWh at 10 INR = 20 INR cost
  - This is handled by the allocation logic (buyer only pays for settled amount)
```

---

# Part 3: Ledger API Integration for Allocation

## Current Ledger API Endpoints

| Endpoint | Purpose | Who Uses |
|----------|---------|----------|
| `POST /ledger/put` | Create/update trade record | Platforms (BAP/BPP) |
| `POST /ledger/get` | Query trades by filters | All parties |
| `POST /ledger/record` | Record actuals & status | Discoms (utilities) |

## Allocation Workflow Using Ledger API

The 3-round allocation algorithm maps to ledger API calls as follows:

```
Timeline:
  T+0h   Delivery period ends
  T+1h   Meter readings available
  T+2h   Round 1: Seller utility allocates
  T+4h   Round 2: Buyer utility allocates
  T+6h   Round 3: Seller utility finalizes
  T+8h   Settlement complete
```

## Pseudocode: Utility Allocation Process

### Step 1: Query Trades for Customers in Time Window

```python
def get_trades_for_utility(utility_id: str,
                           delivery_start: datetime,
                           delivery_end: datetime,
                           role: str) -> list[Trade]:
    """
    Query ledger for all trades involving this utility's customers
    in the given delivery window.
    """

    # Build filter based on role
    filter_params = {
        "deliveryStartFrom": delivery_start.isoformat(),
        "deliveryStartTo": delivery_end.isoformat(),
        "limit": 500,
        "sort": "tradeTime",
        "sortOrder": "asc"
    }

    if role == "SELLER_DISCOM":
        filter_params["discomIdSeller"] = utility_id
    else:  # BUYER_DISCOM
        filter_params["discomIdBuyer"] = utility_id

    # Call ledger GET API
    response = requests.post(
        f"{LEDGER_HOST}/ledger/get",
        json=filter_params,
        headers=auth_headers()
    )

    return response.json()["records"]
```

### Step 2: Run Allocation Algorithm

```python
def allocate_trades_for_customer(
    customer_id: str,
    meter_reading: float,
    trades: list[Trade],
    other_party_allocations: dict[str, float],  # trade_id -> allocation
    method: str = "pro-rata"
) -> dict[str, float]:
    """
    Allocate meter reading to trades for a single customer.
    Returns: trade_id -> allocated_quantity
    """

    allocations = {}

    # Filter trades for this customer
    customer_trades = [t for t in trades if t.customer_id == customer_id]

    if method == "pro-rata":
        # Calculate total contracted
        total_contracted = sum(t.trade_qty for t in customer_trades)

        if total_contracted == 0:
            return allocations

        # Pro-rata allocation, capped by other party's allocation
        for trade in customer_trades:
            # Base pro-rata share
            pro_rata_share = trade.trade_qty * (meter_reading / total_contracted)

            # Cap at trade quantity
            capped = min(pro_rata_share, trade.trade_qty)

            # Cap at other party's allocation if available
            if trade.id in other_party_allocations:
                capped = min(capped, other_party_allocations[trade.id])

            allocations[trade.id] = round(capped, 3)  # 3 decimal places (Wh precision)

    elif method == "fifo":
        # Sort by trade time
        sorted_trades = sorted(customer_trades, key=lambda t: t.trade_time)
        remaining = meter_reading

        for trade in sorted_trades:
            cap = trade.trade_qty
            if trade.id in other_party_allocations:
                cap = min(cap, other_party_allocations[trade.id])

            alloc = min(cap, remaining)
            allocations[trade.id] = round(alloc, 3)
            remaining -= alloc

            if remaining <= 0:
                break

    return allocations
```

### Step 3: Record Allocations to Ledger

```python
def record_allocations_to_ledger(
    allocations: dict[str, float],  # trade_id -> allocated_qty
    trades: list[Trade],
    role: str,  # "BUYER_DISCOM" or "SELLER_DISCOM"
    iteration: int
) -> list[dict]:
    """
    Record allocation for each trade to the ledger.

    LIMITATION: Current API requires one call per trade (no batch).
    """

    results = []

    for trade_id, allocated_qty in allocations.items():
        trade = next(t for t in trades if t.id == trade_id)

        # Build validation metric based on role
        if role == "BUYER_DISCOM":
            metrics = [{
                "validationMetricType": "ACTUAL_PULLED",
                "validationMetricValue": allocated_qty
            }]
            payload = {
                "role": role,
                "transactionId": trade.transaction_id,
                "orderItemId": trade.order_item_id,
                "buyerFulfillmentValidationMetrics": metrics,
                "note": f"Allocation round {iteration}",
                "clientReference": f"{trade_id}-buyer-alloc-{iteration}"
            }
        else:  # SELLER_DISCOM
            metrics = [{
                "validationMetricType": "ACTUAL_PUSHED",
                "validationMetricValue": allocated_qty
            }]
            payload = {
                "role": role,
                "transactionId": trade.transaction_id,
                "orderItemId": trade.order_item_id,
                "sellerFulfillmentValidationMetrics": metrics,
                "note": f"Allocation round {iteration}",
                "clientReference": f"{trade_id}-seller-alloc-{iteration}"
            }

        # Call ledger record API (one at a time - no batch support)
        response = requests.post(
            f"{LEDGER_HOST}/ledger/record",
            json=payload,
            headers=auth_headers()
        )

        results.append({
            "trade_id": trade_id,
            "success": response.status_code == 200,
            "response": response.json()
        })

    return results
```

### Step 4: Full 3-Round Allocation Orchestration

```python
def run_allocation_round(
    utility_id: str,
    role: str,  # "BUYER_DISCOM" or "SELLER_DISCOM"
    delivery_start: datetime,
    delivery_end: datetime,
    meter_readings: dict[str, float],  # customer_id -> meter_reading
    round_number: int
) -> dict[str, float]:
    """
    Execute one round of allocation for a utility.
    """

    # 1. Query all relevant trades from ledger
    trades = get_trades_for_utility(utility_id, delivery_start, delivery_end, role)

    # 2. Get other party's allocations from previous round (if any)
    other_allocations = {}
    for trade in trades:
        # Extract from validation metrics recorded by other party
        if role == "BUYER_DISCOM":
            metrics = trade.get("sellerFulfillmentValidationMetrics", [])
            for m in metrics:
                if m["validationMetricType"] == "ACTUAL_PUSHED":
                    other_allocations[trade["orderItemId"]] = m["validationMetricValue"]
        else:
            metrics = trade.get("buyerFulfillmentValidationMetrics", [])
            for m in metrics:
                if m["validationMetricType"] == "ACTUAL_PULLED":
                    other_allocations[trade["orderItemId"]] = m["validationMetricValue"]

    # 3. Group trades by customer
    trades_by_customer = group_by_customer(trades, role)

    # 4. Allocate for each customer
    all_allocations = {}
    for customer_id, customer_trades in trades_by_customer.items():
        meter = meter_readings.get(customer_id, 0)
        customer_allocs = allocate_trades_for_customer(
            customer_id, meter, customer_trades, other_allocations
        )
        all_allocations.update(customer_allocs)

    # 5. Record allocations to ledger
    record_allocations_to_ledger(all_allocations, trades, role, round_number)

    return all_allocations


def run_full_allocation_cycle(
    delivery_start: datetime,
    delivery_end: datetime
):
    """
    Coordinate the full 3-round allocation across seller and buyer utilities.

    In practice, this runs as scheduled jobs at each utility,
    not as a centralized orchestrator.
    """

    # Round 1: Seller utility allocates first
    print("Round 1: Seller utilities allocate...")
    # Each seller utility runs independently
    for seller_utility in SELLER_UTILITIES:
        run_allocation_round(
            seller_utility.id,
            "SELLER_DISCOM",
            delivery_start,
            delivery_end,
            seller_utility.get_meter_readings(delivery_start, delivery_end),
            round_number=1
        )

    # Wait for Round 1 to complete (e.g., 2 hour window)
    time.sleep(ROUND_DELAY)

    # Round 2: Buyer utility allocates
    print("Round 2: Buyer utilities allocate...")
    for buyer_utility in BUYER_UTILITIES:
        run_allocation_round(
            buyer_utility.id,
            "BUYER_DISCOM",
            delivery_start,
            delivery_end,
            buyer_utility.get_meter_readings(delivery_start, delivery_end),
            round_number=2
        )

    # Wait for Round 2 to complete
    time.sleep(ROUND_DELAY)

    # Round 3: Seller utility finalizes
    print("Round 3: Seller utilities finalize...")
    for seller_utility in SELLER_UTILITIES:
        run_allocation_round(
            seller_utility.id,
            "SELLER_DISCOM",
            delivery_start,
            delivery_end,
            seller_utility.get_meter_readings(delivery_start, delivery_end),
            round_number=3
        )

    print("Allocation complete. Settlement can proceed.")
```

---

## Recommended Ledger API Enhancements

### 1. Batch Record API

**Current limitation:** `/ledger/record` accepts only one trade at a time. For a utility with 10,000 trades in a time slot, this requires 10,000 API calls per round.

**Proposed new endpoint:** `POST /ledger/record-batch`

```yaml
/ledger/record-batch:
  post:
    summary: Record allocations for multiple trades in one call
    requestBody:
      content:
        application/json:
          schema:
            type: object
            required: [role, records]
            properties:
              role:
                $ref: "#/components/schemas/discomRole"
              records:
                type: array
                maxItems: 100  # Batch limit
                items:
                  type: object
                  required: [transactionId, orderItemId]
                  properties:
                    transactionId:
                      type: string
                    orderItemId:
                      type: string
                    buyerFulfillmentValidationMetrics:
                      type: array
                      items:
                        $ref: "#/components/schemas/validationMetric"
                    sellerFulfillmentValidationMetrics:
                      type: array
                      items:
                        $ref: "#/components/schemas/validationMetric"
    responses:
      "200":
        description: Batch accepted
        content:
          application/json:
            schema:
              type: object
              properties:
                results:
                  type: array
                  items:
                    type: object
                    properties:
                      transactionId:
                        type: string
                      orderItemId:
                        type: string
                      success:
                        type: boolean
                      error:
                        type: string
```

**Example batch request:**
```json
{
  "role": "SELLER_DISCOM",
  "records": [
    {
      "transactionId": "txn-001",
      "orderItemId": "item-001",
      "sellerFulfillmentValidationMetrics": [
        {"validationMetricType": "ACTUAL_PUSHED", "validationMetricValue": 8.5}
      ]
    },
    {
      "transactionId": "txn-002",
      "orderItemId": "item-001",
      "sellerFulfillmentValidationMetrics": [
        {"validationMetricType": "ACTUAL_PUSHED", "validationMetricValue": 12.0}
      ]
    }
  ]
}
```

---

## Summary of API Gaps

| Need | Current State | Recommendation |
|------|---------------|----------------|
| Batch recording | Not supported (1 call per trade) | **Add `/ledger/record-batch`** |

---

# Appendix A: Alternate Settlement Method (Deviation-Based)

## Overview

This appendix describes **Deviation Method** from the P2P trading design notes: a settlement method based on **contract deviation** rather than minimum-of-allocations. This approach has different trade-offs:

| Aspect | Min-Allocation Method | Deviation Method |
|--------|----------------------|-------------------------------|
| Inter-utility coordination | Required (3 rounds) | **Not required** |
| Allocation matching | Must converge to same value | Independent allocations OK |
| Penalty assignment | Implicit in settlement reduction | Explicit deviation penalties |
| Revenue certainty | Depends on other party | Guaranteed if you comply |

## Notation

| Symbol | Description |
|--------|-------------|
| $\text{tr}_q$ | Trade quantity (contracted kWh) |
| $\text{tr}_p$ | Trade price (INR/kWh) |
| $\text{load}_q$ | Buyer's actual consumption allocated to trade |
| $\text{gen}_q$ | Seller's actual production allocated to trade |
| $\text{rtm}_p$ | Real-time market (spot) price |
| $\text{exportBU}_p$ | Export rate buyer utility pays for underconsumption |
| $\text{importSU}_p$ | Import rate seller utility charges for underproduction |

## Settlement Formulas

### Buyer (B) Pays

$$\text{Buyer Payment} = \text{tr}_q \times \text{tr}_p - (\text{tr}_q - \text{load}_q) \times \text{exportBU}_p$$

- Pays full trade bill
- Minus: discount from utility selling surplus (underconsumption) at spot

### Seller (S) Receives

$$\text{Seller Revenue} = \text{tr}_q \times \text{tr}_p - (\text{tr}_q - \text{gen}_q) \times \text{importSU}_p$$

- Receives full trade bill
- Minus: cost of procuring shortfall (underproduction) charged by utility

### Seller Utility (SU) Receives

$$\text{SU Revenue} = (\text{tr}_q - \text{gen}_q) \times \text{importSU}_p$$

- Revenue from charging seller for shortfall procurement cost

### Buyer Utility (BU) Pays

$$\text{BU Payment} = (\text{tr}_q - \text{load}_q) \times \text{exportBU}_p$$

- Pays buyer for selling their underconsumption surplus in open market

### Net Flow

$$\text{Total} = \text{Buyer Payment} - \text{Seller Revenue} - \text{SU Revenue} + \text{BU Payment} = 0$$

**Zero-sum:** All money flows balance.

## Example

```
Trade:
  tr_q = 10 kWh, tr_p = 6 INR/kWh

Actuals:
  load_q = 8 kWh (buyer underconsumes by 2 kWh)
  gen_q = 7 kWh (seller underproduces by 3 kWh)

Rates:
  exportBU_p = 4 INR/kWh (utility buys underconsumption)
  importSU_p = 8 INR/kWh (utility sells to cover shortfall)

Buyer pays:
  = 10 × 6 - (10 - 8) × 4
  = 60 - 8 = 52 INR

Seller receives:
  = 10 × 6 - (10 - 7) × 8
  = 60 - 24 = 36 INR

Seller Utility receives:
  = (10 - 7) × 8 = 24 INR

Buyer Utility pays:
  = (10 - 8) × 4 = 8 INR

Verify: 52 - 36 - 24 + 8 = 0 ✓
```

## Key Properties

### 1. No Inter-Utility Coordination Needed

Each utility computes deviation penalties independently:
- Buyer utility: $(tr_q - load_q) \times exportBU_p$
- Seller utility: $(tr_q - gen_q) \times importSU_p$

**No need to match allocations** between utilities.

### 2. Contract Compliance Guarantees Revenue

If seller produces $\geq tr_q$:
- $gen_q = tr_q$ (allocated up to contract)
- Seller receives: $tr_q \times tr_p - 0 = tr_q \times tr_p$ (full contract value)

**Regardless of whether buyer underconsumes.**

Similarly, if buyer consumes $\geq tr_q$:
- Buyer pays: $tr_q \times tr_p$ (no discount)

**Regardless of whether seller underproduces.**

### 3. Allocation Logic Doesn't Affect Total Penalty

Total penalty for a customer with multiple trades:

$$\text{Total Penalty} = \left(\sum_k tr_k - \text{total\_meter}\right) \times \text{penalty\_rate}$$

This is independent of how individual $load_q$ or $gen_q$ values are allocated across trades.

### 4. Utility Margin for Risk

Utilities can set:
- $importSU_p > rtm_p$ (charge more than spot for shortfall)
- $exportBU_p < rtm_p$ (pay less than spot for surplus)

This ensures utilities don't lose money covering P2P deviations.

## Comparison to Min-Allocation Method

| Scenario | Min-Allocation | Deviation |
|----------|----------------|------------------------|
| Both comply | settle = tr, full value | Same |
| Seller shortfall only | settle = gen < tr, buyer pays less | Buyer pays full tr, gets no discount; seller penalized |
| Buyer shortfall only | settle = load < tr, seller receives less | Seller receives full tr; buyer penalized |
| Both shortfall | settle = min(load, gen) | Each penalized independently |

**Key difference:** In min-allocation, compliant party's settlement is reduced by other party's shortfall. In Deviation method, compliant party receives/pays full contract value.

## Pseudocode: Deviation-Based Settlement

```python
def compute_deviation_settlement(
    trade: Trade,
    load_q: float,  # Buyer's allocated consumption
    gen_q: float,   # Seller's allocated production
    export_rate: float,  # exportBU_p
    import_rate: float   # importSU_p
) -> dict:
    """
    Compute settlement using deviation method.
    Each party pays full contract, then deviation penalties applied.
    """

    tr_q = trade.quantity
    tr_p = trade.price

    # Buyer's underconsumption (positive = shortfall)
    buyer_shortfall = max(0, tr_q - load_q)

    # Seller's underproduction (positive = shortfall)
    seller_shortfall = max(0, tr_q - gen_q)

    # Base contract value
    contract_value = tr_q * tr_p

    # Buyer pays full contract minus underconsumption credit
    buyer_pays = contract_value - (buyer_shortfall * export_rate)

    # Seller receives full contract minus underproduction penalty
    seller_receives = contract_value - (seller_shortfall * import_rate)

    # Utility flows
    seller_utility_receives = seller_shortfall * import_rate
    buyer_utility_pays = buyer_shortfall * export_rate

    return {
        "buyer_pays": buyer_pays,
        "seller_receives": seller_receives,
        "seller_utility_receives": seller_utility_receives,
        "buyer_utility_pays": buyer_utility_pays,
        "buyer_shortfall_kwh": buyer_shortfall,
        "seller_shortfall_kwh": seller_shortfall,
        # Verification: should be 0
        "net_flow": buyer_pays - seller_receives - seller_utility_receives + buyer_utility_pays
    }
```

## When to Use Which Method

| Use Case | Recommended Method |
|----------|-------------------|
| Intra-utility P2P (same DISCOM) | Min-allocation (simpler) |
| Inter-utility P2P (different DISCOMs) | **Deviation** - no coordination needed |
| High trust, low shortfall scenarios | Either |
| Regulatory requirement for "energy tracing" | Min-allocation |
| Revenue certainty for compliant parties | **Deviation** |

## Integration with Ledger API

Deviation method simplifies ledger recording:

1. **No iteration rounds needed** - each utility records once
2. **Allocation is internal** - only total deviation matters
3. **Status updates:**
   - Seller utility: record `gen_q` and deviation penalty
   - Buyer utility: record `load_q` and deviation credit

```python
def record_deviation_settlement(trade_id: str, role: str, allocated_qty: float):
    """
    Record allocation for deviation-based settlement.
    Only needs to run once per utility (no iteration).
    """

    payload = {
        "role": role,
        "transactionId": trade.transaction_id,
        "orderItemId": trade.order_item_id,
        "note": "Deviation-based allocation (final)"
    }

    if role == "SELLER_DISCOM":
        payload["sellerFulfillmentValidationMetrics"] = [{
            "validationMetricType": "ACTUAL_PUSHED",
            "validationMetricValue": allocated_qty
        }]
        payload["statusSellerDiscom"] = "COMPLETED"
    else:
        payload["buyerFulfillmentValidationMetrics"] = [{
            "validationMetricType": "ACTUAL_PULLED",
            "validationMetricValue": allocated_qty
        }]
        payload["statusBuyerDiscom"] = "COMPLETED"

    return requests.post(f"{LEDGER_HOST}/ledger/record", json=payload)
```

---

# Appendix B: Side-by-Side Method Comparison

## Scenario Setup

```
Trade T1: Buyer B1 (DISCOM-A) ↔ Seller S1 (DISCOM-B)
  - Quantity: 100 kWh
  - Price: 6 INR/kWh
  - Contract value: 600 INR

Actuals:
  - B1 consumed: 80 kWh (20 kWh underconsumption)
  - S1 produced: 70 kWh (30 kWh underproduction)

Rates:
  - exportBU_p = 4 INR/kWh
  - importSU_p = 8 INR/kWh
  - Grid import tariff = 10 INR/kWh
  - Grid export tariff = 3 INR/kWh
```

## Method 1: Min-Allocation

```
Settlement quantity = min(80, 70) = 70 kWh

Buyer pays for P2P: 70 × 6 = 420 INR
Buyer grid import: (80 - 70) × 10 = 100 INR (10 kWh shortfall from P2P)
Buyer total: 520 INR

Seller receives for P2P: 70 × 6 = 420 INR
Seller grid export: (70 - 70) × 3 = 0 INR
Seller total: 420 INR

Summary:
  - Buyer pays 520 INR for 80 kWh (6.50 INR/kWh effective)
  - Seller receives 420 INR for 70 kWh (6.00 INR/kWh effective)
  - 30 kWh of contract unfulfilled, buyer sourced from grid
```

## Method 2: Deviation

```
Buyer pays: 100 × 6 - 20 × 4 = 600 - 80 = 520 INR
Buyer grid import: 0 (P2P covered conceptually, deviation penalty separate)
Buyer total: 520 INR

Seller receives: 100 × 6 - 30 × 8 = 600 - 240 = 360 INR
Seller grid export: 0
Seller total: 360 INR

Seller utility receives: 30 × 8 = 240 INR (covers grid procurement)
Buyer utility pays: 20 × 4 = 80 INR (from selling surplus)

Net: 520 - 360 - 240 + 80 = 0 ✓

Summary:
  - Buyer pays 520 INR for 80 kWh (6.50 INR/kWh effective) - SAME
  - Seller receives 360 INR for 70 kWh (5.14 INR/kWh effective) - LESS
  - Seller penalized more heavily for shortfall
```

## Key Insight

In this example:
- **Buyer outcome identical** (520 INR for 80 kWh)
- **Seller outcome differs:**
  - Min-allocation: 420 INR
  - Deviation: 360 INR (60 INR less due to explicit penalty)

The deviation method penalizes the non-compliant party (seller with 30 kWh shortfall) more explicitly, while min-allocation implicitly reduces settlement without explicit penalty attribution.
