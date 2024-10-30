package models

import (
	"fmt"

	"github.com/n0h4rt/chadango/utils"
)

// UploadedImage represents an uploaded image.
type UploadedImage struct {
	ID       int
	Username string
}

// ThumbURL returns a thumbnail url of the image.
//
// It formated as http://ust.chatango.com/um/u/s/username/img/t_{ID}.jpg
func (i UploadedImage) ThumbURL() string {
	return fmt.Sprintf(utils.UsernameToURL(API_UM_THUMB, i.Username), i.ID)
}

// LargeURL returns an url of the image.
//
// It formated as http://ust.chatango.com/um/u/s/username/img/l_{ID}.jpg
func (i UploadedImage) LargeURL() string {
	return fmt.Sprintf(utils.UsernameToURL(API_UM_LARGE, i.Username), i.ID)
}
