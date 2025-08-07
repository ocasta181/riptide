# RIPTIDE: Resilient Information Protocol for Transmission In Degraded Environments

A secure UDP-based, rsync-compatible synchronization protocol engineered for highly lossy and high-latency networks. RIPTIDE is designed as a drop-in replacement for rsync, with modern cryptography, adaptive congestion control, and robust reliability via acknowledgments (AWK), negative acknowledgments (NAK), and forward error correction (FEC). The design emphasizes composition and functional pipelines to maintain clarity, testability, and extensibility.

---

## Objectives

- Rsync-compatible semantics: single files, directories, deltas, metadata preservation.
- Secure by default: modern cryptography, mutual authentication, perfect forward secrecy.
- Highly resilient on lossy/laggy links: UDP-based, selective repeat ARQ, FEC, adaptive packet sizes.
- Full-duplex awareness: sender and receiver track shared state via AWK and AWK-ACK cycles.
- MTU-aware and adaptive: avoid fragmentation, lower size on loss patterns.
- SOLID and functional design: composition-first architecture, pure transformations, immutability, and separation of concerns.
- Efficient lockless queues: atomic circular FIFO for hot-path packet scheduling and reordering.

---

## MTU and Packetization Strategy

- MTU (Maximum Transmission Unit) is the largest IP-layer packet without fragmentation. On Ethernet with IPv4/UDP:
  - IP header ~20 bytes, UDP header ~8 bytes; payload budget ≈ MTU − 28.
  - Typical Ethernet MTU: 1500 bytes → ~1472 bytes safe payload.
- RIPTIDE defaults to a conservative payload ceiling (e.g., 1200-1400 bytes) to avoid fragmentation across common tunnels and encapsulations, and offers `--mtu` override.
- Adaptive packet sizing:
  - Monitor NAK rates, checksum failures, path RTT variance, and suspected fragmentation patterns.
  - Reduce payload size stepwise when corruption/fragmentation is suspected.
  - Probe larger sizes slowly when link quality improves.

---

## Security Architecture

- Key Exchange: X25519 (ECDH) for ephemeral key agreement (PFS).
- Mutual Authentication: Ed25519 identity keys. Options:
  - Pre-shared public keys (pinned fingerprints).
  - Trust-on-first-use with fingerprint pinning.
  - Optional X.509/TLS-style identities can be layered if needed later.
- Key Derivation: HKDF(sha256) over ECDH shared secret + transcript binding to derive per-direction traffic keys.
- Encryption: ChaCha20-Poly1305 AEAD with per-direction 96-bit nonces (monotonic counters; no reuse).
- Integrity:
  - AEAD tag provides authenticated integrity for ciphertext.
  - Additional plaintext chunk checksum (BLAKE3-128) used for AWK/NAK correlation and rsync-like block identity. This aids deduplication, idempotence, and explicit detection of payload corruption beyond transport anomalies.

---

## Handshake and Session Lifecycle

States:
- IDLE → HELLO → KEY_EXCHANGE → AUTH → SESSION_ESTABLISHED → SYNC_METADATA → TRANSFER → VERIFY → COMPLETE
- ERROR can be entered from any state on fatal conditions.

Messages:
- HELLO: version, capabilities, identity fingerprint, nonce seeds.
- KEY_EXCHANGE: X25519 public keys; transcript hash accumulates all handshake fields.
- AUTH: Ed25519 signatures over transcript hash; mutual verification.
- SESSION: session ID, negotiated parameters (MTU ceiling, FEC profile, pacing mode, window limits, crypto ciphers), initial nonces.

All subsequent messages are AEAD-encrypted with derived keys.

---

## Rsync-Compatible Synchronization Engine

- Discovery:
  - Directory tree walk, metadata capture (mode, uid/gid, mtime, symlinks).
  - File signatures: rolling weak checksum (rsync-style; e.g., Adler32 variant) + strong checksum (BLAKE3-256) for blocks.
  - Optional Merkle trees over block hashes to accelerate large-file comparisons and resumable operations.
