// Copyright 2011 Andy Balholm. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ICAP status codes.

package icap

import (
	"net/http"
)

var statusText = map[int]string{
	100: "Continue after ICAP Preview",
	204: "No modifications needed",
	400: "Bad request",
	404: "ICAP Service not found",
	405: "Method not allowed for service",
	408: "Request timeout",
	500: "Server error",
	501: "Method not implemented",
	502: "Bad Gateway",
	503: "Service overloaded",
	505: "ICAP version not supported by server",
}

// StatusText returns a text for the ICAP status code. It returns the empty string if the code is unknown.
func StatusText(code int) string {
	text, ok := statusText[code]
	if ok {
		return text
	}
	return http.StatusText(code)
}
