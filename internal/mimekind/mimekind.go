// Package mimekind classifies files so the server and dashboard can decide
// how to present them (render vs. download, icon choice, preview eligibility).
package mimekind

import (
	"mime"
	"path/filepath"
	"strings"
)

type Category string

const (
	CategoryHTML  Category = "html"
	CategoryPDF   Category = "pdf"
	CategoryImage Category = "image"
	CategoryAudio Category = "audio"
	CategoryVideo Category = "video"
	CategoryText  Category = "text"
	CategoryDir   Category = "directory"
	CategoryOther Category = "other"
)

var textExt = map[string]bool{
	".txt": true, ".md": true, ".markdown": true, ".json": true, ".xml": true,
	".csv": true, ".tsv": true, ".yml": true, ".yaml": true, ".toml": true,
	".ini": true, ".cfg": true, ".conf": true, ".log": true,
	".go": true, ".js": true, ".jsx": true, ".ts": true, ".tsx": true,
	".py": true, ".rb": true, ".rs": true, ".java": true, ".c": true,
	".h": true, ".cpp": true, ".css": true, ".scss": true, ".sh": true,
	".sql": true, ".env.example": true,
}

var imageExt = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".svg": true, ".avif": true, ".ico": true, ".bmp": true,
}

var audioExt = map[string]bool{".mp3": true, ".wav": true, ".ogg": true, ".m4a": true, ".flac": true}
var videoExt = map[string]bool{".mp4": true, ".webm": true, ".mov": true, ".m4v": true}

// Detect returns the MIME type string and a display category for name.
func Detect(name string) (mimeType string, category Category) {
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case ext == ".html" || ext == ".htm":
		return "text/html; charset=utf-8", CategoryHTML
	case ext == ".pdf":
		return "application/pdf", CategoryPDF
	case imageExt[ext]:
		return imageMime(ext), CategoryImage
	case audioExt[ext]:
		return audioMime(ext), CategoryAudio
	case videoExt[ext]:
		return videoMime(ext), CategoryVideo
	case textExt[ext]:
		return "text/plain; charset=utf-8", CategoryText
	}
	if t := mime.TypeByExtension(ext); t != "" {
		return t, CategoryOther
	}
	return "application/octet-stream", CategoryOther
}

func imageMime(ext string) string {
	switch ext {
	case ".svg":
		return "image/svg+xml"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "image/" + strings.TrimPrefix(ext, ".")
	}
}

func audioMime(ext string) string {
	if ext == ".mp3" {
		return "audio/mpeg"
	}
	return "audio/" + strings.TrimPrefix(ext, ".")
}

func videoMime(ext string) string {
	if ext == ".mov" {
		return "video/quicktime"
	}
	return "video/" + strings.TrimPrefix(ext, ".")
}

// IsPreviewableText reports whether files of this category should be
// rendered as text in the file browser's inline preview.
func IsPreviewableText(c Category) bool { return c == CategoryText }
