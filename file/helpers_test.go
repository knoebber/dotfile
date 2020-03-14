package file

import (
	"github.com/pkg/errors"
)

const (
	testAlias   = "testalias"
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
	saveRevisionErr     bool
	revertErr           bool
}

func (ms *MockStorer) GetTracked(string) (*Tracked, error) {
	if ms.getTrackedErr {
		return nil, errors.New("get tracked error")
	}
	if ms.testAliasNotTracked {
		return nil, nil
	}
	return new(Tracked), nil
}

func (ms *MockStorer) GetContents(string) ([]byte, error) {
	if ms.getContentsErr {
		return nil, errors.New("get contents error")
	}
	return []byte(testContent), nil
}

func (ms *MockStorer) SaveTracked(*Tracked) error {
	if ms.saveTrackedErr {
		return errors.New("save contents error")
	}
	return nil
}

func (ms *MockStorer) GetRevision(string, string) ([]byte, error) {
	if ms.getRevisionErr {
		return nil, errors.New("get contents error")
	}

	return []byte{}, nil
}

func (ms *MockStorer) SaveRevision(*Tracked, *Commit) error {
	if ms.saveRevisionErr {
		return errors.New("save revision error")
	}
	return nil
}

func (ms *MockStorer) Revert([]byte, string) error {
	if ms.revertErr {
		return errors.New("revert error")
	}

	return nil
}
