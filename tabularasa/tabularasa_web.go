package tabularasa

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (tabula *TabulaRasa) WebPluginSetup(router *mux.Router) {

	router.HandleFunc("/plugins/tabularasa", func(w http.ResponseWriter, r *http.Request) {

		tabula.TabulaRasta()

	})

}
