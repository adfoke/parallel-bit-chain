# PTC × Bitcoin 工程对比

> v1 · 2026-08-29 · 配套 DESIGN v0.3.9
> 官网 mirror 区是营销版（7 行）；本文件是工程版，逐项附规范引用。
> 定性：PTC 对 Bitcoin 是**净室重实现 + 更换共识引擎**，不是 fork——零行 Bitcoin Core 代码，复刻的是语义。

---

## 1. 完全相同的账本语义（"Bitcoin 同构"的全部含义）

| 维度 | 值（两链一致） | 规范引用 |
|---|---|---|
| 共识规则 | Nakamoto 最长链 | DESIGN §4.1 · WP §2 |
| 出块间隔 | 600 秒 | DESIGN §4.1 |
| 供给曲线 | 50/块 → 每 210,000 块减半 → 终量 21,000,000 | DESIGN §4.1 · WP §9 |
| 难度调整 | BTC 公式：2016 块窗口，±4x 限幅，nBits 紧凑编码 | DESIGN §4.1 · 决议 #4 |
| 交易模型 | UTXO，SegWit 重量制，区块 4,000,000 WU | DESIGN §4.1/§4.3 |
| 签名 | Schnorr（BIP-340）默认，ECDSA 兼容 | DESIGN §4.5 · 决议 #5 |
| 脚本 | Taproot/tapscript，CLTV/CSV 时间锁，MuSig2 多签 | DESIGN §4.5 · 决议 #3 |
| 地址 | Bech32m（BIP-350），HRP `ptc` | DESIGN §4.1 |
| 时间语义 | BIP113 median-time-past | DESIGN §4.2 |
| Coinbase 成熟期 | 100 块 | DESIGN §4.1 |

## 2. 有意的结构性差异

| 维度 | Bitcoin | PTC | 为什么 | 引用 |
|---|---|---|---|---|
| PoW | SHA256d（无状态，ASIC 结构性最优） | **Geyser**：带宽绑定 + 每块随机程序 + algo_version 轮换 | 项目存在的理由：压制 ASIC 至 ≤2x 并毁掉回本 | DESIGN §5/§6 · WP §3 |
| 区块头 | 80 B | **128 B**：+8B nonce（防回绕）、+2B algo_version、+32B mix_digest | 轻客户端 48 MiB cache 即可验头，无需 6 GiB dataset | DESIGN §4.2 |
| tapscript | 全量操作码 | **白名单子集**（禁裸 multisig、禁 MUL/DIV/CAT；多签走 MuSig2） | 攻击面收敛 | DESIGN §4.5 · 决议 #3 |
| 算法治理 | 无内建机制，靠社会共识硬分叉 | **algo_version 预定硬分叉**（2–3 年轮换日程） | 共识内建的敏捷性锚点 | DESIGN §6.6 · WP §4 |
| 启动防御 | 2009 年无人竞争 | 签名 checkpoint 前 10,000 块 + ≥6 月公开测试网 | 2026 年启动的新链必须防早期私有链延展 | DESIGN §9 |
| 抗量子阶梯 | 无内建 | 输出类型 v0/v1/v2 + 能力里程碑触发 | 现在预留扩展点，吸取 BTC 事后改造难的教训 | DESIGN §7.7 · WP §6.4 |
| 共识内数据政策 | OP_RETURN 80B（历史演化出的默认） | OP_RETURN 80B（创世即定，OTS 兼容） | 收编数据需求至最无害载体，UTXO 零污染 | DESIGN §4.6 |

## 3. mempool 与 L2 政策

| 项 | Bitcoin | PTC | 引用 |
|---|---|---|---|
| RBF | 2015 年 BIP-125 opt-in，十年争论 | **创世即 BIP-125 opt-in**（照抄终态语义） | DESIGN §4.7 |
| CPFP | 实战演化（祖先包费率） | **创世即支持** | DESIGN §4.7 |
| package relay | 2023–24 年才为 LN 补齐 | **创世即支持**（包提交 + 包内 RBF） | DESIGN §4.7 |
| TRUC(v3)/锚输出 | 逐步激活（LN 曾被 pinning 卡多年） | **创世启用 TRUC + ephemeral anchors** | DESIGN §4.7 |
| 通道构造 | HTLC 起家，向 PTLC 迁移中（历史包袱） | **PTLC-native**（Schnorr adaptor，day-one） | DESIGN §4.8 |
| 通道工厂/splice | 给旧格式留兼容，慢慢推 | 按 2026 终态协议设计 | DESIGN §4.8 |
| rollup 型 L2 | 有脚本执行平台，canonical rollup 可行 | **不可行（by design）**：tapscript 白名单无执行平台；sovereign 弱形态=OP_RETURN 数据锚，不是 rollup | DESIGN §4.8 · 决议 #3 |
| eltoo | 待 SIGHASH_ANYPREVOUT（未激活） | 延后：APO 不在 v1 白名单；通道走 penalty + PTLC | DESIGN §4.8 |

## 4. 无疤红利清单（新链的复利）

Bitcoin 的部分复杂度是**历史手术疤痕**（P2PKH→P2SH→SegWit→Taproot 一路兼容）。PTC 创世即终态：

1. **day-one 100% Taproot**：无遗留地址类型，链上全是 Schnorr。Bitcoin 至今 Taproot 采用率仍为少数
2. **PTLC-native**：通道按点时间锁设计，多跳隐私优于 HTLC，且无迁移债
3. **mempool 终态政策**：Bitcoin 十年打磨出的通道安全套餐，PTC 零成本携带
4. **通道协议按终态设计**：工厂、splice 不给 2015 年格式留兼容
5. **bech32m-only**：无 BIP-350 的 v0 历史包袱

## 5. Bitcoin 的护城河（诚实账）

| 维度 | Bitcoin | PTC |
|---|---|---|
| 算力历史 | 17 年未被打穿，攻击成本以 GW 计 | 0，创世从零爬坡（R4：启动期靠 checkpoint 防御） |
| 公信力 | 法币级社会共识 | 新链弱于 BTC 是客观事实——**双锚定的存在理由**（DESIGN §4.6） |
| 生态 | 钱包/交易所/矿业/闪电网络全齐 | 全部自建（P1 起） |
| 多客户端 | Core 之外多个独立实现制衡 | 目前只有规范，零实现（P0–P1 补） |
| 矿业供应链 | ASIC 深度市场 | 依赖消费级 GPU 市场被动供给 |

## 6. 结论

**账本语义可以复制，信任积累复制不了。** 因此 PTC 不与 Bitcoin 竞争价值存储叙事——镜像的定位是去 Bitcoin 结构性到不了的地方：机器经济的需求侧（费用市场、机器公证、通道结算）。平行线，永不相交；双锚定让两条链从竞争变为分工。

---

*维护规则：DESIGN.md 更新触及对比内容时同步本文件；版本随 DESIGN 版本号走。*
