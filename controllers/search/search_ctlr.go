package search_ctlr

import (
	"github.com/curt-labs/GoAPI/helpers/apicontext"
	"net/http"
	"strconv"

	"github.com/curt-labs/GoAPI/helpers/encoding"
	"github.com/curt-labs/GoAPI/helpers/error"
	"github.com/curt-labs/GoAPI/models/search"
	"github.com/go-martini/martini"
)

func Search(rw http.ResponseWriter, r *http.Request, params martini.Params, enc encoding.Encoder, dtx *apicontext.DataContext) string {
	terms := params["term"]
	qs := r.URL.Query()
	page, _ := strconv.Atoi(qs.Get("page"))
	count, _ := strconv.Atoi(qs.Get("count"))

	res, err := search.Dsl(terms, page, count, dtx)
	if err != nil {
		apierror.GenerateError("Trouble searching", err, rw, r)
		return ""
	}

	return encoding.Must(enc.Encode(res))
}
