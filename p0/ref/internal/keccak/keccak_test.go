package keccak

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// 已知答案向量：原版 Keccak（0x01 padding），非 NIST SHA-3。
// 空串与 "abc" 的 Keccak-256 值即 Ethereum 生态长期使用的空哈希与
// 经典测试值，属公开可交叉验证的常量。
func TestSum256(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"},
		{"abc", "abc", "4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Sum256([]byte(tt.in))
			got := hex.EncodeToString(d[:])
			if got != tt.want {
				t.Fatalf("Sum256(%q) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}

func TestSum512(t *testing.T) {
	// 空串 Keccak-512 的公开测试值
	want := "0eab42de4c3ceb9235fc91acffe746b29c29a8c366b7c60e4e67c466f36a4304" +
		"c00fa9caf9d87976ba469bcbe06713b435f091ef2769fb160cdab33d3670680e"
	d := Sum512(nil)
	got := hex.EncodeToString(d[:])
	if got != want {
		t.Fatalf("Sum512(\"\") = %s, want %s", got, want)
	}
}

// 多块吸收路径（输入 > rate512=72 与 > rate256=136）：
// 无公开记忆向量，验证确定性与块边界不崩溃。
func TestMultiBlockDeterminism(t *testing.T) {
	buf := bytes.Repeat([]byte{0xA5}, 300)
	a1, a2 := Sum256(buf), Sum256(buf)
	if a1 != a2 {
		t.Fatal("同输入两次摘要不同")
	}
	b1, b2 := Sum512(buf), Sum512(buf)
	if b1 != b2 {
		t.Fatal("同输入两次摘要不同（512）")
	}
	// 长度跨越块边界的相邻输入必须产生不同摘要
	if Sum256(buf[:135]) == Sum256(buf[:136]) {
		t.Fatal("135/136 字节输入摘要不应相同")
	}
	if Sum512(buf[:71]) == Sum512(buf[:72]) {
		t.Fatal("71/72 字节输入摘要不应相同（512）")
	}
}
