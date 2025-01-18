package view

import (
	"context"
	"ht/helper"
	"net/url"
)

func GetCurrentUrl(c context.Context) string {
	url, ok := c.Value(helper.UrlKey).(*url.URL)
	if !ok {
		return ""
	}
	return url.Path
}
