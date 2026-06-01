package ocr

import (
	"bytes"
	"fmt"
	"os/exec"
)

// DDddocr 基于 Python ddddocr 的验证码识别器
type DDddocr struct {
	pythonCmd string
}

// NewDDddocr 创建 ddddocr 识别器
func NewDDddocr() *DDddocr {
	return &DDddocr{pythonCmd: "python"}
}

// NewDDddocrWithPython 指定 Python 解释器路径
func NewDDddocrWithPython(python string) *DDddocr {
	return &DDddocr{pythonCmd: python}
}

// Recognize 识别验证码图片，返回识别结果
func (d *DDddocr) Recognize(image []byte) (string, error) {
	script := `
import sys
import ddddocr

ocr = ddddocr.DdddOcr()
result = ocr.classification(sys.stdin.buffer.read())
print(result)
`
	cmd := exec.Command(d.pythonCmd, "-c", script)
	cmd.Stdin = bytes.NewReader(image)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ddddocr run failed: %w, stderr: %s", err, stderr.String())
	}

	result := bytes.TrimSpace(stdout.Bytes())
	return string(result), nil
}
