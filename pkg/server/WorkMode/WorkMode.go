package wm

import (
	"fmt"
	"strings"
)

const (
	ErrUnknownWorkModeString = "unknown work mode name: %v"
)

// WorkModeString.
const (
	WorkModeStringDefault = WorkModeStringPublic
	WorkModeStringPublic  = "public"
	WorkModeStringPrivate = "private"
)

// WorkModeByte.
const (
	WorkModePublic  = 1
	WorkModePrivate = 2
)

type WorkMode struct {
	mode              byte
	whitelistFilePath string
	whiteListMap      WhiteListMap
}

func New(workModeString string, listFile string) (wm *WorkMode, err error) {
	var mode byte
	switch strings.ToLower(workModeString) {
	case strings.ToLower(WorkModeStringPublic):
		mode = WorkModePublic

	case strings.ToLower(WorkModeStringPrivate):
		mode = WorkModePrivate

	default:
		return nil, fmt.Errorf(ErrUnknownWorkModeString, workModeString)
	}

	wm = &WorkMode{
		mode: mode,
	}

	// Configure a white-list for the private mode.
	if mode == WorkModePrivate {
		wm.whitelistFilePath = listFile

		wm.whiteListMap, err = NewWhiteListMapFromFile(wm.whitelistFilePath)
		if err != nil {
			return nil, err
		}

		return wm, nil
	}

	return wm, nil
}

func (wm *WorkMode) IsPublic() bool {
	return wm.mode == WorkModePublic
}

func (wm *WorkMode) IsPrivate() bool {
	return wm.mode == WorkModePrivate
}

func (wm *WorkMode) WhiteList() WhiteListMap {
	return wm.whiteListMap
}
