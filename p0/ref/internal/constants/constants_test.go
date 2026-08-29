package constants

import (
	"encoding/hex"
	"testing"
)

// 冻结值回归钉：2026-08-29 由 ref keccak 首次计算并写入文档
// （p0/README.md 决议表、DESIGN.md 附录 A）。若本测试失败，说明
// keccak 实现或常量原像被改动——P0 冻结后这是不允许的破坏性变更。
func TestFrozenConstants(t *testing.T) {
	epoch := GenesisEpochSeedV0()
	got := hex.EncodeToString(epoch[:])
	want := "05bc07f76525e02d921bdf17412b8037dbd3f4324e02cf2ad2f03fa68cb557ba"
	if got != want {
		t.Fatalf("epoch_seed(0) = %s, want %s", got, want)
	}

	prog := ProgramSeedP0()
	got = hex.EncodeToString(prog[:])
	want = "bce3f7616c64866be6509888369e5bad7e655e05efa10f154e6b691a68da732a"
	if got != want {
		t.Fatalf("program_seed = %s, want %s", got, want)
	}
}
