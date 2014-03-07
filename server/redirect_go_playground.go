package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "http://play.golang.org"

func (srv *Server) handleRedirectToGoPlaygroundAppEngine(w http.ResponseWriter, r *http.Request) {
	b := new(bytes.Buffer)

	if err := passThru(b, r); err != nil {
		http.Error(w, "Server error.", http.StatusInternalServerError)
		srv.logger.Errorf("while redirect to go playground: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.Copy(w, b)
}

func passThru(w io.Writer, req *http.Request) error {
	defer req.Body.Close()
	url := baseURL + req.URL.Path
	resp, err := http.DefaultClient.Post(url, req.Header.Get("Content-type"), req.Body)
	if err != nil {
		return fmt.Errorf("making POST request: %v", err)
	}

	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("copying response Body: %v", err)
	}
	return nil
}
