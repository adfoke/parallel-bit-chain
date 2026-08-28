# GTC (GPU Chain)

一条 **GPU-only 挖矿**的 PoW 公链。ticker: **GTC**。当前处于设计阶段，尚无代码。

**核心思路**：链层最大化复刻 Bitcoin（UTXO、10 分钟出块、2100 万枚、四年减半），唯一的核心差异是共识算法 —— 不是"禁止 ASIC"（技术上不可能），而是让"造 ASIC 不如买显卡"。

## 设计一览

| 项 | 决定 |
|---|---|
| 共识 | Nakamoto 最长链，BTC 难度公式（2016 块，±4x 限幅） |
| 出块 / 供给 | 600 秒；50 GTC 起步，210,000 块减半，总量 21,000,000 |
| 区块大小 | 4,000,000 WU（SegWit 重量制），典型 1.5–2 MB，不扩 |
| 交易 / 脚本 | UTXO + Taproot；Schnorr 默认、ECDSA 兼容；tapscript 白名单（含 CLTV/CSV 时间锁，多签走 MuSig2） |
| 地址 | Bech32m，HRP `gtc` |
| PoW | **Geyser**（见下），区块头 128 B（+`algo_version` +`mix_digest`） |
| 启动 | 零预挖 / 零 dev tax；公开测试网 ≥6 个月；Stratum v2 |

## Geyser：三层抗 ASIC 防御

1. **带宽绑定**（学 Ethash）—— 每次哈希随机读 6 GiB 数据集 8 KiB，瓶颈锁死显存带宽；消费级 GPU 是全平台 $/GB/s 的最优点，CPU / FPGA / AI 加速器结构性出局
2. **随机程序层**（学 KawPow）—— 每块 PoW 程序由上块哈希生成（256 条混合 INT/FP 指令），ASIC 想跑赢必须"造一个 GPU"
3. **算法轮换契约**（学 Monero）—— `algo_version` 预定硬分叉每 2–3 年轮换；GPU 矿工 = 更新软件，ASIC = 流片作废

目标：ASIC 效率优势压制在 **≤ 2x** 且无法回本；dataset 以 ≤ 同期中位数显卡 VRAM 60% 的政策增长（当前 6 GiB，+32 MiB/epoch）。

## 几个立场鲜明的取舍

- **UMA（Apple Silicon 等）**：共识零改动，Mac 是合法矿机（M4 Max ≈ 0.8×RTX 5070，每 hash 电耗更低）；Metal 后端进 P2 路线
- **区块不扩**：一边用 Geyser 保矿工分布、一边用大区块毁节点分布是左手打右手；要吞吐走 L2
- **存证 / 锚定**：共识中立，OP_RETURN 80 B 政策收编（兼容 OpenTimestamps）；日常锚 GTC、终局双锚 BTC，链上只放哈希
- **诚实清单**：GPU 农场无法用算法排除、新链启动期最脆弱、绝对 ASIC 免疫不存在 —— 全部写在文档 §7.4 / §11，不装看不见

## 文档与路线图

- 白皮书（草案 v0.9，双语）：**[English](docs/whitepaper/gtc-whitepaper-en.md)**（canonical）· **[中文](docs/whitepaper/gtc-whitepaper-zh.md)** —— 纯技术向；v1.0 于 P0 测试向量冻结时发布
- 设计文档：**[docs/DESIGN.md](docs/DESIGN.md)**（v0.3.4）—— 含威胁模型、Geyser 完整规格、带宽经济学量化、前人算法成败史（Ethash / Kaspa / RandomX / KawPow）与风险登记册

| 阶段 | 内容 |
|---|---|
| **P0** | 规范冻结；五方测试向量（NVIDIA / AMD / Intel / Apple GPU / ARM64 CPU）位级一致；白皮书 v1.0 发布 |
| **P1** | Go 全节点 + CUDA/ROCm 矿工 + Stratum v2 |
| **P2** | 激励测试网 ≥6 个月、Metal 后端、双审计 |
| **P3** | 主网启动（签名 checkpoint 防早期私有链） |

设计层未决问题已清零（决议记录见文档 §12）。P0 数值参数在测试向量发布前仍可微调。