- Delta Algorithm:
  - Receiver sends signatures for existing files/blocks.
  - Sender computes delta: emit COPY (from existing block) and LITERAL (new data) instructions.
  - Literal data is compressed (LZ4) then encrypted and packetized.
- Metadata:
  - Preserve permissions, timestamps, symlinks, extended attributes where supported.
  - Atomic rename-on-complete to ensure consistency.

---

## Packet Types and Headers

Header (encrypted except when specified for initial session bootstrap):
- Version (1)
- Type (1): HELLO, KX, AUTH, SESSION, DATA, ACK, ACK_ACK, NAK, CONTROL, FEC_PARITY, HEARTBEAT, CLOSE
- Flags (2)
- Sequence Number (8): per-stream monotonic
- Total Packets (8): populated in initial control when known (e.g., for a transfer segment) or 0 if streaming/unknown
- Timestamp (8): sender wall-clock or monotonic ticks
- Checksum (4): header checksum (unencrypted control may require), data checksum use BLAKE3-128 in payload

DATA payload:
- Nonce (12)
- AEAD Tag (16)
- Content:
  - Chunk-ID
  - Plaintext-Checksum (BLAKE3-128)
  - Stream-Offset
  - Data bytes (bounded by negotiated payload size)

ACK payload:
- For each acknowledged seq: Sequence Number, Plaintext-Checksum (BLAKE3-128)
- Optional SACK blocks to compress ranges

ACK_ACK payload:
- Mirrors ACK entries to confirm receipt of ACKs

NAK payload:
- Sequence Number that failed, expected Plaintext-Checksum, error code (e.g., checksum mismatch, decrypt failure, malformed)

CONTROL:
- Window size updates, pacing suggestions, RTT samples, loss estimates, MTU probes

FEC_PARITY:
- Parity for a coding block (e.g., up to 32 data+parity per block)

HEARTBEAT:
- Keepalive and liveness sampling under long RTTs or idle periods

CLOSE:
- Graceful session termination

---

## AWK/NAK Reliability Scheme

- Three-step positive acknowledgment:
  - Sender → Receiver: DATA(seq, checksum)
  - Receiver → Sender: ACK(seq, checksum)
  - Sender → Receiver: ACK_ACK(seq)
- Until ACK is received, sender schedules retransmissions for DATA(seq). Until ACK_ACK is received, receiver resends ACK(seq).
- Negative Acknowledgment (NAK):
  - Emitted upon checksum mismatch or decryption/authentication failure for a given seq.
  - NAK(seq, expected_checksum) signals a transmission occurred but content invalid; sender must retransmit corrected packet or re-derive content.
  - NAKs are subject to pacing and deduplication to avoid storms.

The combination ensures mutual awareness of state despite loss and corruption, and distinguishes loss from corruption.

---

## Congestion Control and Pacing

- Goal: maximize throughput while minimizing queueing delay and loss, on lossy and high-latency networks.
- Approach: BBR-style model-based control adapted for UDP:
  - Measure bottleneck bandwidth via delivered bytes per RTT from ACK reception.
  - Track minimal RTT window for propagation delay.
  - Maintain a pacing rate close to estimated bandwidth; periodically probe bandwidth.
  - Congestion window in packets (cwnd) tied to BDP (bandwidth-delay product).
- Loss/Corruption Adaptation:
  - If loss/NAK rates rise, reduce pacing and/or payload size, increase FEC redundancy within limits.
  - Karn’s algorithm for RTT with exponential backoff and jitter for retransmission timers.
- LEDBAT/low-queue footprints optionally supported for background sync modes.

---

## FEC (Forward Error Correction)

- Reed-Solomon (e.g., 255,223) via `github.com/klauspost/reedsolomon`.
- Coding blocks assembled across consecutive DATA frames:
  - For each block of N data packets, add M parity packets.
  - Adaptive redundancy: increase M as loss rises, decrease as path stabilizes.
- FEC complements ARQ:
  - Attempt decode before scheduling retransmit to amortize losses.
  - Use NAKs to accelerate recovery when corruption detected.

---

## Atomic Lockless Circular FIFO Queues

