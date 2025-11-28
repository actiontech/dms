package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// IsGzip 检查数据是否是 Gzip 压缩格式
func IsGzip(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

// DecodeGzipBytes 解码 Gzip 压缩的数据，返回解码后的字节数组
func DecodeGzipBytes(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decoded, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read gzip data: %v", err)
	}
	return decoded, nil
}
