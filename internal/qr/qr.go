// Package qr generates real QR codes (SVG for the dashboard, ANSI blocks
// for the terminal) that encode a share's actual URL.
package qr

import (
	"fmt"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

// SVG renders url as a scalable QR code, sized in CSS pixels.
func SVG(url string, size int) (string, error) {
	q, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return "", err
	}
	bitmap := q.Bitmap()
	n := len(bitmap)
	if n == 0 {
		return "", fmt.Errorf("qr: empty bitmap")
	}
	cell := float64(size) / float64(n)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" shape-rendering="crispEdges">`, size, size, size, size)
	b.WriteString(`<rect width="100%" height="100%" fill="#ffffff"/>`)
	for y, row := range bitmap {
		for x, dark := range row {
			if !dark {
				continue
			}
			fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="#000000"/>`,
				float64(x)*cell, float64(y)*cell, cell+0.5, cell+0.5)
		}
	}
	b.WriteString(`</svg>`)
	return b.String(), nil
}

// Terminal renders url as block-character QR art for CLI output.
func Terminal(url string) (string, error) {
	q, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return "", err
	}
	bitmap := q.Bitmap()
	// Pad with a quiet zone so terminal scanners can lock on.
	pad := 2
	w := len(bitmap[0]) + pad*2

	var b strings.Builder
	blank := strings.Repeat(" ", w)
	for i := 0; i < pad/2+1; i++ {
		b.WriteString(blank + "\n")
	}
	for y := 0; y < len(bitmap); y += 2 {
		b.WriteString(strings.Repeat(" ", pad))
		for x := 0; x < len(bitmap[y]); x++ {
			top := bitmap[y][x]
			bottom := false
			if y+1 < len(bitmap) {
				bottom = bitmap[y+1][x]
			}
			b.WriteString(blockChar(top, bottom))
		}
		b.WriteString(strings.Repeat(" ", pad) + "\n")
	}
	for i := 0; i < pad/2+1; i++ {
		b.WriteString(blank + "\n")
	}
	return b.String(), nil
}

func blockChar(top, bottom bool) string {
	switch {
	case top && bottom:
		return "█"
	case top && !bottom:
		return "▀"
	case !top && bottom:
		return "▄"
	default:
		return " "
	}
}
