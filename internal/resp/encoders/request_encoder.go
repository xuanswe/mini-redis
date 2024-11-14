package encoders

import (
	"github.com/pkg/errors"
	respModels "github.com/xuanswe/mini-redis/internal/resp/models"
	"io"
)

func ReadRequest(r io.Reader) (*respModels.Request, error) {
	if r == nil {
		return nil, errors.Errorf("nil reader")
	}

	// TODO: implementation

	return &respModels.Request{}, nil
}
