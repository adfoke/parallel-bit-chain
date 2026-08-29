# geyser-oneapi（Intel）

**状态：未实现** —— 本目录暂无代码，只登记平台要点。契约见 `../README.md`。

## 计划要点

- 工具链：DPC++ / SYCL（icpx -fsycl）；设备矩阵（Arc A/B 系列、iGPU）记入 FREEZE.md
- **FP 纪律**：`-ffp-model=precise`；禁 `-ffast-math` / `-ffp-contract=fast`（FMA 收缩会改变位级结果，需显式关掉）
- sub_group 大小与 LANES=32 的映射在实现时定；语义等价由向量集背书
- keccak512/256 原版 padding
- **TODO**：main.cpp、Makefile（实现时补）
