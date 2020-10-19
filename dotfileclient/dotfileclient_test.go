package dotfileclient

import (
	"fmt"
	"github.com/knoebber/dotfile/dotfile"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testTrackingData = `
{
    "commits": [
        {
            "hash": "000d687705f0be9cef73a8599cdfc215d591dae2",
            "message": "Initial commit",
            "timestamp": 1595600967
        },
        {
            "hash": "40a86bc3b22dfe3ab92a64390599d18c7bed7e88",
            "message": "",
            "timestamp": 1602690861
        }
    ],
    "path": "~/.bash_ps1",
    "revision": "40a86bc3b22dfe3ab92a64390599d18c7bed7e88"
}
`

func setupTest(code int, response string) (*httptest.Server, *Client) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		fmt.Fprintln(w, response)
	}))

	client := New(ts.URL, "test", "test")
	return ts, client
}

func TestClient_List(t *testing.T) {
	t.Run("http error", func(t *testing.T) {
		client := New("no host", "test", "test")
		_, err := client.List(true)
		assert.Error(t, err)
	})

	t.Run("not 200 error", func(t *testing.T) {
		ts, client := setupTest(http.StatusInternalServerError, "[]")
		defer ts.Close()

		_, err := client.List(true)
		assert.Error(t, err)
	})

	t.Run("json error", func(t *testing.T) {
		ts, client := setupTest(http.StatusOK, "json error")
		defer ts.Close()

		_, err := client.List(true)
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		ts, client := setupTest(http.StatusOK, `["file1", "file2","file3"]`)
		defer ts.Close()
		result, err := client.List(true)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}

func TestClient_TrackingData(t *testing.T) {
	t.Run("http error", func(t *testing.T) {
		client := New("no host", "test", "test")
		_, err := client.TrackingDataBytes("test")
		assert.Error(t, err)
	})

	t.Run("error status code", func(t *testing.T) {
		ts, client := setupTest(http.StatusBadRequest, testTrackingData)
		defer ts.Close()

		res, err := client.TrackingData("test")
		assert.Error(t, err)
		assert.Empty(t, res)
	})

	t.Run("ok and nil on 404", func(t *testing.T) {
		ts, client := setupTest(http.StatusNotFound, testTrackingData)
		defer ts.Close()

		res, err := client.TrackingData("test")
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("error on invalid JSON", func(t *testing.T) {
		ts, client := setupTest(http.StatusOK, "bad json")
		defer ts.Close()

		_, err := client.TrackingData("test")
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		ts, client := setupTest(http.StatusOK, testTrackingData)
		defer ts.Close()

		res, err := client.TrackingData("test")
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	})
}

func TestClient_Revisions(t *testing.T) {
	t.Run("no requests with no hashes", func(t *testing.T) {
		client := New("no host", "test", "test")
		res, err := client.Revisions("test", []string{})
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("http error", func(t *testing.T) {
		client := New("no host", "test", "test")
		_, err := client.Revisions("test", []string{"a", "b"})
		assert.Error(t, err)
	})

	t.Run("non 200 error", func(t *testing.T) {
		ts, client := setupTest(http.StatusBadGateway, "")
		defer ts.Close()

		_, err := client.Revisions("test", []string{"a", "b"})
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		ts, client := setupTest(http.StatusOK, "content")
		defer ts.Close()

		_, err := client.Revisions("test", []string{"a", "b"})
		assert.NoError(t, err)
	})
}

func TestClient_Content(t *testing.T) {
	t.Run("http error", func(t *testing.T) {
		client := New("no host", "test", "test")
		_, err := client.Content("test")
		assert.Error(t, err)
	})

	t.Run("non 200 error", func(t *testing.T) {
		ts, client := setupTest(http.StatusNotFound, "")
		defer ts.Close()

		res, err := client.Content("test")
		assert.Error(t, err)
		println(res)
	})

	t.Run("ok", func(t *testing.T) {
		ts, client := setupTest(http.StatusOK, "content")
		defer ts.Close()

		_, err := client.Content("test")
		assert.NoError(t, err)
	})
}

func TestClient_UploadRevisions(t *testing.T) {
	t.Run("ok with no hashes", func(t *testing.T) {
		client := New("no host", "test", "test")
		assert.NoError(t, client.UploadRevisions("", new(dotfile.TrackingData), []*Revision{}))
	})

	t.Run("http error", func(t *testing.T) {
		client := New("no host", "test", "test")
		assert.Error(t, client.UploadRevisions("", new(dotfile.TrackingData), []*Revision{{}}))
	})

	t.Run("error on non 200", func(t *testing.T) {
		ts, client := setupTest(http.StatusInternalServerError, "")
		defer ts.Close()
		assert.Error(t, client.UploadRevisions("", new(dotfile.TrackingData), []*Revision{{}}))
	})

	t.Run("ok", func(t *testing.T) {
		ts, client := setupTest(http.StatusOK, "")
		defer ts.Close()
		assert.NoError(t, client.UploadRevisions("", new(dotfile.TrackingData), []*Revision{{}}))
	})
}
