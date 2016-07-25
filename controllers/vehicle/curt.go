package vehicle

import (
	"github.com/curt-labs/API/helpers/apicontext"
	"github.com/curt-labs/API/helpers/encoding"
	"github.com/curt-labs/API/helpers/error"
	"github.com/curt-labs/API/models/products"

	"net/http"
)

func CurtLookup(w http.ResponseWriter, r *http.Request, enc encoding.Encoder, dtx *apicontext.DataContext) string {
	var v products.CurtVehicle

	// Get vehicle year
	v.Year = r.FormValue("year")
	delete(r.Form, "year")

	// Get vehicle make
	v.Make = r.FormValue("make")
	delete(r.Form, "make")

	// Get vehicle model
	v.Model = r.FormValue("model")
	delete(r.Form, "model")

	// Get vehicle submodel
	v.Style = r.FormValue("style")
	delete(r.Form, "style")

	cl := products.CurtLookup{
		CurtVehicle: v,
	}

	var err error
	if v.Year == "" {
		err = cl.GetYears()
	} else if v.Make == "" {
		err = cl.GetMakes()
	} else if v.Model == "" {
		err = cl.GetModels()
	} else {
		err = cl.GetStyles()
		if err != nil {
			apierror.GenerateError("Trouble finding styles.", err, w, r)
			return ""
		}
		err = cl.GetParts(dtx)
	}

	if err != nil {
		apierror.GenerateError("Trouble finding vehicles.", err, w, r)
		return ""
	}

	return encoding.Must(enc.Encode(cl))
}

func CURTApps(w http.ResponseWriter, r *http.Request, enc encoding.Encoder, dtx *apicontext.DataContext) string {
	var vr []products.VehicleApp
	var err error
	// get datestr from url
	var dateStr = ""
	vr, err = products.CurtVehicleApps(dateStr)

	if err != nil {
		apierror.GenerateError("Trouble finding vehicles.", err, w, r)
		return ""
	}

	return encoding.Must(enc.Encode(vr))
}
