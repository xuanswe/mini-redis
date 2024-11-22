package encoders

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/xuanswe/mini-redis/internal/resp/consts/datatypes"
	"github.com/xuanswe/mini-redis/internal/resp/models"
	"github.com/xuanswe/mini-redis/internal/support"
	"io"
	"strconv"
)

var respTerminator = []byte("\r\n")

func NextRespData(r io.Reader) (models.RespData, error) {
	dataType, err := nextDataType(r)
	if err != nil {
		return nil, err
	}

	switch dataType {
	case datatypes.SimpleIntType:
		return NextSimpleInt(r)
	case datatypes.SimpleStringType:
		return NextSimpleString(r)
	case datatypes.BulkStringType:
		return NextBulkString(r)
	case datatypes.ArrayType:
		return NextArray(r)
	}

	return nil, errors.Errorf("unknown RESP data type: %d", dataType)
}

func nextDataType(r io.Reader) (datatypes.RespDataType, error) {
	dataType, err := r.(io.ByteReader).ReadByte()
	if err != nil {
		return 0, err
	}
	return rune(dataType), nil
}

func NextSimpleInt(r io.Reader) (models.RespData, error) {
	line, err := support.NextLine(r, respTerminator)
	if err != nil {
		return nil, err
	}

	parseInt, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return nil, err
	}

	return models.SimpleInt(parseInt), nil
}

func NextSimpleString(r io.Reader) (models.RespData, error) {
	line, err := support.NextLine(r, respTerminator)
	if err != nil {
		return nil, err
	}

	return models.SimpleString(line), nil
}

func SimpleString(str string) string {
	return fmt.Sprintf("+%s\r\n", str)
}

func SimpleError(str string) string {
	return fmt.Sprintf("-%s\r\n", str)
}

func NextBulkString(r io.Reader) (models.RespData, error) {
	length, err := support.NextLineAsInt(r, respTerminator, 10, 64)
	if err != nil {
		return nil, err
	}

	if length == -1 {
		return nil, nil
	} else if length < 0 {
		return nil, errors.New("invalid length for RESP BulkString")
	}

	data := make([]byte, length)
	_, err = io.ReadFull(r, data)
	if err != nil {
		return nil, err
	}

	terminator := make([]byte, len(respTerminator))
	_, err = io.ReadFull(r, terminator)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(terminator, respTerminator) {
		return nil, errors.New("invalid terminator for RESP BulkString")
	}

	return models.BulkString(data), nil
}

const NullBulkString = "$-1\r\n"

func BulkString(str string) string {
	return fmt.Sprintf("$%s\r\n%s\r\n", strconv.Itoa(len(str)), str)
}

func NextArray(r io.Reader) (models.RespData, error) {
	length, err := support.NextLineAsInt(r, respTerminator, 10, 64)
	if err != nil {
		return nil, err
	}

	if length == -1 {
		return nil, nil
	} else if length < 0 {
		return nil, errors.New("invalid length for RESP Array")
	}

	data := make([]models.RespData, length)
	for i := int64(0); i < length; i++ {
		data[i], err = NextRespData(r)
	}

	return models.Array(data), nil
}
