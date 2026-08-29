# geyser-cuda（NVIDIA）

**状态：未实现** —— 本目录暂无代码，只登记平台要点。契约见 `../README.md`。

## 计划要点

- 工具链：nvcc；架构矩阵（sm_80 / sm_89 / sm_90 / sm_120…）跑哪几个记入 FREEZE.md
- **FP 纪律**：禁用 `-use_fast_math`；显式 `-ftz=false -prec-div=true -prec-sqrt=true`
- LANES=32 正好半个 warp：一轮内 32 lane 可用单个 warp 协作（shuffle 折叠），跨轮串行
- keccak512/256 需要 in-kernel 实现或 per-item host 调用（P0 正确性运行无所谓性能，先对、后快）
- dataset 6 GiB 常驻显存；epoch 切换的重建走 cache（48 MiB）在设备端展开
- **TODO**：main.cu、Makefile（实现时补）
