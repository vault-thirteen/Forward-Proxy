package wm

import (
	"fmt"
	"io"
	"os"
	"strings"

	ipa "github.com/vault-thirteen/auxie/IPA"
	"github.com/vault-thirteen/auxie/reader"
	"github.com/vault-thirteen/errorz"
)

const (
	ErrDuplicateIPAddressInList = "duplicate IP address in list: %v"
)

type WhiteListMap = map[ipa.IPAddressV4]bool

func NewWhiteListMapFromFile(path string) (wlm WhiteListMap, err error) {
	var f *os.File
	f, err = os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		derr := f.Close()
		if derr != nil {
			err = errorz.Combine(err, derr)
		}
	}()

	r := reader.New(f)
	var line []byte
	var ipaddr ipa.IPAddressV4
	var isDuplicate bool
	wlm = make(WhiteListMap)
	for {
		line, err = r.ReadLineEndingWithCRLF()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}

		ipaddr, err = ipa.NewFromString(strings.TrimSpace(string(line)))
		if err != nil {
			return nil, err
		}

		_, isDuplicate = wlm[ipaddr]
		if isDuplicate {
			return nil, fmt.Errorf(ErrDuplicateIPAddressInList, string(line))
		}
		wlm[ipaddr] = true
	}

	return wlm, nil
}
