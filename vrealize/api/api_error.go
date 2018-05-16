package api

import "fmt"

//Error struct is used to store REST call errors
type Error struct {
	Errors []struct {
		Code          int    `json:"code"`
		Message       string `json:"message"`
		SystemMessage string `json:"systemMessage"`
	} `json:"errors"`
}

func (e Error) Error() string {
	return fmt.Sprintf("vRealize API: %+v", e.Errors)
}

func (e Error) isEmpty() bool {
	return len(e.Errors) == 0
}
