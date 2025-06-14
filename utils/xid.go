package utils

import "github.com/rs/xid"

func GenerateXid() string {
	return xid.New().String()
}
