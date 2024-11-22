package encoders

import (
	"github.com/pkg/errors"
	"github.com/xuanswe/mini-redis/internal/resp/models"
	"io"
)

func ReadRequest(r io.Reader) (*models.Request, error) {
	if r == nil {
		return nil, errors.Errorf("nil reader")
	}

	respData, err := NextRespData(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read resp request")
	}

	array, ok := respData.(models.Array)
	if !ok {
		return nil, errors.Errorf("received invalid request, type '%c', value '%v'", respData.RespDataType(), respData)
	}

	var command = models.Request{}

	name, ok := array[0].(models.BulkString)
	if !ok {
		return nil, errors.Errorf("invalid command type %T", array[0])
	}
	command.Command = string(name)

	for _, bs := range array[1:] {
		arg, ok := bs.(models.BulkString)
		if !ok {
			return nil, errors.Errorf("invalid command type %T", bs)
		}

		command.Args = append(command.Args, string(arg))
	}

	return &command, nil
}
