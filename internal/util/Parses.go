package util

import (
	"fmt"
	"strconv"
)

func StringToInt64(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		fmt.Println("Erro ao converter:", err)
	}
	return n
}