package utils

import (
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

var (
	typeDMG = filetype.NewType("dmg", "application/octect-stream")
	typeCOM = filetype.NewType("com", "unknown")

	registeredMatchers = map[types.Type]func([]byte) bool{
		typeDMG: func(buf []byte) bool {
			return len(buf) > 1 && (buf[0] == 0x78 && buf[1] == 0xDA)
		},
		typeCOM: func(buf []byte) bool {
			return len(buf) > 1 && (buf[0] == 0x58 && buf[1] == 0x35)
		},
	}
)

func init() {
	for ftype, matcher := range registeredMatchers { // adding the file types which are missing in the filetype package to better look for extension types during file processing
		filetype.AddMatcher(ftype, matcher)
	}
}
