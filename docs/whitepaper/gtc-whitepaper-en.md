# GTC (GPU Chain): A GPU-Native, Bitcoin-Isomorphic Proof-of-Work Blockchain

**Whitepaper Draft v0.9 · August 2026**

> Status: technical parameters are frozen only upon completion of Phase 0 cross-platform
> test vectors (five platforms, ≥10⁶ samples each). **v1.0 will be published at P0 exit.**
> This draft is circulated for technical review; numeric parameters may change before freeze.

---

## Abstract

GTC is a Nakamoto-consensus proof-of-work blockchain that reproduces Bitcoin's ledger semantics — UTXO model, 600-second blocks, 21,000,000-unit supply with 210,000-block halvings, 2016-block difficulty retargeting — and replaces only the proof-of-work function.

The algorithm, **Geyser**, is engineered so that the economically rational mining device is a commodity consumer GPU. Each hash performs 64 rounds of 128-byte random accesses into a 6 GiB epoch dataset (8 KiB per hash), binding throughput to memory bandwidth; and executes a 256-instruction program of mixed integer and IEEE-754 floating-point operations regenerated from the previous block hash, binding efficiency to general-purpose execution units.

The design claims no absolute ASIC immunity — no honest design can. Instead it enforces two economic constraints on custom silicon: a compressed efficiency ceiling (target ≤ ~2×), and a scheduled algorithm rotation every 2–3 years that invalidates silicon before payback. Consumer GPUs are simultaneously the global cost optimum for memory bandwidth (≈ $0.8–1.1 per GB/s in 2026), which structurally excludes CPUs (~5–60× worse per hash-dollar), FPGAs (~10×), and AI datacenter accelerators (~5–9×). Unified-memory machines (Apple Silicon class) mine natively with zero protocol change.

Verification is cheap: a full node holds a 48 MiB cache, regenerates accessed dataset items on demand, and validates a block header in ~8 ms of CPU time. The chain launches with zero premine, no consensus-level development tax, a ≥ 6-month public testnet, and signed checkpoints for the first 10,000 blocks.

---

## 1. Introduction

### 1.1 Workload shape determines hardware class

SHA-256d is stateless and memory-free. That shape is maximally friendly to full-custom silicon: Bitcoin mining migrated CPU → GPU → FPGA → ASIC within four years of launch, and hash power concentrated into an oligopoly of foundry-backed manufacturers. A single modern ASIC miner outstrips a desktop computer by more than eight orders of magnitude; participation became an industrial supply-chain question rather than a software download.

Ethereum's Ethash remains the most successful counterexample: from 2015 to 2022, its gigabyte-scale DAG bound mining to memory bandwidth, and the network was secured by millions of dispersed consumer GPUs. Contemporaneous Ethash ASICs achieved only 2–3× efficiency and were attrited by dataset growth (the Bitmain E3 died of its 4 GB memory ceiling). Kaspa provides the modern counterexample: its lightweight kHeavyHash was dominated by ASICs within ~18 months.

**Thesis: anti-ASIC policy succeeds or fails on the *shape* of the workload, not on intent.** Geyser's shape is the shape of a commodity GPU.

### 1.2 Position statement

Geyser does not *prevent* ASICs. It makes them a bad investment:

1. **Bandwidth binding** — the proof is memory-bandwidth-bound; custom silicon must procure commodity memory (GDDR/HBM) in the same market as consumer-GPU vendors, at worse volume terms;
2. **Programmatic mix** — a per-block random program of 256 mixed integer/FP instructions requires a general-purpose programmable core; specialized logic can accelerate only a fraction of the workload;
3. **Agility contract** — a pre-scheduled, consensus-internal algorithm rotation every 2–3 years; for GPU miners this is a software update, for ASICs it is a re-tape-out.

### 1.3 Honest scope

This paper quantifies, rather than hides, the limits of the approach: GPU farms cannot be excluded by algorithm design; commodity hardware's resale value lowers 51%-attack cost to depreciation-plus-electricity; and in the network's early days its total hash rate is smaller than a rentable cloud fleet (§6.3, §10).

---

## 2. Chain Layer (Bitcoin-Isomorphic)

