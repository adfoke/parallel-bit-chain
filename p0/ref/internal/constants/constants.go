// Package constants 持有 P0 已冻结的协议常量（p0/README.md 决议记录）。
package constants

import "github.com/adfoke/parallel-bit-chain/p0/ref/internal/keccak"

// GenesisEpochSeedInput 是 epoch-0 数据集种子的原像（DESIGN §6.2 冻结）。
const GenesisEpochSeedInput = "PTC/mainnet/genesis-seed-v1"

// ProgramSeedP0Input 是 P0 向量集程序层种子的原像（p0/README TODO#3 决议）。
// 生产环境中程序层种子 = prev_hash（DESIGN §6.4）；向量集以本常量代入该角色。
const ProgramSeedP0Input = "PTC/p0/program-seed-v1"

// GenesisEpochSeedV0 返回 epoch_seed(0) = keccak256(GenesisEpochSeedInput)。
func GenesisEpochSeedV0() [32]byte {
	return keccak.Sum256([]byte(GenesisEpochSeedInput))
}

// ProgramSeedP0 返回 P0 向量集的 program_seed = keccak256(ProgramSeedP0Input)。
func ProgramSeedP0() [32]byte {
	return keccak.Sum256([]byte(ProgramSeedP0Input))
}
