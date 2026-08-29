# P0 测试向量框架 —— 骨架目标
# 详见 p0/README.md（工作流 / 向量协议 / 未决项）

PY      ?= python3
COUNT   ?= 1024
SEED    ?= p0-smoke-1
VECDIR  ?= vectors/smoke
OUT     ?= /tmp/geyser-out.jsonl

.PHONY: inputs golden verify selftest smoke run-cuda run-hip run-oneapi run-metal run-cpu

# 1) 确定性输入集（已可用）
inputs:
	$(PY) harness/gen_inputs.py --count $(COUNT) --seed $(SEED) --out-dir $(VECDIR)

# 2) golden：参考实现（ref/，Go）就绪后解锁
golden:
	@echo "ERROR: 参考实现未就绪 —— 见 p0/ref/README.md（Go，P1 前置）"; exit 1

# 4) 位级比对
verify:
	$(PY) harness/verify.py --golden $(VECDIR)/golden.jsonl --output $(OUT)

# 比对器自检（正例 + 两类负例）
selftest:
	$(PY) harness/verify.py --self-test

# 3) 平台 runner：实现后解锁（契约见 p0/platforms/README.md）
run-cuda run-hip run-oneapi run-metal run-cpu:
	@echo "ERROR: 平台 runner 未就绪 —— 见 p0/platforms/$(@:run-%=%)/README.md"; exit 1

# 冒烟：骨架自身可用性（不涉及 Geyser 实现）
smoke: selftest inputs
	@echo "[ok] P0 骨架冒烟通过（比对器自检 + 向量生成）"