Everything except the proof-of-work and an extended header reproduces Bitcoin.

| Parameter | Value | Bitcoin parity |
|---|---|---|
| Consensus | Nakamoto longest-chain | same |
| Target block interval | 600 s | same |
| Difficulty retarget | every 2016 blocks, BTC formula, ±4× clamp | same |
| Initial subsidy | 50 GTC | same |
| Halving | every 210,000 blocks (~4 y) | same |
| Total supply | 21,000,000 GTC | same |
| Coinbase maturity | 100 blocks | same |
| Block weight limit | 4,000,000 WU (SegWit discounting) | same |
| Ledger | UTXO | same |
| Default signature | Schnorr (BIP340), ECDSA accepted | Taproot semantics |
| Addresses | Bech32m, HRP `gtc` | format same |
| Script | tapscript v1 whitelist, incl. CLTV/CSV; bare multisig replaced by key aggregation (MuSig2-style); MUL/DIV/CAT deferred | superset policy |
| OP_RETURN carrier | 1 output, ≤ 80 bytes (policy, not consensus) | same |

**Block header: 128 bytes** (a superset of Bitcoin's 80):

```
off  size  field
0    4     version
4    32    prev_hash
36   32    merkle_root
68   4     timestamp         (BIP113 median-time-past)
72   4     nBits             (compact 256-bit target, as Bitcoin)
76   8     nonce             (64-bit)
84   2     algo_version      (Geyser version; see §5)
86   32    mix_digest        (PoW mid-state digest for light verification)
118  10    reserved (0)
```

The 64-bit nonce removes extranonce churn; `mix_digest` is the analogue of Ethash's mixHash — a light client verifies a header from the 48 MiB cache without the 6 GiB dataset.

**Throughput.** Typical saturation blocks are 1.5–2 MB: ~12 TPS for SegWit-typical transactions, ~15 TPS for Taproot keypath. The 4M WU limit is deliberately *not* raised: block-size growth raises full-node cost (IBD, bandwidth, storage) and directly conflicts with the chain's distribution objective. Higher throughput belongs on Layer 2.

---

## 3. Geyser Proof of Work

### 3.1 Hash construction

One hash = 32 parallel lanes (one CUDA warp / half a wavefront / one Apple SIMD-group) × 64 rounds. Each round performs one dataset access (shared index across lanes; each lane consumes a distinct word) followed by 4 program instructions.

```
hash_geyser(header, nonce, program, dataset) -> u256:
    seed = keccak512(header_128B || nonce)
    for lane in 0..32:
        R[lane] = 64 × u32 registers  (derived from seed per lane)
        F[lane] = 64 × f32 registers  (bit-reinterpretation of R)
    for round in 0..64:
        idx  = index(R[0][0..8], round) mod N_items
        item = dataset[idx]                       # 128 B = 32 × u32
        for lane in 0..32: R[lane][k] = fnv1a(R[lane][k], item[lane])
        execute program[round*4 .. round*4+3]     # same 4 instrs, all lanes
    return keccak256(concat_lane_states(R, F))    # also written to header.mix_digest
```

Random dataset traffic per hash: 64 × 128 B = **8 KiB**. On an RTX 5090 (1.79 TB/s) the theoretical ceiling is ~220 Mh/s; expected realized rate 60–120 Mh/s — the same order as Ethash-era hardware, confirming bandwidth (not logic) is the binding constraint.

### 3.2 Epoch dataset (two-level structure)

```
epoch_seed(0) = keccak256("GTC/mainnet/genesis-seed-v1")
epoch_seed(e) = keccak256(epoch_seed(e-1))

cache(e)      = gen_cache(epoch_seed(e))     # 48 MiB, Argon2id-style sequential KDF
dataset(e)[i] = gen_item(cache(e), i)        # items independent, parallelizable
```

- **Miners** hold the full dataset (6 GiB at genesis, +32 MiB per epoch) in VRAM — or in unified memory with zero-copy, or in system RAM at a performance penalty.
- **Full nodes / light clients** hold only the 48 MiB cache and regenerate the ≤ 64 accessed items per header: ~8 ms CPU. Initial block download does not require the dataset.
- Epoch length is 2016 blocks (~14 days), aligned with the difficulty window. Miners pre-generate at 1024 blocks before the boundary; datasets distribute over the p2p network (Ethereum-proven practice).

### 3.3 Program layer

```
program_seed = keccak256(prev_block_hash || epoch_seed)
rng          = xorshift128+(program_seed)
program      = 256 instructions:
               dest  ∈ [8, 64)      # first 8 registers are mix state, write-protected
               src1  ∈ [0, 64)
               src2  ∈ [0, 64) or immediate
               opcode ∈ instruction set
               one FP instruction per 8 (fixed pattern)
final 32 ops: merge — F registers bit-reinterpreted and multiplied back into R
```

**Instruction set (v1: 18 integer + 3 floating-point):**

| Class | Instructions |
|---|---|
| Arithmetic | iadd, isub, imul (u32), imulhi, umin, umax |
| Bitwise | and, or, xor, andnot, shl, shr, rol, ror |
| Bit ops | popcount, clz, bswap, brev |
| Mix | fnv1a-mul |
| Floating-point | fadd, fmul, fma (IEEE-754 f32, round-to-nearest-even) |

**Determinism rules** (floating point is where prior algorithms broke):

1. No transcendental functions, no division, no square root — these are not bit-identical across vendors;
2. FMA *is* permitted: IEEE-754 defines fma as infinite-precision multiply-add with single rounding, bit-identical on all compliant hardware;
3. Denormals handled explicitly: every FP instruction is followed by a compare-and-select flushing subnormal results to zero (one select, negligible cost), neutralizing the GPU-flushes/CPU-doesn't divergence;
4. P0 publishes cross-verified test vectors (NVIDIA, AMD, Intel, Apple GPU, ARM64 CPU; ≥10⁶ program–state pairs per platform) — bit-exact agreement is the freeze criterion.

Compilation discipline: `--use_fast_math` / `-ffast-math` forbidden; `--ftz=false` set explicitly (belt-and-suspenders over the algorithmic rule).

### 3.4 Difficulty

Identical to Bitcoin: `nBits` compact-encodes the 256-bit target; a block is valid when `mix_digest ≤ target`. Retargeting every 2016 blocks with the ±4× clamp, unmodified.

### 3.5 Dataset growth policy (VRAM policy)

| Item | Value |
|---|---|
| Initial dataset | 6 GiB |
| Growth | +32 MiB per epoch (≈ 0.83 GiB/yr) |
| Binding policy | dataset ≤ 60% of min(median gaming-GPU VRAM, median UMA-machine memory) |
| Review cadence | every 2 years; growth may be frozen or slowed via `algo_version` |

Growth keeps raising the memory BOM floor for custom silicon — the mechanism that killed the Ethash E3 — while the 60% cap keeps the median consumer card viable (6 GiB + headroom fits 8 GB cards today; projected ~9.3 GiB vs 16–24 GB median in 2030).

---

## 4. Algorithm Agility

`algo_version` (header offset 84) selects the full parameter set. Scheduled rotations are consensus-internal hard forks with heights fixed at genesis:

| Version | Planned height | Change |
|---|---|---|
| 0x0001 | genesis | Geyser v1 freeze |
| 0x0002 | ~ year 3 | instruction-set extension (+4–6 ops), FP ratio adjustment |
| 0x0003 | ~ year 6 | growth-policy review; lane restructuring |
| 0xFFxx | emergency | pre-positioned spec; parameters undisclosed, clients pre-ship implementation |

The asymmetry is the defense: a GPU miner's upgrade cost is a software download; an ASIC's is a re-tape-out with NRE re-paid. Scheduled rotation intervals (2–3 y) sit below the empirical ASIC payback period (~18–24 months), making the investment irrational even under a best-case efficiency advantage. Monero's 2018–2019 rotations — four consecutive successful ASIC evictions — are the operative precedent for this social contract.

---

## 5. Hardware Economics

### 5.1 Bandwidth cost (2026 street prices)

| Platform | Bandwidth | Price | $/GB/s | vs. baseline |
|---|---|---|---|---|
| RTX 5070 (baseline) | 672 GB/s | $550 | ~0.82 | 1× |
| RX 9070 XT | 640 GB/s | $600 | ~0.94 | ~1× |
| RTX 5090 | 1.79 TB/s | $2,000 | ~1.12 | ~1.4× |
| RTX 4060 | 272 GB/s | $300 | ~1.10 | ~1× |
| Desktop CPU (2-ch DDR5-6000) | ~96 GB/s | ~$400 | ~4.2 | ~5× (latency-limited in practice) |
| Server EPYC (12-ch DDR5) | ~576 GB/s | $8,000+ | ~14 | ~16× |
| NVIDIA H100 | 3.35 TB/s | ~$25,000 | ~7.5 | ~9× |
| NVIDIA H200 | 4.8 TB/s | ~$32,000 | ~6.7 | ~8× |
| NVIDIA B200 | 8 TB/s | ~$40,000 | ~5.0 | ~6× |
| Apple M4 Max (UMA) | 546 GB/s | ~$2,000 system | ~3.7 | ~4.5× |
| Apple M3 Ultra (UMA) | 819 GB/s | ~$4,000 system | ~4.9 | ~6× |
| FPGA board (4-ch DDR4) | ~100 GB/s | $5,000+ | ~50 | ~60× |

$/GB/s is a *necessary* condition; the platform must also have parallel compute units capable of saturating it. GPUs (discrete and UMA-integrated) qualify; CPUs are latency-bound far below peak; FPGAs lack the bandwidth outright.

**Consumer GPUs are the global cost optimum for memory bandwidth.** Binding proof-of-work to bandwidth locks the rational mining device to that optimum.

### 5.2 Best-case ASIC analysis

Assume a vendor perfectly implements Geyser silicon: memory purchased in the same GDDR/HBM market as GPU vendors (worse volume terms); logic stripped of display/ray-tracing (−~30% die); self-built board/power/cooling; NRE $10–30M amortized over tens-of-thousands-scale volumes versus the consumer-GPU market's tens of millions. Realized BOM advantage ~25–35% — largely consumed by NRE amortization and volume disadvantage — before the rotation schedule (§4) destroys residual payback. The historical calibration point: real-world Ethash ASICs achieved 2–3× against a *weaker* algorithm (no program layer, no rotation). Geyser's expected ceiling is lower.

### 5.3 Platform matrix ($ per hash, relative to a 5070-class card = 1.0)

| Platform | Relative cost | Verdict |
|---|---|---|
| Consumer GPU (≥8 GB) | 1.0 | rational choice |
| Custom ASIC (best case) | 0.5–0.7, negative after rotation risk | irrational to build |
| FPGA | ~10× | excluded |
| CPU | ~20–60× | excluded |
| AI accelerator | ~5–15× | excluded |
| iGPU (UMA laptops) | ~10× | marginal |

---

## 6. Adversarial Analyses

### 6.1 AI datacenter clusters

Stated plainly: AI accelerators are general-purpose hardware; they execute Geyser's program layer at full speed and hold the dataset trivially. **Nothing architectural stops them; the defense is purely economic.**

- **Mining**: H100/H200/B200 carry a structural 5–9× $/GB/s disadvantage. The gap is generation-stable (H100 vs. 4090 ≈ 4.7×; B200 vs. 5090 ≈ 4.4×) because AI cards and consumer cards ride the same memory generation — the HBM premium is datacenter margin plus advanced packaging, not a transient technology gap. Renting at spot rates (~$2.5–3.5/hr for ~5× a 5070's throughput at ~50× the hourly cost) loses ~10× per hash.
- **Burst 51% attack**: against an early network of 50,000 consumer cards, a 51% attack needs ~10,000 H100s ≈ $30k/hr — paying a 5–10× "AI premium" per hash, sustained, versus defenders who only need to hold the main chain. Sweeping spot capacity across providers moves public prices — itself a visible early-warning signal. Attack cost grows linearly with honest hash rate; spot depth does not.
- **Post-AI-winter idle fleets**: the only scenario where AI hardware's marginal cost (electricity only) beats consumer GPUs. The threat then is concentration (>50% single entity), not mining per se. Mitigations are operational: pool-concentration monitoring, anomaly alerts, emergency-checkpoint social coordination (§10). The opportunity-cost gate — AI workloads historically pay 10–20× mining per hour — must fail first.

**Conclusion: AI clusters are the most expensive hash rate, not the strongest adversary.** ASIC defense argues "they would be too cheap"; AI-cluster defense argues "they are too expensive." Both ends are closed.

### 6.2 Unified-memory architecture (UMA)

Apple Silicon, Strix Halo-class SoCs, and successors share memory between CPU and GPU. Geyser binds to *memory bandwidth* — a physical quantity — not to *discrete VRAM* — a packaging choice. Consequences:

- **Zero consensus change.** The dataset resides in unified memory and is randomly accessed by the GPU with no copy penalty; UMA is in fact the optimal dataset-residency form.
- **Mac-class machines are legitimate miners**: an M4 Max (546 GB/s) realizes ~30–40 Mh/s ≈ 0.8× an RTX 5070; an M3 Ultra reaches 4080-class. LPDDR energy efficiency is higher (~0.26 vs ~0.42 W per GB/s including system) — per-hash energy is *lower* on UMA.
- **No new centralization vector**: system $/GB/s remains 4.5–8× worse than discrete GPUs; sunk-cost mining (an owned Mac's marginal cost is electricity) is a distribution *gain*, symmetric with idle gaming PCs.
- **Engineering cost**: a third GPU backend (Metal). Apple's SIMD-group is also 32 — lane structure unchanged. The §3.3 denormal rule pre-neutralizes Apple's FTZ default. Test vectors extend to five platforms.

If UMA becomes the mainstream PC form factor, the miner base widens from "computers with a discrete GPU" to "essentially all modern computers" — closer to the original one-CPU-one-vision of Satoshi, with GPU-grade bandwidth economics intact.

### 6.3 GPU-side concentration (honest ledger)

Farms cannot be excluded by algorithm design. Mitigations: radical accessibility (one card mines), Stratum v2 (small miners keep block-construction rights without centralized pool proxies). Commodity-hardware resale lowers attack cost to depreciation + electricity — the inherent price of distribution; symmetrically, ASICs' zero resale value binds miners to their manufacturer, itself a centralizing force. Early-network vulnerability to cloud rental is addressed at launch (§8).

---

## 7. Data Anchoring (Neutrality Policy)

GTC is content-neutral at consensus. Chain-embedded data cannot meaningfully be censored (witness signature grinding embeds arbitrary bits) — so, like Bitcoin, the chain *channels* data demand into the most harmless carrier rather than pretending to prevent it:

- **OP_RETURN policy**: one output per transaction, ≤ 80-byte payload (mempool policy, not consensus — adjustable without a fork). 80 B fits a 32-byte digest + protocol tag + metadata, and is byte-compatible with OpenTimestamps attachment format.
- OP_RETURN outputs are unspendable → **zero UTXO pollution** (without the policy, users embed data in fake P2PKH outputs, which do pollute).
- **Aggregation pattern**: N evidence hashes → Merkle tree → root on-chain. A saturated block carries 0.6–0.9M bare anchors/day; with aggregation the limit is fees, not capacity.
- **Credibility boundary, stated honestly**: evidentiary weight tracks the anchoring chain's own credibility; a new chain is weaker than Bitcoin. The recommended pattern is **dual anchoring** — high-frequency anchoring on GTC (cheap), periodic terminal anchoring on Bitcoin — at zero additional design cost via OTS compatibility.
- Raw files never go on-chain; privacy and content compliance stay off-chain. GTC carries hashes, not content, and assumes no regulatory surface.

Anchoring is fee-paying, UTXO-neutral, and independent of token speculation — a sustainable fee source for the long-term security budget as subsidy decays.

---

## 8. Launch & Governance

| Item | Decision |
|---|---|
| Premine | **zero**; no instamine |
| Dev funding | none in consensus (no dev tax); donations/foundations only |
| Genesis | public spec + three-platform client binaries + ≥ 6-month public testnet, then a scheduled UTC genesis |
| Early defense | signed checkpoints for the first 10,000 blocks (official channels), precluding private-chain extensions |
| Mining protocol | Stratum v2 (reference clients ship it): small miners retain block-construction rights |
| Algorithm governance | parameter changes = pre-scheduled `algo_version` forks; BIP-style spec process; multiple client implementations as checks |

---

## 9. Token Economics

Deliberately Bitcoin-identical: 50 GTC genesis subsidy, halving every 210,000 blocks, terminal supply 21,000,000 GTC, 100-block coinbase maturity. 144 blocks/day → 7,200 GTC/day first-year emission. Simplicity is the feature: no new emission schedule to argue about, one security-budget curve, one fee market. Long-term security is carried by fee demand (transaction fees + anchoring demand, §7) against declining subsidy — the same equilibrium Bitcoin faces, faced honestly rather than re-engineered.

---

## 10. Risk Register (condensed)

| # | Risk | Level | Mitigation |
|---|---|---|---|
| R1 | ≥3× ASIC within 3 years | medium | emergency rotation; hashrate-anomaly monitoring |
| R2 | cross-platform FP divergence | medium | §3.3 determinism rules; P0 five-platform vectors are a hard gate |
| R3 | VRAM market stagnation strands miners | low-med | growth policy freeze mechanism (§3.5) |
| R4 | early-network cloud/AI rental 51% | high (early) | signed checkpoints; testnet-inflated initial hash rate; decays with growth |
| R5 | GPU farm concentration | medium | Stratum v2; accessibility; accepted as distribution cost |
| R6 | community split at rotation | low | heights fixed at genesis; rotation is consensus-internal |
| R7 | epoch-boundary orphans | low | pre-generation at −1024 blocks; p2p dataset distribution |
| R8 | AI cycle (supply absorption / idle fleets) | low-med | opportunity-cost gate; pool-concentration monitoring |
| R9 | UMA mainstreaming widens backend surface | low | backend abstraction; Metal scheduled P2 |
| R10 | UMA thermals/paging degrade miner UX | low | miner-side engineering; not a protocol risk |
| R11 | data-embedding abuse | low | OP_RETURN channeling + fee market |

---

## 11. Roadmap

| Phase | Deliverable | Exit criterion |
|---|---|---|
| **P0** | spec freeze; five-platform test vectors (NVIDIA / AMD / Intel / Apple GPU / ARM64 CPU; ≥10⁶ samples each); whitepaper v1.0 | bit-exact agreement across all five |
| **P1** | Go full node (UTXO, consensus, p2p); CUDA + ROCm miners; Stratum v2 | public testnet running continuously |
| **P2** | incentivized testnet (valueless coins); Metal backend; consensus + cryptography audits; attack exercises | ≥ 6 months stable + two audits passed |
| **P3** | mainnet genesis with signed checkpoints | — |

Reference client in Go (btcd/secp256k1 lineage for portability); miners open-source on CUDA, ROCm, and Metal. No closed-source mining software is part of the reference stack — closed miners are a trust problem this project declines to create.

---

## 12. References

- Buterin & Dryden, *Ethereum Design Rationale* (Dagger/Ethash epoch dataset)
- Minehan et al., *ProgPow specification* (programmatic mix; GPU-pipeline occupancy argument)
- Ravencoin *KawPow* activation and three-year operational record (cross-vendor FP consistency practice)
- tevador, *RandomX specification* (shape-matching thesis; Monero 2018–19 anti-ASIC fork history)
- Ergo *Autolykos v2* (dataset-in-VRAM design)
- Kaspa kHeavyHash ASIC timeline, 2021–2023 (negative precedent)
- BIP-113, BIP-141/143, BIP-340, BIP-341/342/343, BIP-350
- Stratum v2 specification (Braiins)
- OpenTimestamps attachment format

---

*Geyser = Ethash's memory core + KawPow's program layer + Monero's rotation contract + an explicit VRAM growth policy. The chain around it is Bitcoin's, deliberately unchanged.*

*— Draft v0.9. Comments to the technical review list. Numeric parameters freeze at P0 exit.*
