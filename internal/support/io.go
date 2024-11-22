package support

import (
	"bufio"
	"bytes"
	"github.com/pkg/errors"
	"io"
	"strconv"
	"strings"
)

func EnsureBufferedReader(r io.Reader) *bufio.Reader {
	if br, ok := r.(*bufio.Reader); ok {
		return br
	} else {
		return bufio.NewReader(r)
	}
}

func EnsureBufferedWriter(w io.Writer) *bufio.Writer {
	if br, ok := w.(*bufio.Writer); ok {
		return br
	} else {
		return bufio.NewWriter(w)
	}
}

// NextLine
// This function assumes that delim is a short string and does not optimize for long delim.
// The result doesn't contain delim
func NextLine(r io.Reader, delim []byte) (string, error) {
	if len(delim) == 0 {
		return "", errors.Errorf("delimiter cannot be empty")
	}

	br := EnsureBufferedReader(r)

	if len(delim) == 1 {
		return br.ReadString(delim[0])
	}

	var data strings.Builder
	firstChar := delim[0]
	lastChar := delim[len(delim)-1]
	for {
		b, err := br.ReadBytes(firstChar)
		if err != nil {
			return "", err
		}

		// read delim candidate
		delimCandidate, err := br.ReadBytes(lastChar)
		if err != nil {
			return "", err
		}

		eol := bytes.Equal(delimCandidate, delim[1:])

		length := len(b) + len(delimCandidate)
		if eol {
			b = b[:len(b)-1]
			length = len(b)
		}

		data.Grow(length)
		data.Write(b)

		if eol {
			break
		} else {
			data.Write(delimCandidate)
		}
	}

	return data.String(), nil
}

func NextLineAsInt(r io.Reader, delim []byte, base int, bitSize int) (int64, error) {
	line, err := NextLine(r, delim)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseInt(line, base, bitSize)
	if err != nil {
		return 0, err
	}

	return value, nil
}
