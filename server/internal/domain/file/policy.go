package file

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	ErrEmptyExtension          = errors.New("文件扩展名不能为空")
	ErrDangerousExtension      = errors.New("不允许上传该类型的文件")
	ErrUnsupportedExtension    = errors.New("不允许上传该扩展名的文件")
	ErrTypeDisabled            = errors.New("当前配置不允许上传该类型文件")
	ErrUnsupportedDeclaredMIME = errors.New("不允许上传该MIME类型的文件")
	ErrDeclaredMismatch        = errors.New("文件扩展名与声明类型不一致")
	ErrContentMismatch         = errors.New("文件内容与扩展名不一致")
)

var dangerousExtensions = map[string]bool{
	"php": true, "php3": true, "php4": true, "php5": true, "phtml": true,
	"exe": true, "msi": true, "bat": true, "cmd": true, "com": true,
	"sh": true, "bash": true, "zsh": true,
	"js": true, "vbs": true, "vbe": true, "wsf": true,
	"html": true, "htm": true, "svg": true,
	"jar": true, "war": true, "ear": true,
	"py": true, "pl": true, "rb": true,
	"asp": true, "aspx": true, "jsp": true,
}

var extensionMIMETypes = map[string]string{
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"webp": "image/webp",
	"pdf":  "application/pdf",
}

func Extension(originalName string) string {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(originalName)))
	if ext == "" {
		return ""
	}
	return strings.TrimPrefix(ext, ".")
}

func IsDangerousExtension(ext string) bool {
	return dangerousExtensions[strings.ToLower(strings.TrimSpace(ext))]
}

func ExpectedMIME(ext string) (string, bool) {
	mimeType, ok := extensionMIMETypes[strings.ToLower(strings.TrimSpace(ext))]
	return mimeType, ok
}

func ValidateUpload(originalName string, declaredMIME string, detectedMIME string, allowedTypes []string) error {
	ext := Extension(originalName)
	if ext == "" {
		return ErrEmptyExtension
	}
	if IsDangerousExtension(ext) {
		return ErrDangerousExtension
	}
	expectedMIME, ok := ExpectedMIME(ext)
	if !ok {
		return ErrUnsupportedExtension
	}

	allowed := make(map[string]bool, len(allowedTypes))
	for _, value := range allowedTypes {
		allowed[strings.TrimSpace(value)] = true
	}
	if !allowed[expectedMIME] {
		return ErrTypeDisabled
	}
	declaredMIME = strings.TrimSpace(declaredMIME)
	if declaredMIME != "" && !allowed[declaredMIME] {
		return ErrUnsupportedDeclaredMIME
	}
	if declaredMIME != "" && declaredMIME != expectedMIME {
		return ErrDeclaredMismatch
	}
	if strings.TrimSpace(detectedMIME) != expectedMIME {
		return ErrContentMismatch
	}
	return nil
}

func CanDelete(referenceCount int64) bool {
	return referenceCount <= 0
}

func IsSafeRelativeStoragePath(storagePath string) bool {
	cleanPath := filepath.Clean(strings.TrimSpace(storagePath))
	return cleanPath != "." &&
		!filepath.IsAbs(cleanPath) &&
		cleanPath != ".." &&
		!strings.HasPrefix(cleanPath, ".."+string(filepath.Separator))
}
