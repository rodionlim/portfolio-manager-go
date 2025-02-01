// Copyright (c) 2025 Rodion Lim
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package server

import (
	"net/http"
	"path/filepath"
)

var mimeTypes = map[string]string{
	".cjs":   "application/javascript",
	".css":   "text/css",
	".eot":   "font/eot",
	".gif":   "image/gif",
	".ico":   "image/x-icon",
	".jpg":   "image/jpeg",
	".js":    "application/javascript",
	".json":  "application/json",
	".less":  "text/plain",
	".map":   "application/json",
	".otf":   "font/otf",
	".png":   "image/png",
	".svg":   "image/svg+xml",
	".ttf":   "font/ttf",
	".txt":   "text/plain",
	".woff":  "font/woff",
	".woff2": "font/woff2",
}

func StaticFileServer(root http.FileSystem) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fileExt := filepath.Ext(r.URL.Path)

			if t, ok := mimeTypes[fileExt]; ok {
				w.Header().Set("Content-Type", t)
			}

			http.FileServer(root).ServeHTTP(w, r)
		},
	)
}
