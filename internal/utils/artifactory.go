package utils

import (
	"net/http"
)

func HasANon404Error(err error, resp *http.Response) bool {
	if err != nil {
		if resp == nil || resp.StatusCode != http.StatusNotFound {
			return true
		}
	}
	return false
}
