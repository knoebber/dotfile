package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/file"
	"github.com/pkg/errors"
)

// Gathers a file/commits and marshals it into the format that package local uses.
func handleFileJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]
	alias := vars["alias"]

	fileRecord, err := db.GetFile(username, alias)
	if err != nil {
		apiError(w, err)
		return
	}

	commits, err := db.GetCommitList(username, alias)
	if err != nil {
		apiError(w, err)
		return
	}

	result := &file.TrackingData{
		Path:     fileRecord.Path,
		Revision: fileRecord.Hash,
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

func validateAPIUser(w http.ResponseWriter, r *http.Request) (ok bool) {
	username, password, ok := r.BasicAuth()
	if !ok {
		authError(w, errors.New("basic auth not provided"))
		return
	}

	user, err := db.GetUser(username)
	if err != nil {
		authError(w, err)
		return
	}

	if user.CLIToken != password {
		authError(w, errors.New("user CLI token does not match password"))
		return
	}

	return true
}

func getMultipartReader(w http.ResponseWriter, r *http.Request) *multipart.Reader {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		apiError(w, err)
		return nil
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		apiError(w, errors.New("expected multipart content type"))
		return nil
	}

	return multipart.NewReader(r.Body, params["boundary"])
}

func readPushedFileData(p *multipart.Part, w http.ResponseWriter) *file.TrackingData {
	result := new(file.TrackingData)
	if err := json.NewDecoder(p).Decode(result); err != nil {
		apiError(w, errors.Wrap(err, "decoding pushed tracked file"))
		return nil
	}

	return result
}

func savePushedRevision(w http.ResponseWriter, tx *db.FileTransaction, mr *multipart.Reader) (hasMore bool) {

	p, err := mr.NextPart()
	if err == io.EOF {
		return
	}

	if err != nil {
		apiError(w, errors.Wrap(err, "reading revision parts"))
		return
	}

	hash := p.FileName()
	buff := new(bytes.Buffer)

	_, err = buff.ReadFrom(p)
	if err != nil {
		apiError(w, errors.Wrap(err, "reading revision from push"))
		return
	}

	tx.SaveCommit(buff, &file.Commit{
		// TODO - other fields.
		Hash: hash,
	})

	hasMore = true
	return
}

// Response body is expected to be multipart.
// The first part should be a JSON encoding of file.TrackingData
// Subsequent parts are the new revisions that the server should save.
// Each revision part should have be named as its hash.
func handlePush(w http.ResponseWriter, r *http.Request) {
	var (
		fileData *file.TrackingData
		mr       *multipart.Reader
	)

	vars := mux.Vars(r)
	username := vars["username"]
	alias := vars["alias"]

	if ok := validateAPIUser(w, r); !ok {
		return
	}
	if mr = getMultipartReader(w, r); mr == nil {
		return
	}

	jsonPart, err := mr.NextPart()
	if err != nil {
		apiError(w, errors.Wrap(err, "reading JSON part"))
		return
	}
	if jsonPart.Header.Get("Content-Type") != "application/json" {
		apiError(w, errors.New("expected first part to be application/json"))
		return
	}

	if fileData = readPushedFileData(jsonPart, w); fileData == nil {
		return
	}

	tx, err := db.NewFileTransaction(username, alias)
	if err != nil {
		apiError(w, err)
		return
	}

	for savePushedRevision(w, tx, mr) {
	}
}

func setJSON(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("encoding json body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func authError(w http.ResponseWriter, err error) {
	log.Print(err)
	w.WriteHeader(http.StatusUnauthorized)
}

func apiError(w http.ResponseWriter, err error) {
	log.Print(err)
	if db.NotFound(err) {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
