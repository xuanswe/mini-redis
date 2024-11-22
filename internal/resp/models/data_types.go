package models

import "github.com/xuanswe/mini-redis/internal/resp/consts/datatypes"

type RespData interface {
	RespDataType() datatypes.RespDataType
}

type Array []RespData

func (s Array) RespDataType() datatypes.RespDataType {
	return datatypes.ArrayType
}

type SimpleInt int64

func (s SimpleInt) RespDataType() datatypes.RespDataType {
	return datatypes.SimpleIntType
}

type SimpleString string

func (s SimpleString) RespDataType() datatypes.RespDataType {
	return datatypes.SimpleStringType
}

type BulkString string

func (s BulkString) RespDataType() datatypes.RespDataType {
	return datatypes.BulkStringType
}
