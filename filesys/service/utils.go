package service

import (
	"fmt"
	"strings"
)

// 生成唯一文件名
func generateUniqueName(original string, existsFunc func(string) bool) string {
	base, ext := splitFilename(original)

	for i := 1; ; i++ {
		newName := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if !existsFunc(newName) {
			return newName
		}
	}
}

// 分割文件名和扩展名
func splitFilename(filename string) (string, string) {
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 {
		return filename, ""
	}
	return filename[:lastDot], filename[lastDot:]
}
