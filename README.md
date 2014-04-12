SqServer
========

Very simple Go HTTP handler that:

* Allows adding Handlers for paths (using ServeMux)
* If there is no matching handler then serves up an asset file if present
* Logs requests

Asset Serving
=============

The assets directory must contain two subdirectories static/ and templates/.

* static/ - contains all static files - images, javascript, css etc.
* templates/ - contains all golang templates

Example Organization:

    static/
      favicon.ico
      stylesheets/
        somefile.css
      javascripts/
        somefile.js
    templates/
      index.html
      test/
        index.html
        whatever.html

All templates are pre-processed and available under their pathname. Example:

    server.HandleFunc("/test/whatever", func(w http.ResponseWriter, r *http.Request) {
      server.ServeTemplate(w, r, "test/whatever.html")
    }

Example Usage
=============

    import(
      ...
      "github.com/squaremill/sqserver"
    )

    func main() {
      ...

      server := sqserver.NewServer("/path/to/assets")

      server.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
      }

      // Special handler for the "/" root path
      server.HandleRootFunc(func(w http.ResponseWriter, r *http.Request) {
        server.ServeTemplate(w, r, "index.html")
      }

      err := http.ListenAndServe(*httpAddr, server)

      ...
    }
