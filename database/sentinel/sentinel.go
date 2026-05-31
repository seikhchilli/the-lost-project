package sentinel

import "errors"

var ErrTitleNotFound = errors.New("Title not found")
var ErrTitleAlreadyExists = errors.New("Title already exists in your library")
