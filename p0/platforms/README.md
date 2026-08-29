# 平台 runner 契约（五方统一）

每个平台交付一个可执行：**`geyser-<platform>`**（cuda / hip / oneapi / metal / cpu）。

## CLI

```
geyser-<platform> --manifest <manifest.json> --inputs <inputs.jsonl> --output <out.jsonl>
```

## 行为要求

1. 读取 manifest，从 `epoch_seed` + `dataset` 参数**确定性重建** cache 与 dataset（Argon2id 风格 KDF，DESIGN §6.2）——禁止从网络或未校验的缓存获取；
2. `program` 从 `manifest.program_seed` 按 DESIGN §6.4 生成；
3. 对每行输入计算 `mix`（32B）与 `hash`（32B），**按输入顺序**写 JSONL（格式见 schema/output.schema.json）；
4. `manifest.epoch_seed` 或 `program_seed` 为 null 时：立即报错退出（规范未冻结，无 golden 可比对）；
5. 任何内部错误 → 非零退出码 + stderr 说明；**禁止静默跳行、禁止截断输出**；
6. 本地缓存 dataset 时，必须以 `(epoch_seed, dataset 参数, algo_version)` 全量匹配为键，键不匹配即重建；
7. runner 应断言输入 header 的 `mix_digest`（offset 86，32B）与 `reserved` 为零（向量协议要求）。

## 确定性红线（详细规则以 DESIGN §6.4 为准）

- f32 运算：IEEE-754 round-to-nearest-even；denormal 处理按 §6.4 红线执行；
- 整数运算：u32 环绕语义，无未定义行为；
- **禁用一切 fast-math 类编译选项**（各平台 README 列出具体 flag）；
- keccak512 / keccak256 必须是原版 Keccak padding（**不是** NIST SHA3 的 0x06 padding）；
- 编译器版本、目标架构、驱动版本记入 FREEZE.md 平台矩阵。

## 验收

```
cd p0 && make run-<platform> && make verify OUT=<平台输出>
```

五方全绿（full profile，≥10^6 样本）= P0 位级一致达成。
