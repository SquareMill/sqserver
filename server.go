// Copyright (c) 2014 Conor Hunt & SquareMill Labs All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqserver

import(
  "html/template"
  "io/ioutil"
  "net/http"
  "os"
  "log"
  "path"
  "path/filepath"
  "strings"
)

type reqHandler func(http.ResponseWriter, *http.Request)

type Server struct {
  appMux *http.ServeMux
  templates *template.Template
  baseDir string
  rootHandler reqHandler
}

// Create a new handler with an asset directory specified
func NewServer(baseDir string) *Server {
  handler := &Server{appMux: http.NewServeMux(), baseDir: baseDir}
  handler.ParseTemplates()
  return handler
}

// Assign a handler to a path
func (h *Server) HandleFunc(path string, handler reqHandler)  {
  if path == "/" {
    panic("Please use HandleRootFunc to handle the root / path")
  }
  h.appMux.HandleFunc(path, handler)
}

// Assign a handler to the root "/" path
func (h *Server) HandleRootFunc(handler reqHandler) {
  h.rootHandler = handler
}

// Parse all of the templates and make them available for use
func (h *Server) ParseTemplates() {
  templateDir := filepath.Join(h.baseDir, "templates")

  h.templates = template.New("not-existing")
  filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
    if !info.IsDir() {
      text, err := ioutil.ReadFile(path)
      if err != nil { return err }
      path = path[len(templateDir)+1:]
      h.templates.New(path).Parse(string(text))
    }
    return nil
  })
}

func (h *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
    path := cleanPath(r.URL.Path)

    // Special case for the root path. If we register "/" as a path with the
    // Mux it will handle all paths not otherwise registered. We do not want
    // this behaviour so we instead use a special rootHandler registered
    // with HandleRoot*
    if path == "/" {
      if h.rootHandler != nil {
        h.rootHandler(w,r)
      } else {
        http.NotFound(w,r)
      }
    } else {
      // Check to see if there is a handler registered for this path
      handler, pattern := h.appMux.Handler(r)
      if pattern == "" {
        // If the app is not handling this path then try a static file from assets
        h.serveStaticFile(w, r)
      } else {
        // If the app is handling this path then run the handler
        handler.ServeHTTP(w, r)
      }
    }
    return
}

func cleanPath(upath string) string {
  if !strings.HasPrefix(upath, "/") {
    upath = "/" + upath
  }
  return path.Clean(upath)
}

// Serve a static file from disk, but only if that file exists
func (h *Server) serveStaticFile(w http.ResponseWriter, r *http.Request) {
  cleanPath := cleanPath(r.URL.Path)
  name := filepath.Join(h.baseDir, "static", cleanPath)

  f, err := os.Open(name)
  if err != nil {
    http.NotFound(w, r)
    return
  }
  defer f.Close()

  d, err1 := f.Stat()
  if err1 != nil {
    http.NotFound(w, r)
    return
  }

  if d.IsDir() {
    http.NotFound(w, r)
    return
  }

  http.ServeContent(w, r, d.Name(), d.ModTime(), f)
}

// Serve up a template from a registered template in the assets/templates
// directory.
func (h *Server) ServeTemplate(w http.ResponseWriter, r *http.Request, name string) {
  h.templates.ExecuteTemplate(w, name, nil)
}
