#!/usr/bin/env python3
"""P0 位级比对器：golden vs 平台输出。exit 0 当且仅当逐行逐字节一致。

用法：
    python3 verify.py --golden golden.jsonl --output out.jsonl [--max-report 10]
    python3 verify.py --self-test
"""
from __future__ import annotations

import argparse
import json
import os
import sys
import tempfile

HEX32 = set("0123456789abcdef")


def _norm_hex32(v: object, what: str, errs: list[str]) -> str | None:
    """归一化 32B hex（去 0x、转小写）；非法则记入 errs 并返回 None。"""
    if not isinstance(v, str):
        errs.append(f"{what}: 非字符串 {v!r}")
        return None
    s = v[2:] if v.lower().startswith("0x") else v
    s = s.lower()
    if len(s) != 64 or not set(s) <= HEX32:
        errs.append(f"{what}: 非法 32B hex（{v[:20]}…）")
        return None
    return s


def load_jsonl(path: str) -> list[tuple[int, dict]]:
    rows: list[tuple[int, dict]] = []
    with open(path, encoding="utf-8") as f:
        for ln, line in enumerate(f, 1):
            line = line.strip()
            if not line:
                continue
            try:
                rows.append((ln, json.loads(line)))
            except json.JSONDecodeError as e:
                sys.exit(f"[fatal] {path}:{ln} JSON 解析失败: {e}")
    return rows


def compare(golden_path: str, output_path: str, max_report: int) -> int:
    """返回：0 = 位级一致；1 = 有差异。"""
    g = load_jsonl(golden_path)
    o = load_jsonl(output_path)

    problems: list[str] = []
    if len(g) != len(o):
        problems.append(f"行数不一致: golden={len(g)} output={len(o)}")

    mism = 0
    for (gl, gr), (ol, orow) in zip(g, o):
        errs: list[str] = []
        if gr.get("i") != orow.get("i"):
            errs.append(f"i 不匹配: {gr.get('i')!r} != {orow.get('i')!r}")
        gm = _norm_hex32(gr.get("mix"), "golden.mix", errs)
        om = _norm_hex32(orow.get("mix"), "output.mix", errs)
        gh = _norm_hex32(gr.get("hash"), "golden.hash", errs)
        oh = _norm_hex32(orow.get("hash"), "output.hash", errs)
        if gm and om and gm != om:
            errs.append(f"mix 不匹配: {gm} != {om}")
        if gh and oh and gh != oh:
            errs.append(f"hash 不匹配: {gh} != {oh}")
        if errs:
            mism += 1
            if mism <= max_report:
                problems.append(f"  i={gr.get('i')} (golden:{gl} output:{ol}): " + "; ".join(errs))

    print(f"[verify] golden={len(g)} 行, output={len(o)} 行, 不一致={mism}")
    if problems:
        print("[verify] 问题列表（截断）:")
        for p in problems:
            print(p)
        return 1
    if len(g) == len(o) == 0:
        print("[verify][warn] 两个文件都是空的")
    print("[verify] PASS —— 位级一致")
    return 0


def self_test() -> int:
    ok = {"i": 0, "mix": "0x" + "ab" * 32, "hash": "0x" + "cd" * 32}
    ok2 = {"i": 1, "mix": "0X" + "ef" * 32, "hash": "0x" + "12" * 32}  # 大写 0X 也应归一化
    tampered = {"i": 0, "mix": "0x" + "ab" * 31, "hash": "0x" + "cd" * 32}  # 长度错
    flipped = {"i": 1, "mix": "0x" + "ef" * 32, "hash": "0x" + "21" * 32}  # 首字节翻位

    def write(rows: list[dict]) -> str:
        fd, path = tempfile.mkstemp(suffix=".jsonl")
        with os.fdopen(fd, "w", encoding="utf-8") as f:
            for r in rows:
                f.write(json.dumps(r) + "\n")
        return path

    tmp = [write([ok, ok2]), write([ok, ok2]), write([ok, tampered]), write([ok, flipped])]
    try:
        if compare(tmp[0], tmp[1], 5) != 0:
            print("[self-test][fail] 正例应通过")
            return 1
        if compare(tmp[0], tmp[2], 5) == 0:
            print("[self-test][fail] 长度非法的负例应被拦截")
            return 1
        if compare(tmp[0], tmp[3], 5) == 0:
            print("[self-test][fail] 单字节翻位的负例应被拦截")
            return 1
    finally:
        for p in tmp:
            os.unlink(p)
    print("[self-test] PASS —— 正例通过 / 长度负例拦截 / 翻位负例拦截")
    return 0


def main() -> None:
    ap = argparse.ArgumentParser(description="P0 位级比对器")
    ap.add_argument("--golden")
    ap.add_argument("--output")
    ap.add_argument("--max-report", type=int, default=10)
    args = ap.parse_args()

    if not args.golden or not args.output:
        sys.exit("用法: verify.py --golden <file> --output <file> | verify.py --self-test")
    sys.exit(compare(args.golden, args.output, args.max_report))


if __name__ == "__main__":
    if len(sys.argv) == 2 and sys.argv[1] == "--self-test":
        sys.exit(self_test())
    main()