High-throughput paths employ cache-friendly, lockless ring buffers. Power-of-two capacities, head/tail indices, atomic load/store with acquire/release semantics, padding to avoid false sharing.

Recommended queues:
- OutboundQueue (SPSC or MPMC): producer is delta engine/segmenter; consumer is sender/pacer.
- RetransmitQueue: time-ordered by next-deadline; consumer is retransmit scheduler.
- ReceiveReorderBuffer: holds out-of-order DATA until in-order delivery; supports fast lookup by seq.
- AckPendingQueue: ACKs awaiting ACK_ACK; retransmits on timer.

In Go, use `sync/atomic` and carefully designed SPSC/MPMC rings for minimal allocations and lock contention. Each queue carries immutable packet descriptors; buffers are pooled to reduce GC pressure.

---

## Functional, Composition-First Pipelines

Sender pipeline (pure transforms where possible):
1. Scan → Signature → DeltaPlan (COPY/LITERAL)
2. Literal → Compress (LZ4)
3. Chunk → Compute PlaintextChecksum (BLAKE3-128)
4. FEC Encode (grouped) → Packetize
5. Encrypt (ChaCha20-Poly1305 with per-direction nonce)
6. Schedule (OutboundQueue) → Pace/Send

Receiver pipeline:
1. Receive → Decrypt/Authenticate
2. Verify PlaintextChecksum
3. FEC Buffer/Decode if needed
4. Reorder → Deliver to ApplyDelta
5. Emit ACK/NAK; Track ACK_ACK state
6. ApplyDelta → Write temp files
7. Verify → Atomic Rename

Composition:
- Each stage exposes a function with explicit input/output types.
- Pipelines are composed via higher-order functions; side effects localized to IO boundaries.
- Immutable descriptors passed between stages; buffers recycled via pools.

---

## CLI and Configuration

- Command shape mirrors rsync:
  - `riptide SRC DEST [options]`
  - Examples:
    - `riptide file.txt user@host:/path/`
    - `riptide -avz --delete /dir/ user@host:/dir/`
- Key options:
  - `--mtu=N` payload sizing ceiling; default 1400
  - `--fec=k/n` target ratio, e.g., 4/20; or `auto`
  - `--congestion={bbr,ledbat}` default `bbr`
  - `--id-key=ed25519_key` identity
  - `--peer-key=ed25519_pub` pin peer
  - `--psk=FILE` optional pre-shared key for additional binding
  - `--cipher={chacha20poly1305}` default
  - `--port=UDP_PORT` default 3703
  - `--parallel=N` parallelism factor
  - `--resume` resumable transfers
  - `--no-compress` disable compression for incompressible data
  - `--checksum` force strong checksum comparison
  - `--dry-run` plan-only

---

## Security and Threat Considerations

- MITM: prevented via mutual authentication and transcript-bound keys.
- Replay/nonce reuse: per-direction counters, session IDs, and close semantics.
- DoS via ACK/NAK storms: rate-limit control messages; coalesce ranges.
- Key protection: identity keys stored securely; hardware-backed keystores when available.
- Metadata privacy: option to encrypt filenames and metadata channels.

---

## Observability

- Structured metrics: RTT, loss, NAK rate, FEC efficiency, goodput, cwnd, pacing rate.
- Logs with correlation IDs per session.
- Trace hooks for replay in simulation.
- Prometheus/OTel exporters optional.

---

## Testing and Simulation Strategy

- Unit tests for all pure transforms: checksums, chunking, encryption/decryption, FEC encode/decode.
- Property-based tests/fuzzing on packet codecs and state machines.
- Deterministic network simulation: scripted loss/duplication/reordering, variable RTT, burst losses.
- Compatibility tests mirroring rsync semantics and edge cases.
- Long-haul tests with tc/netem or in-process netem to validate adaptation logic.
- Security tests: handshake transcript binding, identity verification, nonce correctness.

---

## Current Progress (2025-08-07)

