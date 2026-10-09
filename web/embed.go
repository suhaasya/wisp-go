// Package web embeds the Vite production build for the Wisp UI.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// DistFS is the production frontend build rooted at dist/.
func DistFS() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		return dist
	}
	return sub
}
