package provider

import (
	"errors"

	gitbook "github.com/GitbookIO/go-gitbook/api"
)

func parseErrorMessage(err error) string {
	var openAPIErr *gitbook.GenericOpenAPIError
	if errors.As(err, &openAPIErr) {
		return string(openAPIErr.Body())
	}
	return err.Error()
}
