package utils

import (
	"fmt"
	"net/http"
)

func HasANon404Error(err error, resp *http.Response) bool {
	if err != nil && (resp == nil || resp.StatusCode != http.StatusNotFound) {
		fmt.Println("Technical error occured when looking at existing group: ", err)
		return true
	}

	return false
}
