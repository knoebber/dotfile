package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/file"
)

// Gathers a file/commits and marshals it into the format that package local uses.
func handleFileJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]
	alias := vars["alias"]

	fileRecord, err := db.GetFileByUsername(username, alias)
	if err != nil {
		dbError(w, err)
		return
	}

	commits, err := db.GetCommitList(username, alias)
	if err != nil {
		dbError(w, err)
		return
	}

	result := &file.TrackingData{
		Path:     fileRecord.Path,
		Revision: fileRecord.CurrentRevision,
		Commits:  make([]file.Commit, len(commits)),
	}

	for i, c := range commits {
		result.Commits[i].Hash = c.Hash
		result.Commits[i].Message = c.Message
		result.Commits[i].Timestamp = c.Timestamp
	}

	setJSON(w, result)
}

func handleRawCompressedCommit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	commit, err := db.GetCommit(vars["username"], vars["alias"], vars["hash"])
	if err != nil {
		rawContentError(w, err)
		return
	}

	_, err = w.Write(commit.Revision)
	if err != nil {
		rawContentError(w, err)
		return
	}

	return
}

func setJSON(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("encoding json body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func dbError(w http.ResponseWriter, err error) {
	log.Print(err)
	if db.NotFound(err) {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
