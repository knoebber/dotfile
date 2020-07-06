package file

import (
	"bytes"
	"errors"
	"time"
)

const (
	testPath    = "/testpath"
	invalidPath = "&"
	testMessage = "Test commit message"
	testHash    = "9abdbcf4ea4e2c1c077c21b8c2f2470ff36c31ce"
	testContent = "test content!\nanother line!\n"
)

type MockStorer struct {
	getTrackedErr       bool
	testAliasNotTracked bool
	getContentsErr      bool
	saveTrackedErr      bool
	getRevisionErr      bool
	uncompressErr       bool
	saveCommitErr       bool
	revertErr           bool
	hasCommit           bool
	hasCommitErr        bool
	closeErr            bool
}

func (ms *MockStorer) HasCommit(string) (bool, error) {
	if ms.hasCommitErr {
		return false, errors.New("has commit error")
	}
	return ms.hasCommit, nil
}

func (ms *MockStorer) GetContents() ([]byte, error) {
	if ms.getContentsErr {
		return nil, errors.New("get contents error")
	}
	return []byte(testContent), nil
}

func (ms *MockStorer) GetRevision(string) ([]byte, error) {
	if ms.getRevisionErr {
		return nil, errors.New("get contents error")
	}
	if ms.uncompressErr {
		return nil, nil
	}

	compressed, _, err := hashAndCompress([]byte(testContent))
	if err != nil {
		return nil, err
	}

	return compressed.Bytes(), nil
}

func (ms *MockStorer) SaveCommit(*bytes.Buffer, string, string, time.Time) error {
	if ms.saveCommitErr {
		return errors.New("save revision error")
	}
	return nil
}

func (ms *MockStorer) Revert(*bytes.Buffer, string) error {
	if ms.revertErr {
		return errors.New("revert error")
	}

	return nil
}

func (ms *MockStorer) Close() error {
	if ms.closeErr {
		return errors.New("close error")
	}

	return nil
}
