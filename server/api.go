package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/knoebber/dotfile/db"
	"github.com/knoebber/dotfile/dotfile"
	"github.com/knoebber/dotfile/usererror"
	"github.com/pkg/errors"
)

// Gathers a file/commits and marshals it into file tracking data.
func handleFileJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	fileData, err := db.FileData(db.Connection, vars["username"], vars["alias"])
	if err != nil {
		apiError(w, err)
		return
	}

	setJSON(w, fileData)
}

// Sets a list of aliases that username owns to the response body.
func handleFileListJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]
	addPath := r.URL.Query().Get("path") == "true"

	files, err := db.FilesByUsername(db.Connection, username, nil)
	if err != nil {
		apiError(w, err)
		return
	}

	result := make([]string, len(files))
	for i, f := range files {
		result[i] = f.Alias
		if addPath {
			result[i] += " " + f.Path
		}
	}

	setJSON(w, result)
}

func handleRawCompressedCommit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	commit, err := db.Commit(db.Connection, vars["username"], vars["alias"], vars["hash"])
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

func validateAPIUser(w http.ResponseWriter, r *http.Request) int64 {
	username, token, ok := r.BasicAuth()
	if !ok {
		authError(w, errors.New("basic auth not provided"))
		return 0
	}

	userID, err := db.UserLoginAPI(db.Connection, username, token)
	if err != nil {
		authError(w, err)
		return 0
	}

	return userID
}

func multipartReader(w http.ResponseWriter, r *http.Request) *multipart.Reader {
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

func readPushedFileData(p *multipart.Part) (*dotfile.TrackingData, error) {
	if p.Header.Get("Content-Type") != "application/json" {
		return nil, errors.New("expected json part to be content type application/json")
	}

	result := new(dotfile.TrackingData)
	if err := json.NewDecoder(p).Decode(result); err != nil {
		return nil, errors.Wrap(err, "decoding pushed tracked file")
	}

	if err := p.Close(); err != nil {
		return nil, errors.Wrap(err, "closing json part")
	}

	return result, nil
}

func savePushedRevision(ft *db.FileTransaction, p *multipart.Part, commitMap map[string]*dotfile.Commit) error {
	hash := p.FileName()
	buff := new(bytes.Buffer)

	n, err := buff.ReadFrom(p)
	if err != nil {
		return errors.Wrapf(err, "reading revision %q from push (%d bytes)", hash, n)
	}

	if err = p.Close(); err != nil {
		return errors.Wrap(err, "closing revision part")
	}

	c, ok := commitMap[hash]
	if !ok {
		return fmt.Errorf("pushed revision %q doesn't exist in file data json", hash)

	}

	if err := ft.SaveCommit(buff, c); err != nil {
		return err
	}

	log.Printf("saved %s (%d bytes)", hash, n)
	return nil
}

func push(mr *multipart.Reader, userID int64, alias string) error {
	jsonPart, err := mr.NextPart()
	if err != nil {
		return errors.Wrap(err, "reading json part")
	}

	fileData, err := readPushedFileData(jsonPart)
	if err != nil {
		return err
	}

	tx, err := db.Connection.Begin()
	if err != nil {
		return errors.Wrap(err, "starting transaction for handle push")
	}

	ft, err := db.NewFileTransaction(tx, userID, alias)
	if err != nil {
		return db.Rollback(tx, err)
	}

	if !ft.FileExists {
		if err := ft.SaveFile(userID, alias, fileData.Path); err != nil {
			return db.Rollback(tx, err)
		}
	} else {
		if ft.Path != fileData.Path {
			return db.Rollback(tx, usererror.Invalid(fmt.Sprintf(
				"local path %q does not match remote path %q",
				ft.Path, fileData.Path)))
		}
	}

	commitMap := fileData.MapCommits()

	for {
		revisionPart, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			return db.Rollback(tx, errors.Wrap(err, "reading revision part"))
		}

		if err = savePushedRevision(ft, revisionPart, commitMap); err != nil {
			return db.Rollback(tx, err)
		}
	}

	if err = ft.SetRevision(fileData.Revision); err != nil {
		return db.Rollback(tx, err)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing handle push transaction")
	}

	return nil
}

// Request body is expected to be multipart.
// The first part is a JSON encoding of dotfile.TrackingData
// Subsequent parts are new revisions that need to be saved.
// Each revision part should have be named as its hash.
func handlePush(w http.ResponseWriter, r *http.Request) {
	var mr *multipart.Reader

	userID := validateAPIUser(w, r)
	if userID < 1 {
		return
	}

	if mr = multipartReader(w, r); mr == nil {
		return
	}

	if err := push(mr, userID, mux.Vars(r)["alias"]); err != nil {
		apiError(w, err)
		return
	}

}

func setJSON(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("encoding json body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func authError(w http.ResponseWriter, err error) {
	log.Print(err)
	w.WriteHeader(http.StatusUnauthorized)
}

func apiError(w http.ResponseWriter, err error) {
	var usererr *usererror.Error

	if db.NotFound(err) {
		// Clients expect this when a file doesn't exist.
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Printf("api error: %v", err)

	if errors.As(err, &usererr) {
		http.Error(w, usererr.Message, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}
