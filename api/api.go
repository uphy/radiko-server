package api

import "github.com/uphy/radiko-server/library"

type API struct {
	library *library.Library
	baseURL string
}

func New(library *library.Library, baseURL string) *API {
	return &API{library, baseURL}
}
