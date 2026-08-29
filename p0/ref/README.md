# 参考实现（ref，Go）

**状态：第 1/4 步已落地（2026-08-29）** —— keccak（原版 0x01 padding，已知答案 + 冻结值回归钉）+ 冻结常量（epoch_seed / program_seed，五 TODO 清零）；`cmd/geyser` 可执行（epoch-seed / program-seed / keccak256 / keccak512）。dataset/program/lane core/golden 生成未实现。

## 职责

1. **实现 Geyser v1 全量**：cache/dataset 生成（Argon2id 风格 KDF，§6.2）、程序层（§6.4）、32-lane 执行核（§6.1）、keccak512/256（原版 padding）
2. **输出冻结常量**：计算 `epoch_seed(0) = keccak256("PTC/mainnet/genesis-seed-v1")` 并回填 manifest（清 TODO#1）；确定 program_seed 回填值（清 TODO#3）
3. **生成 golden**：smoke（进 git）+ full（≥10^6，release artifact + SHA256）
4. **交叉验证方**：与五平台互为独立对照

## API 草案（实现落地时对齐，允许演化）

```go
package geyser

type Config struct { /* 附录 A 参数 */ }

// 单次哈希：P0 与全节点共用此入口
func Hash(header [128]byte, nonce [8]byte, prog *Program, ds *Dataset) (mix, hash [32]byte)

// 确定性构造
func GenerateCache(epochSeed []byte, cfg Config) (*Cache, error)
func GenerateDataset(c *Cache, cfg Config) (*Dataset, error)
func GenerateProgram(seed []byte, cfg Config) (*Program, error)

// 冻结常量
func GenesisEpochSeed() []byte // keccak256("PTC/mainnet/genesis-seed-v1")
```

## 独立性原则（P0 签核前必须决议）

- cpu 平台与 ref **可以共享**：dataset/program 的确定性构造代码（这部分错了五方一起错，共享无害且省力）；
- cpu 平台与 ref **不应共享**：32-lane 执行核（lane 寄存器、FNV、f32 程序执行）——这正是要跨实现验证的部分；
- 争议时原则：**被验证的代码路径必须存在至少两个独立实现**。

## 落地顺序建议

1. keccak + 常量回填（解锁 golden 前置）
2. cache/dataset 生成 + program 生成
3. lane 执行核 → smoke golden
4. 性能版 dataset 生成（full profile 10^6 样本要过夜的话优化点全在这）
