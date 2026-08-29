# geyser-metal（Apple）

**状态：未实现** —— 本目录暂无代码，只登记平台要点。契约见 `../README.md`。

## 计划要点

- 工具链：Swift/Obj-C 外壳 + MSL kernel；Apple GPU family 矩阵（M 系列各代）记入 FREEZE.md
- **FP 纪律（本平台最大的坑）**：MTLCompileOptions `fastMathEnabled = false` 必须显式设置——Metal 某些路径默认开 fast-math，会直接改变 f32 位级结果
- Apple GPU 的 denormal/FTZ 行为差异正是 DESIGN §6.4 denormal 红线的由来，实现时逐条对表
- UMA 零拷贝：dataset 直接驻留统一内存（这也顺便验证 UMA 挖矿路径，DESIGN §7.5）
- **TODO**：main.swift / Geyser.metal（实现时补）
