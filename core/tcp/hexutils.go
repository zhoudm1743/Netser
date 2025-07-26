package tcp

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// HexUtils 提供十六进制数据处理工具
type HexUtils struct{}

// BytesToHex 将字节数组转换为十六进制字符串
func BytesToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// HexToBytes 将十六进制字符串转换为字节数组
func HexToBytes(hexStr string) ([]byte, error) {
	// 去除可能的空格和特殊字符
	cleanHex := cleanHexString(hexStr)

	// 解析十六进制
	return hex.DecodeString(cleanHex)
}

// FormatHexString 格式化十六进制字符串（添加空格分隔）
func FormatHexString(hexStr string, bytesPerGroup int) string {
	if bytesPerGroup <= 0 {
		bytesPerGroup = 1
	}

	// 清理输入
	cleanHex := cleanHexString(hexStr)
	if cleanHex == "" {
		return ""
	}

	// 补全奇数长度
	if len(cleanHex)%2 != 0 {
		cleanHex = "0" + cleanHex
	}

	// 计算分组长度
	groupSize := bytesPerGroup * 2

	// 分组输出
	var result strings.Builder
	for i := 0; i < len(cleanHex); i += groupSize {
		end := i + groupSize
		if end > len(cleanHex) {
			end = len(cleanHex)
		}

		if i > 0 {
			result.WriteByte(' ')
		}
		result.WriteString(cleanHex[i:end])
	}

	return result.String()
}

// ParseHexWithWildcards 解析带通配符的十六进制字符串
// 例如: "12 ?? 34"，其中??为通配符，可以匹配任何字节
func ParseHexWithWildcards(hexStr string) ([]byte, []bool) {
	// 清理和标准化输入
	hexStr = strings.ReplaceAll(hexStr, " ", "")

	// 提取所有的十六进制对和通配符
	hexPairs := make([]string, 0)
	wildcards := make([]bool, 0)

	for i := 0; i < len(hexStr); i += 2 {
		if i+1 >= len(hexStr) {
			// 处理奇数长度
			pair := hexStr[i:] + "0"
			hexPairs = append(hexPairs, pair)
			wildcards = append(wildcards, false)
			continue
		}

		pair := hexStr[i : i+2]
		if pair == "??" {
			hexPairs = append(hexPairs, "00") // 占位
			wildcards = append(wildcards, true)
		} else {
			hexPairs = append(hexPairs, pair)
			wildcards = append(wildcards, false)
		}
	}

	// 转换为字节数组
	hexData := strings.Join(hexPairs, "")
	data, _ := hex.DecodeString(hexData)

	return data, wildcards
}

// HexDump 生成十六进制转储输出（类似于xxd工具）
func HexDump(data []byte, bytesPerLine int) string {
	if len(data) == 0 {
		return "(empty)"
	}

	if bytesPerLine <= 0 {
		bytesPerLine = 16
	}

	var result strings.Builder

	for i := 0; i < len(data); i += bytesPerLine {
		// 地址部分
		result.WriteString(fmt.Sprintf("%08x: ", i))

		// 十六进制部分
		end := i + bytesPerLine
		if end > len(data) {
			end = len(data)
		}

		// 十六进制显示
		for j := i; j < i+bytesPerLine; j++ {
			if j < end {
				result.WriteString(fmt.Sprintf("%02x ", data[j]))
			} else {
				result.WriteString("   ")
			}

			// 中间空格
			if j == i+bytesPerLine/2-1 {
				result.WriteByte(' ')
			}
		}

		// ASCII部分
		result.WriteString(" |")
		for j := i; j < end; j++ {
			b := data[j]
			if b >= 32 && b <= 126 { // 可打印ASCII字符
				result.WriteByte(b)
			} else {
				result.WriteByte('.')
			}
		}
		result.WriteString("|\n")
	}

	return result.String()
}

// CompareHex 比较两个十六进制字符串（忽略格式）
func CompareHex(hex1, hex2 string) bool {
	b1, err1 := HexToBytes(hex1)
	b2, err2 := HexToBytes(hex2)

	if err1 != nil || err2 != nil {
		return false
	}

	// 比较长度和内容
	if len(b1) != len(b2) {
		return false
	}

	for i := 0; i < len(b1); i++ {
		if b1[i] != b2[i] {
			return false
		}
	}

	return true
}

// HexStringToBinary 将十六进制字符串转换为二进制字符串
func HexStringToBinary(hexStr string) (string, error) {
	data, err := HexToBytes(hexStr)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	for _, b := range data {
		result.WriteString(fmt.Sprintf("%08b ", b))
	}

	return strings.TrimSpace(result.String()), nil
}

// FindHexPattern 在数据中查找十六进制模式
func FindHexPattern(data []byte, pattern string) []int {
	patternBytes, wildcards := ParseHexWithWildcards(pattern)
	indices := make([]int, 0)

	if len(patternBytes) == 0 {
		return indices
	}

	// 搜索模式
	for i := 0; i <= len(data)-len(patternBytes); i++ {
		match := true

		for j := 0; j < len(patternBytes); j++ {
			if !wildcards[j] && data[i+j] != patternBytes[j] {
				match = false
				break
			}
		}

		if match {
			indices = append(indices, i)
		}
	}

	return indices
}

// cleanHexString 清理十六进制字符串
func cleanHexString(hexStr string) string {
	// 去除空格、制表符、换行符等
	hexStr = strings.ReplaceAll(hexStr, " ", "")
	hexStr = strings.ReplaceAll(hexStr, "\t", "")
	hexStr = strings.ReplaceAll(hexStr, "\n", "")
	hexStr = strings.ReplaceAll(hexStr, "\r", "")

	// 去除0x/0X前缀
	if strings.HasPrefix(hexStr, "0x") || strings.HasPrefix(hexStr, "0X") {
		hexStr = hexStr[2:]
	}

	// 去除非十六进制字符
	re := regexp.MustCompile("[^0-9A-Fa-f]")
	hexStr = re.ReplaceAllString(hexStr, "")

	return hexStr
}

// IsValidHexString 检查是否为有效的十六进制字符串
func IsValidHexString(hexStr string) bool {
	cleanHex := cleanHexString(hexStr)
	return len(cleanHex) > 0
}

// CombineHexStrings 合并多个十六进制字符串
func CombineHexStrings(hexStrings ...string) (string, error) {
	var combinedBytes []byte

	for _, hexStr := range hexStrings {
		bytes, err := HexToBytes(hexStr)
		if err != nil {
			return "", fmt.Errorf("invalid hex string: %s", hexStr)
		}
		combinedBytes = append(combinedBytes, bytes...)
	}

	return BytesToHex(combinedBytes), nil
}
