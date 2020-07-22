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
	"github.com/knoebber/dotfile/file"
	"github.com/knoebber/dotfile/usererror"
	"github.com/pkg/errors"
)

// Gathers a file/commits and marshals it into file tracking data.
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

func validateAPIUser(w http.ResponseWriter, r *http.Request) *db.User {
	username, password, ok := r.BasicAuth()
	if !ok {
		authError(w, errors.New("basic auth not provided"))
		return nil
	}

	user, err := db.GetUser(username)
	if err != nil {
		authError(w, err)
		return nil
	}

	if user.CLIToken != password {
		authError(w, errors.New("user CLI token does not match password"))
		return nil
	}

	return user
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

func readPushedFileData(p *multipart.Part) (*file.TrackingData, error) {
	if p.Header.Get("Content-Type") != "application/json" {
		return nil, errors.New("expected json part to be content type application/json")
	}

	result := new(file.TrackingData)
	if err := json.NewDecoder(p).Decode(result); err != nil {
		return nil, errors.Wrap(err, "decoding pushed tracked file")
	}

	if err := p.Close(); err != nil {
		return nil, errors.Wrap(err, "closing json part")
	}

	return result, nil
}

func savePushedRevision(ft *db.FileTransaction, p *multipart.Part, commitMap map[string]*file.Commit) error {
	hash := p.FileName()
	buff := new(bytes.Buffer)

	n, err := buff.ReadFrom(p)
	if err != nil {
		err = errors.Wrapf(err, "reading revision %q from push (%d bytes)", hash, n)
		return ft.Rollback(err)
	}

	if err = p.Close(); err != nil {
		err = errors.Wrap(err, "closing revision part")
		return ft.Rollback(err)
	}

	c, ok := commitMap[hash]
	if !ok {
		err = fmt.Errorf("pushed revision %q doesn't exist in file data json", hash)
		return ft.Rollback(err)
	}

	if err := ft.SaveCommit(buff, c); err != nil {
		return err
	}

	log.Printf("saved %s (%d bytes)", hash, n)
	return nil
}

// Response body is expected to be multipart.
// The first part should be a JSON encoding of file.TrackingData
// Subsequent parts are the new revisions that the server should save.
// Each revision part should have be named as its hash.
func handlePush(w http.ResponseWriter, r *http.Request) {
	var mr *multipart.Reader

	vars := mux.Vars(r)
	username := vars["username"]
	alias := vars["alias"]

	user := validateAPIUser(w, r)
	if user == nil {
		return
	}

	if mr = getMultipartReader(w, r); mr == nil {
		return
	}

	jsonPart, err := mr.NextPart()
	if err != nil {
		apiError(w, errors.Wrap(err, "reading json part"))
		return
	}

	fileData, err := readPushedFileData(jsonPart)
	if err != nil {
		apiError(w, err)
		return
	}

	ft, err := db.NewFileTransaction(username, alias)
	if err != nil {
		apiError(w, err)
		return
	}

	if !ft.FileExists {
		if err := ft.SaveFile(user.ID, alias, fileData.Path); err != nil {
			apiError(w, err)
			return
		}
	} else {
		if ft.Path != fileData.Path {
			msg := fmt.Sprintf("local path %q does not match remote path %q", ft.Path, fileData.Path)
			apiError(w, usererror.Invalid(msg))
			return

		}
	}

	commitMap := fileData.MapCommits()

	for {
		revisionPart, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			err = ft.Rollback(errors.Wrap(err, "reading revision part"))
			apiError(w, err)
			return
		}

		if err = savePushedRevision(ft, revisionPart, commitMap); err != nil {
			apiError(w, err)
			return
		}
	}

	if err = ft.SetRevision(fileData.Revision); err != nil {
		apiError(w, err)
		return
	}

	if err = ft.Close(); err != nil {
		apiError(w, err)
		return
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
	var usererr *usererror.Error

	log.Print(err)
	if db.NotFound(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if errors.As(err, &usererr) {
		http.Error(w, usererr.Message, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}
