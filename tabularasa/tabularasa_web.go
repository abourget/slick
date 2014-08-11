package tabularasa

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func (tabula *TabulaRasa) WebPluginSetup(router *mux.Router) {
	router.HandleFunc("/plugins/tabulaRasa", func(w http.ResponseWriter, r *http.Request) {
}
