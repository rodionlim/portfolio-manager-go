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
)

// AssetsHandler returns a handler that serves the embedded static files.
func AssetsHandler() http.Handler {

	reactAssetsRoot := "/static"
	assetsFS := http.FS(assets.New(EmbedFS))
	fileServer := http.FileServer(assetsFS)

	serveReactApp := func(w http.ResponseWriter, r *http.Request) {
		route := r.URL.Path

		// Check specific asset routes and rewrite the URL.Path.
		for _, p := range []string{"/favicon.svg", "/favicon.ico", "/manifest.json"} {
			if route == p {
				// Prepend the reactAssetsRoot to get the correct asset location.
				r.URL.Path = reactAssetsRoot + p
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// If the route is "/" serve the React index.
		if route == "/" {
			r.URL.Path = reactAssetsRoot + "/index.html"
		} else {
			// For any other route, assume it's a static asset under reactAssetsRoot.
			r.URL.Path = reactAssetsRoot + route
		}

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

		// Optionally, add caching headers or content type based on file extension.
		w.Header().Set("Content-Type", http.DetectContentType(data))
		// Serve the file content
		w.Write(data)
	}

	return http.HandlerFunc(serveReactApp)
}
