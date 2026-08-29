// Package keccak implements the original Keccak sponge construction.
//
// P0 一致性红线：Geyser 使用**原版 Keccak**（域分隔字节 0x01），
// 不是 NIST SHA-3（域分隔字节 0x06）。两者对同一输入产生不同摘要，
// 任何把本包替换为 SHA-3 标准库的实现都会位级偏离 golden。
// 参考：DESIGN.md §6.1（HASH_INIT = keccak512，HASH_FINAL = keccak256）。
package keccak

import (
	"encoding/binary"
	"math/bits"
)

const (
	rate256 = 136 // 1088 bit —— Keccak-256 吞吐率
	rate512 = 72  // 576 bit —— Keccak-512 吞吐率
)

// rc 是 Keccak-f[1600] 的 24 个轮常数。
var rc = [24]uint64{
	0x0000000000000001, 0x0000000000008082, 0x800000000000808a, 0x8000000080008000,
	0x000000000000808b, 0x0000000080000001, 0x8000000080008081, 0x8000000000008009,
	0x000000000000008a, 0x0000000000000088, 0x0000000080008009, 0x000000008000000a,
	0x000000008000808b, 0x800000000000008b, 0x8000000000008089, 0x8000000000008003,
	0x8000000000008002, 0x8000000000000080, 0x000000000000800a, 0x800000008000000a,
	0x8000000080008081, 0x8000000000008080, 0x0000000080000001, 0x8000000080008008,
}

// rotc[x+5*y] 是 lane (x,y) 的 ρ 步旋转量。
var rotc = [25]int{
	0, 1, 62, 28, 27, // y=0
	36, 44, 6, 55, 20, // y=1
	3, 10, 43, 25, 39, // y=2
	41, 45, 15, 21, 8, // y=3
	18, 2, 61, 56, 14, // y=4
}

// keccakF1600 就地执行 Keccak-f[1600] 置换（24 轮 θ/ρπ/χ/ι）。
// 状态布局：a[x+5*y]，lane 以小端序映射字节。
func keccakF1600(a *[25]uint64) {
	var b [25]uint64
	var c, d [5]uint64
	for round := range 24 {
		// θ
		for x := range 5 {
			c[x] = a[x] ^ a[x+5] ^ a[x+10] ^ a[x+15] ^ a[x+20]
		}
		for x := range 5 {
			d[x] = c[(x+4)%5] ^ bits.RotateLeft64(c[(x+1)%5], 1)
		}
		for x := range 5 {
			for y := range 5 {
				a[x+5*y] ^= d[x]
			}
		}
		// ρ + π
		for x := range 5 {
			for y := range 5 {
				b[y+5*((2*x+3*y)%5)] = bits.RotateLeft64(a[x+5*y], rotc[x+5*y])
			}
		}
		// χ
		for x := range 5 {
			for y := range 5 {
				a[x+5*y] = b[x+5*y] ^ (^b[(x+1)%5+5*y] & b[(x+2)%5+5*y])
			}
		}
		// ι
		a[0] ^= rc[round]
	}
}

// xorIn 把一个 rate 大小的块按小端 lane 异或进状态。
func xorIn(a *[25]uint64, block []byte) {
	for i := range len(block) / 8 {
		a[i] ^= binary.LittleEndian.Uint64(block[i*8:])
	}
}

// squeeze 从状态读出前 n 字节（n ≤ rate，256/512 输出均无需多轮挤压）。
func squeeze(a *[25]uint64, n int) []byte {
	out := make([]byte, n)
	for i := range n / 8 {
		binary.LittleEndian.PutUint64(out[i*8:], a[i])
	}
	return out
}

// sum 以 rate 为吞吐率、输出 n 字节摘要的海绵函数；
// pad10*1 填充使用原版 Keccak 域分隔字节 0x01（见包注释红线）。
func sum(data []byte, rate, n int) []byte {
	var a [25]uint64
	for len(data) >= rate {
		xorIn(&a, data[:rate])
		keccakF1600(&a)
		data = data[rate:]
	}
	block := make([]byte, rate)
	copy(block, data)
	block[len(data)] ^= 0x01 // 原版 Keccak 域分隔（SHA-3 为 0x06）
	block[rate-1] ^= 0x80    // pad10*1 收尾
	xorIn(&a, block)
	keccakF1600(&a)
	return squeeze(&a, n)
}

// Sum256 返回 data 的原版 Keccak-256 摘要（Geyser HASH_FINAL）。
func Sum256(data []byte) [32]byte {
	var out [32]byte
	copy(out[:], sum(data, rate256, 32))
	return out
}

// Sum512 返回 data 的原版 Keccak-512 摘要（Geyser HASH_INIT）。
func Sum512(data []byte) [64]byte {
	var out [64]byte
	copy(out[:], sum(data, rate512, 64))
	return out
}
