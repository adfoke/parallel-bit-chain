#!/usr/bin/env python3
"""P0 输入向量生成器 —— 确定性、仅标准库。

generator.algorithm = sha256-counter-v1:
    b0      = sha256(f"{seed}|{i}|{tag}")
    b_{k+1} = sha256(b_k)                          # 链式展开
    header(i) = b0‖b1‖b2‖b3 截取 128B（tag="header"）
    nonce(i)  = sha256(f"{seed}|{i}|nonce")[:8]

字段约束（向量协议，见 p0/README.md）：
    header[76:84]   nonce 字段，与 nonce(i) 同值（TODO#4）
    header[84:86]   algo_version，小端（TODO#2 字节序未冻结）
    header[86:118]  mix_digest = 0（哈希输入要求）
    header[118:128] reserved = 0
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os

# DESIGN.md 附录 A（Geyser v1）参数快照 —— 冲突时以 DESIGN.md 为准
APPENDIX_A = {
    "EPOCH_BLOCKS": 2016,
    "DATASET_BYTES_INIT": 6442450944,
    "DATASET_GROWTH_BYTES_PER_EPOCH": 33554432,
    "CACHE_BYTES_INIT": 50331648,
    "CACHE_GROWTH_BYTES_PER_EPOCH": 262144,
    "DATASET_ITEM_BYTES": 128,
    "ACCESS_ROUNDS": 64,
    "LANES": 32,
    "PROG_OPS": 256,
    "REGISTERS": "64 x u32 + 64 x f32 per lane",
    "HASH_INIT": "keccak512",
    "HASH_FINAL": "keccak256",
}


def _sha(b: bytes) -> bytes:
    return hashlib.sha256(b).digest()


def _expand(tag: str, n: int) -> bytes:
    out = b""
    cur = _sha(tag.encode())
    while len(out) < n:
        out += cur
        cur = _sha(cur)
    return out[:n]


def make_sample(seed: str, i: int, algo_version: int) -> tuple[bytes, bytes]:
    header = bytearray(_expand(f"{seed}|{i}|header", 128))
    nonce = _sha(f"{seed}|{i}|nonce".encode())[:8]
    header[76:84] = nonce                                   # TODO#4：与附加 nonce 同值
    header[84:86] = algo_version.to_bytes(2, "little")      # TODO#2：字节序未冻结
    header[86:118] = b"\x00" * 32                           # mix_digest 必须为零
    header[118:128] = b"\x00" * 10                          # reserved
    return bytes(header), nonce


def main() -> None:
    ap = argparse.ArgumentParser(description="P0 确定性输入向量生成器")
    ap.add_argument("--count", type=int, default=1024)
    ap.add_argument("--seed", default="p0-smoke-1")
    ap.add_argument("--out-dir", default="vectors/smoke")
    ap.add_argument("--profile", choices=["smoke", "full"], default="smoke")
    ap.add_argument("--algo-version", type=int, default=1)
    ap.add_argument("--epoch", type=int, default=0)
    ap.add_argument("--program-seed", default=None,
                    help="32B hex；缺省 null（TODO#3：ref 冻结时回填）")
    args = ap.parse_args()

    if args.profile == "full" and args.count < 1_000_000:
        print(f"[warn] full profile 建议 count >= 10^6（当前 {args.count}）")

    os.makedirs(args.out_dir, exist_ok=True)

    manifest = {
        "format_version": 1,
        "profile": args.profile,
        "algo_version": args.algo_version,
        "epoch": args.epoch,
        "epoch_seed": None,             # TODO#1：keccak256("PTC/mainnet/genesis-seed-v1")，ref 回填
        "program_seed": args.program_seed,  # TODO#3
        "dataset": APPENDIX_A,
        "samples": args.count,
        "generator": {
            "tool": "p0/harness/gen_inputs.py",
            "seed": args.seed,
            "algorithm": "sha256-counter-v1",
        },
    }
    with open(os.path.join(args.out_dir, "manifest.json"), "w", encoding="utf-8") as f:
        json.dump(manifest, f, ensure_ascii=False, indent=2)
        f.write("\n")

    with open(os.path.join(args.out_dir, "inputs.jsonl"), "w", encoding="utf-8") as f:
        for i in range(args.count):
            header, nonce = make_sample(args.seed, i, args.algo_version)
            f.write(json.dumps(
                {"i": i, "header": "0x" + header.hex(), "nonce": "0x" + nonce.hex()},
                separators=(",", ":"),
            ) + "\n")

    print(f"[ok] {args.count} samples -> {args.out_dir}/inputs.jsonl (+ manifest.json)")


if __name__ == "__main__":
    main()
