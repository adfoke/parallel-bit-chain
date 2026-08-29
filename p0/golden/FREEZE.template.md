# P0 规范冻结记录（FREEZE）

> 复制本模板为 FREEZE.md 并填写。填写完成之日 = P0 出口 = 白皮书 v1.0 发布日。

- 冻结日期：YYYY-MM-DD
- 规范版本：DESIGN.md commit `<hash>`（vX.Y.Z）· 白皮书 v1.0（双语，commit `<hash>`）
- `algo_version`：1
- 向量协议：format_version 1

## 向量集

| 项 | 值 |
|---|---|
| epoch | 0 |
| epoch_seed | `0x…`（回填后非 null） |
| program_seed | `0x…` |
| full 样本量 | ≥ 10^6 |
| full golden SHA256 | `<hash>`（release artifact URL: …） |
| smoke golden SHA256 | `<hash>`（git 内） |

## 平台矩阵（五方全绿记录）

| 平台 | 实现 | 硬件 | 编译器/驱动版本 | 结果 |
|---|---|---|---|---|
| NVIDIA | geyser-cuda `<commit>` | | | PASS |
| AMD | geyser-hip `<commit>` | | | PASS |
| Intel | geyser-oneapi `<commit>` | | | PASS |
| Apple | geyser-metal `<commit>` | | | PASS |
| ARM64 CPU | geyser-cpu `<commit>` | | | PASS |

## 未决项清零确认

- [ ] TODO#1 epoch_seed 常量已回填
- [ ] TODO#2 algo_version 字节序已冻结（LE/BE）
- [ ] TODO#3 program_seed 派生映射已冻结
- [ ] TODO#4 header nonce 与附加 nonce 关系已冻结
- [ ] TODO#5 缩容 profile 决议已记录（引入与否）

## 签核

- 签核人 / 日期：
