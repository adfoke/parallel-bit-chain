# geyser-hip（AMD）

**状态：未实现** —— 本目录暂无代码，只登记平台要点。契约见 `../README.md`。

## 计划要点

- 工具链：hipcc（clang 系）；gfx 架构矩阵（gfx1100 / gfx1151…）记入 FREEZE.md
- **FP 纪律**：禁用 `-ffast-math` 及任何 relax-fast-math 变体；默认 precise 模式显式确认
- wavefront（wave64）与 LANES=32 的映射：一 wave 跑两组 lane 或半 wave，实现时定，但语义等价必须由向量集背书
- keccak512/256 原版 padding，勿用硬件 SHA 指令替代（那是 SHA3）
- **TODO**：main.hip、Makefile（实现时补）
