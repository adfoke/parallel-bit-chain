# P0 测试向量框架（骨架）

> 状态：**骨架** —— 向量格式、平台契约、验收与冻结流程已定义；参考实现（ref/）与五方 runner（platforms/）尚未实现。Geyser 核心算法零行实现，本目录只搭验收机器。

## 目标与验收标准（源自 DESIGN §10 / 白皮书 §11）

- **五平台位级一致**：NVIDIA (CUDA) · AMD (HIP) · Intel (oneAPI) · Apple (Metal) · ARM64 CPU
- 每平台跑满 **≥ 10^6 样本**，与 golden **byte-exact**（32B mix + 32B hash 逐字节相等）
- P0 出口 = 规范冻结：此后数值参数不可再动，任何改动 = `algo_version` 升版
- 白皮书 v0.9 → **v1.0** 随冻结发布

## 目录结构

```
p0/
  README.md            本文件：工作流 / 向量协议 / 未决项 / 冻结清单
  Makefile             make inputs / golden / verify / smoke
  harness/
    gen_inputs.py      确定性输入生成器（纯 stdlib，已可用）
    verify.py          位级比对器（纯 stdlib，已可用，含 --self-test）
  schema/              manifest / inputs / output 的 JSON Schema
  platforms/           五方 runner 契约 + 各平台 README（待实现）
  ref/                 参考实现（Go，P1 前置；生成 golden，待实现）
  golden/              golden 冻结记录（FREEZE.md 模板已备）
  vectors/             生成的向量集（smoke 进 git，full 不进）
```

## 工作流（五步）

```
1. make inputs                    # 确定性输入集（smoke 1k 进 git；full 10^6 本地/CI 生成）
2. make golden                    # 参考实现就绪后：生成 golden.jsonl
3. make run-<platform>            # 每平台实现 runner 后：读 inputs，算 mix/hash，写输出
4. make verify OUT=<输出文件>      # 位级比对 golden vs 平台输出
5. 五方全绿 → 填 golden/FREEZE.md → git tag → 白皮书 v1.0
```

## 向量协议

### manifest.json（每向量集一份）

| 字段 | 说明 |
|---|---|
| `format_version` | 协议版本（当前 1） |
| `profile` | `smoke`（进 git，~1k 样本）/ `full`（≥10^6，release artifact） |
| `algo_version` | Geyser 算法版本（DESIGN §6.6） |
| `epoch` | 测试所用 epoch（从 0 开始） |
| `epoch_seed` | epoch 种子；epoch 0 冻结值 = `keccak256("PTC/mainnet/genesis-seed-v1")` = `0x05bc07f7…cb557ba`（决议 #1） |
| `program_seed` | 程序层种子；P0 冻结值 = `keccak256("PTC/p0/program-seed-v1")` = `0xbce3f761…68da732a`（决议 #3；生产映射 = prev_hash） |
| `dataset_profile` | `full-v1`（canonical，DESIGN 附录 A 原值）/ `smoke-v1`（缩容档 256 MiB/2 MiB，仅 CI/开发，决议 #5） |
| `dataset` | 按所选 profile 生效的参数快照（EPOCH_BLOCKS / DATASET_BYTES_INIT / …） |
| `samples` | 样本数 |
| `generator` | `{tool, seed, algorithm}`——重生成输入集所需的全部信息 |

### inputs.jsonl（每行一个样本）

```json
{"i":0,"header":"0x<256 hex = 128B>","nonce":"0x<16 hex = 8B>"}
```

### golden.jsonl / 平台输出（每行一个结果）

```json
{"i":0,"mix":"0x<64 hex = 32B>","hash":"0x<64 hex = 32B>"}
```

### 哈希输入构造（对齐 DESIGN §6.1 伪代码）

```
seed_state = HASH_INIT(header_128B || nonce_8B)      # keccak512
```

- header 中 **`mix_digest` 字段（offset 86，32B）必须为零** —— 生成器已强制，runner 应断言；
- header 中 `reserved`（offset 118，10B）必须为零；
- header 内 nonce 字段（offset 76）与行内 `nonce` 同值（生成器已同步，见 TODO#4）；
- `program` 由 `manifest.program_seed` 按 §6.4 派生；
- dataset 由 runner 从 `epoch_seed` + `dataset` 参数**确定性重建**（Argon2id 风格 KDF，§6.2），禁止从网络或未校验的缓存获取。

## 决议记录（原未决项，2026-08-29 ref 第一步落地时清零）

| # | 事项 | 决议 |
|---|---|---|
| TODO#1 | `epoch_seed(0)` 常量值 | `keccak256("PTC/mainnet/genesis-seed-v1")` = `0x05bc07f76525e02d921bdf17412b8037dbd3f4324e02cf2ad2f03fa68cb557ba`（ref Go keccak 实现，原版 0x01 padding；已知答案向量 + 冻结值回归钉双验证；同步钉入 DESIGN 附录 A） |
| TODO#2 | `algo_version` 字节序 | **小端 LE**（与 header 其余整数字段的 Bitcoin 序列化惯例一致；生成器已按 LE，冻结） |
| TODO#3 | `program_seed` 派生源 | 生产映射 = **prev_hash**（§6.4）；向量集代入冻结常量 `keccak256("PTC/p0/program-seed-v1")` = `0xbce3f7616c64866be6509888369e5bad7e655e05efa10f154e6b691a68da732a`。每个向量集一个 program；多变体 program 留作后续向量集 |
| TODO#4 | header 内 nonce 双现 | 冻结现状：hash 输入 = 128B header ‖ 8B nonce，header[76:84] 与附加 nonce 同值（nonce 在种子材料中出现两次），对齐 §6.1 字面 |
| TODO#5 | 缩容 dataset profile | 引入 `dataset_profile`：`smoke-v1` = 256 MiB dataset / 2 MiB cache（比例保持 /128，其余参数与代码路径完全同 full-v1，仅 N_items 缩小；**非 canonical，仅 CI/开发**）；冻结验收唯一有效档仍为 `full-v1`（6 GiB） |

## 冻结清单（golden/FREEZE.md）

1. [x] TODO#1–#5 全部清零（2026-08-29，决议记录见上）；golden 生成时复核 manifest 无 null 字段
2. [ ] ref 生成 smoke + full golden；smoke 进 git，full 进 release artifact 并记录 SHA256
3. [ ] 五平台各跑满 full profile，`verify.py` 全绿
4. [ ] 平台/环境矩阵（GPU 型号、驱动、编译器版本）记入 FREEZE.md
5. [ ] DESIGN.md / 白皮书双语文本锁定，git tag，白皮书 v1.0 发布

## 与规范的关系

数值参数以 **DESIGN.md 附录 A** 为准（本目录 manifest 只是快照）。冲突时 DESIGN 赢，改这里。