- Core building blocks are implemented with unit tests:
  - Checksum: BLAKE3-128 utilities and tests.
  - Cryptography: X25519 KEX, HKDF-based session derivation, Ed25519 sign/verify, ChaCha20-Poly1305 AEAD wrapper and tests.
  - Handshake messages: HELLO/KX/AUTH/SESSION encoders/decoders with transcript hashing and tests.
  - Packet codecs: header, DATA, ACK, NAK, HEARTBEAT encoders/decoders with tests.
  - Lockless ring: atomic SPSC ring buffer with concurrency tests.
  - MTU helpers: payload budgeting for IPv4/UDP with AEAD overhead.
  - Framing: encrypted DATA packet framing/deframing.

- Reliability scaffolding:
  - Outbound/Inbound trackers for ACK/ACK-ACK/NAK timing with exponential backoff and jitter (with unit tests).

- Next up:
  - Wire up ACK/ACK-ACK emission/handling to the trackers.
  - Add CONTROL and FEC_PARITY packet types and codecs.
  - Integrate Reed-Solomon FEC and adaptive redundancy.
  - Implement BBR-style pacing and adaptive packet sizing.
  - Build rsync-compatible delta pipelines and CLI.

## Implementation Plan (Actionable TODO)

- [ ] Bootstrap repo (Go 1.22+), `go.mod`, CI, linting, formatting, static analysis, fuzz harness.
- [x] Define core types and interfaces (composition-first): Packet, Descriptor, Pipeline stages, Queues.
- [x] Cryptographic primitives: X25519 KEX, HKDF, Ed25519 auth, ChaCha20-Poly1305 wrappers, nonce management.
- [x] Handshake protocol: HELLO/KX/AUTH/SESSION message formats and transcript binding.
- [x] Packet codecs: header encode/decode, DATA/ACK/ACK_ACK/NAK/CONTROL/FEC_PARITY, checksums (BLAKE3-128).
- [x] Atomic ring buffers: SPSC and MPMC variants; buffer pooling and zero-copy boundaries.
- [ ] Sender pipeline: delta computation (rolling checksum + BLAKE3 strong), compression, chunking, FEC encode, encrypt, queue.
- [ ] Receiver pipeline: decrypt/authenticate, checksum verify, FEC recovery, reorder, apply delta, ACK/NAK emission.
- [x] AWK/NAK/ACK-ACK state machines with timers, SACK ranges, deduplication, and rate-limits.
- [ ] Congestion control: BBR-style pacing, bandwidth/RTT estimation, adaptive payload sizing, backoff with jitter.
- [x] MTU management: CLI cap, path probing, adaptive downsizing on loss/corruption.
- [x] CLI compatible with rsync flags subset; argument parsing and mapping to engine config.
- [ ] Metadata handling: permissions, timestamps, symlinks, atomic rename-on-complete; cross-platform nuances.
- [ ] Resume/Checkpoint: content hashing and Merkle indices to support restarts and partial transfer continuation.
- [ ] Observability: metrics, logs, traces; hooks for simulation and test harnesses.
- [ ] Security hardening: key handling, fingerprint pinning store, session key rotation policy.
- [ ] End-to-end tests: local, lossy simulated network, large files, mixed directories, churn.
- [ ] Performance tuning: CPU profiles, GC reductions, ring sizing, parallelism auto-tuning.
- [ ] Packaging: binaries, reproducible builds, release pipeline; user docs.

---

## Milestones

1. M1: Encrypted UDP channel with handshake and authenticated session; ping/heartbeat with ACK/ACK-ACK.
2. M2: Packetization with checksums, AWK/NAK/ACK-ACK reliability, basic retransmission.
3. M3: FEC integrated; reorder buffers; adaptive packet sizing.
4. M4: BBR-style congestion control; pacing; stability on lossy links.
5. M5: Rsync-compatible delta and metadata; basic directory sync.
6. M6: Resume, observability, test coverage, performance targets.

---

## Notes on Composition

Every stage is a composable function transforming immutable descriptors. Pipelines are built by composing small pure stages with side-effect boundaries at IO and cryptography. This maximizes clarity, enables isolated testing, and aligns with SICP-style functional design while respecting Go’s pragmatic concurrency model.
