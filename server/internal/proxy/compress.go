package proxy

import (
	"bytes"
	"fmt"
	"io"

	"github.com/klauspost/compress/zlib"
)

// Compressor 数据压缩器
type Compressor struct {
	level int
}

// NewCompressor 创建压缩器
func NewCompressor(level int) *Compressor {
	if level < 1 || level > 9 {
		level = 6
	}
	return &Compressor{level: level}
}

// Compress 压缩数据
func (c *Compressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := zlib.NewWriterLevel(&buf, c.level)
	if err != nil {
		return nil, fmt.Errorf("创建压缩器失败: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, fmt.Errorf("压缩数据失败: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("关闭压缩器失败: %w", err)
	}

	return buf.Bytes(), nil
}

// Decompress 解压数据
func (c *Compressor) Decompress(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("创建解压器失败: %w", err)
	}
	defer r.Close()

	result, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("解压数据失败: %w", err)
	}

	return result, nil
}
