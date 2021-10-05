package controllers

import (
	"net/http"

	"github.com/dolfly/owncast/core"
	"github.com/dolfly/owncast/utils"
)

// Ping is fired by a client to show they are still an active viewer.
func Ping(w http.ResponseWriter, r *http.Request) {
	id := utils.GenerateClientIDFromRequest(r)
	core.SetViewerIDActive(id)
}
