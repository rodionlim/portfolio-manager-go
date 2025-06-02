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

//go:build builtinassets

package ui

import (
	"fmt"
	"io"
	"net/http"
	"portfolio-manager/web/assets"
	"portfolio-manager/web/server"
	"strings"
)

// AssetsHandler returns a handler that serves the embedded static files.
func AssetsHandler() http.Handler {

	reactAssetsRoot := "/static"
	assetsFS := http.FS(assets.New(EmbedFS))
	fileServer := server.StaticFileServer(assetsFS)

	serveReactApp := func(w http.ResponseWriter, r *http.Request) {
		route := r.URL.Path

		// Skip API routes - these should be handled by other handlers
		if strings.HasPrefix(route, "/api/") || strings.HasPrefix(route, "/swagger/") || route == "/healthz" {
			http.NotFound(w, r)
			return
		}

		// If the route is "/" serve the React index.
		if route == "/" {
			r.URL.Path = reactAssetsRoot + "/index.html"
			f, err := assetsFS.Open(r.URL.Path)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error opening file %s: %v", r.URL.Path, err)
				return
			}
			defer f.Close()

			data, err := io.ReadAll(f)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error reading file %s: %v", r.URL.Path, err)
				return
			}
			w.Write(data)
		} else {
			// Try to serve as a static asset first
			r.URL.Path = reactAssetsRoot + route

			// Check if the file exists
			f, err := assetsFS.Open(r.URL.Path)
			if err == nil {
				// File exists, serve it as a static asset
				f.Close()
				fileServer.ServeHTTP(w, r)
			} else {
				// File doesn't exist, serve index.html for SPA routing
				r.URL.Path = reactAssetsRoot + "/index.html"
				f, err := assetsFS.Open(r.URL.Path)

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Error opening index.html: %v", err)
					return
				}
				defer f.Close()

				data, err := io.ReadAll(f)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "Error reading index.html: %v", err)
					return
				}
				w.Write(data)
			}
		}
	}

	return http.HandlerFunc(serveReactApp)
}
