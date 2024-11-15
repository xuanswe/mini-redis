package encoders

import (
	"github.com/pkg/errors"
	respModels "github.com/xuanswe/mini-redis/internal/models"
	"github.com/xuanswe/mini-redis/internal/support"
	"io"
	"strings"
)

func ReadRequest(r io.Reader) (*respModels.Request, error) {
	if r == nil {
		return nil, errors.Errorf("nil reader")
	}

	reader := support.EnsureBufferedReader(r)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, errors.Wrap(err, "failed to read line")
	}

	return &respModels.Request{Data: strings.TrimSpace(line)}, nil
}
