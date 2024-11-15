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

	respData, err := NextRespData(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read resp request")
	}

	array, ok := respData.(respModels.Array)
	if !ok {
		return nil, errors.Errorf("received invalid request, type '%c', value '%v'", respData.RespDataType(), respData)
	}

	var command = respModels.Request{}

	name, ok := array[0].(respModels.BulkString)
	if !ok {
		return nil, errors.Errorf("invalid command type %T", array[0])
	}
	command.Command = string(name)

	for _, bs := range array[1:] {
		arg, ok := bs.(respModels.BulkString)
		if !ok {
			return nil, errors.Errorf("invalid command type %T", bs)
		}

		command.Args = append(command.Args, string(arg))
	}

	return &command, nil
}
