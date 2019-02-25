package main

import (
    "html/template"
    "net/http"
    "os"
    "path/filepath"
)

var staticAssetsDir = os.Getenv("STATIC_ASSETS_DIR")
var templatesDir = os.Getenv("TEMPLATES_DIR")
var tlsCertPath = os.Getenv("TLS_CERT_PATH")
var tlsKeyPath = os.Getenv("TLS_KEY_PATH")

// neuteredFileSystem is used to prevent directory listing of static assets
type neuteredFileSystem struct {
    fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
    // Check if path exists
    f, err := nfs.fs.Open(path)
    if err != nil {
        return nil, err
    }

    // If path exists, check if is a file or a directory.
    // If is a directory, stop here with an error saying that file
    // does not exist. So user will get a 404 error code for a file/directory
    // that does not exist, and for directories that exist.
    s, err := f.Stat()
    if err != nil {
        return nil, err
    }
    if s.IsDir() {
        return nil, os.ErrNotExist
    }

    // If file exists and the path is not a directory, let's return the file
    return f, nil
}

// loveMountains renders love-mountains page after passing some data to the HTML template
func loveMountains(w http.ResponseWriter, r *http.Request) {
    // Load template from disk
    tmpl := template.Must(template.ParseFiles("love-mountains.html"))
    // Inject data into template
    data := "Any dynamic data"
    tmpl.Execute(w, data)
}

// httpsRedirect redirects http requests to https
func httpsRedirect(w http.ResponseWriter, r *http.Request) {
    http.Redirect(
        w, r,
        "https://"+r.Host+r.URL.String(),
        http.StatusMovedPermanently,
    )
}

func main() {
    // http to https redirection
    go http.ListenAndServe(":80", http.HandlerFunc(httpsRedirect))

    // Serve static files while preventing directory listing
    mux := http.NewServeMux()
    fs := http.FileServer(neuteredFileSystem{http.Dir(staticAssetsDir)})
    mux.Handle("/", fs)

    // Serve one page site dynamic pages
    mux.HandleFunc("/love-mountains", loveMountains)

    // Launch TLS server
    log.Fatal(http.ListenAndServeTLS(":443", tlsCertPath, tlsKeyPath, mux))
}