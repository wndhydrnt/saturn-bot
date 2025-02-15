package ui

import (
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/version"
)

type dataInfoIndex struct {
	Version version.VersionInfo
}

func (u *Ui) StatusIndex(w http.ResponseWriter, r *http.Request) {
	data := dataInfoIndex{Version: version.Info}
	renderTemplate(data, w, "status_index.html")
}
