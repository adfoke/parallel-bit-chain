# golden（黄金向量）

## 规则

- golden 由 **ref（Go 参考实现）** 生成，源头唯一；
- **smoke** golden（~1k 样本）进 git：`../vectors/smoke/golden.jsonl`；
- **full** golden（≥10^6 样本）不进 git（体积），以 release artifact 发布，SHA256 记入 `FREEZE.md`；
- **一经冻结，不可修改**：任何 golden 变更 = `algo_version` 升版，或整轮 P0 重走；
- 冻结前 manifest 里 `epoch_seed` / `program_seed` 为 null，golden 不存在也不应存在。

## 冻结流程

五方全绿后，复制 `FREEZE.template.md` 为 `FREEZE.md` 填写，然后 git tag。
