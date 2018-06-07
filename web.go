package carpe

import (
	"log"
	"net/http"
	"path/filepath"
)

func StartWeb(bind string, spool string) {
	serve := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		path = filepath.Join(spool, filepath.Clean(path))
		log.Printf("Serving web: %q %q", r.Method, path)
		http.ServeFile(w, r, path)
	}

	http.HandleFunc("/", serve)

	go func() {
		log.Printf("ListenAndServe on %q", bind)
		err := http.ListenAndServe(bind, nil)
		log.Fatalf("FATAL: Cannot ListenAndServe: %v: %q", err, bind)
	}()
}
