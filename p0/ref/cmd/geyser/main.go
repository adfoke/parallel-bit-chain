// Command geyser 是 PTC Geyser 参考实现工具链的入口。
//
// ref 路线图（p0/ref/README.md）第 1 步：keccak + 冻结常量。
// 后续步骤加入 cache/dataset 生成、程序层、32-lane 执行核与 golden 生成。
package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/adfoke/parallel-bit-chain/p0/ref/internal/constants"
	"github.com/adfoke/parallel-bit-chain/p0/ref/internal/keccak"
)

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "epoch-seed":
		seed := constants.GenesisEpochSeedV0()
		fmt.Println(hex.EncodeToString(seed[:]))
	case "program-seed":
		seed := constants.ProgramSeedP0()
		fmt.Println(hex.EncodeToString(seed[:]))
	case "keccak256", "keccak512":
		hashStdin(os.Args[1])
	default:
		usage()
		os.Exit(2)
	}
}

// hashStdin 对标准输入做原版 Keccak，摘要打印到标准输出（调试用）。
func hashStdin(alg string) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "geyser: 读 stdin: %v\n", err)
		os.Exit(1)
	}
	switch alg {
	case "keccak256":
		d := keccak.Sum256(data)
		fmt.Println(hex.EncodeToString(d[:]))
	default:
		d := keccak.Sum512(data)
		fmt.Println(hex.EncodeToString(d[:]))
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `用法:
  geyser epoch-seed    打印 epoch_seed(0)（32B hex）
  geyser program-seed  打印 P0 向量集 program_seed（32B hex）
  geyser keccak256     原版 Keccak-256(stdin)（调试）
  geyser keccak512     原版 Keccak-512(stdin)（调试）`)
}
