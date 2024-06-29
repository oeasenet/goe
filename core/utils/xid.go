package utils

import "github.com/rs/xid"

func GenXid() string {
	return xid.New().String()
}
