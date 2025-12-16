package chain

import (
	"errors"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
)

func parseDerivationPath(path string) ([]uint32, error) {
	if path == "" {
		return nil, errors.New("empty path")
	}
	if path[:2] != "m/" {
		return nil, errors.New("path must start with m/")
	}

	parts := strings.Split(path[2:], "/")
	var indices []uint32
	for _, p := range parts {
		hardened := false
		if strings.HasSuffix(p, "'") {
			hardened = true
			p = strings.TrimSuffix(p, "'")
		}
		num, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		if num < 0 {
			return nil, errors.New("negative index not allowed")
		}
		if hardened {
			num += hdkeychain.HardenedKeyStart
		}
		indices = append(indices, uint32(num))
	}
	return indices, nil
}
