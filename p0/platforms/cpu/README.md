# geyser-cpu（ARM64）

**状态：未实现** —— 本目录暂无代码，只登记平台要点。契约见 `../README.md`。

## 计划要点

- 目标：ARM64（Apple Silicon / Cortex）；语言 C 或 Go（与 ref 同血统则注意独立性，见 `../../ref/README.md`）
- NEON 允许使用，但 f32 规则同红线：round-to-nearest-even、denormal 按 §6.4
- 定位：五方里的**位基线**（bit baseline）——最慢但语义最透明的实现，出分歧时以它为参照系定位
- 性能无关紧要（P0 只验正确性）；单线程跑 10^6 样本也行，能过夜就过夜
- **TODO**：入口源码（实现时补）
