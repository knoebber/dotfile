package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/knoebber/dotfile/db"
)

// Sets the contents of file to response writer.
func handleRawFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	file, err := db.File(db.Connection, vars["username"], vars["alias"])
	if err != nil {
		rawContentError(w, err)
		return
	}

	_, err = w.Write(file.Content)
	if err != nil {
		rawContentError(w, err)
		return
	}
}

// Sets the contents of file at hash to response writer.
func handleRawUncompressedCommit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	commit, err := db.UncompressedCommit(db.Connection, vars["username"], vars["alias"], vars["hash"])
	if err != nil {
		rawContentError(w, err)
		return
	}

	_, err = w.Write(commit.Content)
	if err != nil {
		rawContentError(w, err)
		return
	}

	return
}

func rawContentError(w http.ResponseWriter, err error) {
	if db.NotFound(err) {
		setError(w, err, "File not found", http.StatusNotFound)
	} else {
		setError(w, err, "Failed to retrieve raw content", http.StatusInternalServerError)
	}
}
