package warranty

import (
	"encoding/json"
	"github.com/curt-labs/GoAPI/helpers/encoding"
	"github.com/curt-labs/GoAPI/models/contact"
	"github.com/curt-labs/GoAPI/models/warranty"
	"github.com/go-martini/martini"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	timeFormat = "2006-01-02"
)

func GetAllWarranties(rw http.ResponseWriter, req *http.Request, enc encoding.Encoder) string {
	var err error

	ws, err := warranty.GetAllWarranties()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return err.Error()
	}
	return encoding.Must(enc.Encode(ws))
}

func GetWarranty(rw http.ResponseWriter, req *http.Request, enc encoding.Encoder, params martini.Params) string {
	var err error
	var w warranty.Warranty
	id := params["id"]
	w.ID, err = strconv.Atoi(id)

	err = w.Get()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return err.Error()
	}
	return encoding.Must(enc.Encode(w))
}

func GetWarrantyByContact(rw http.ResponseWriter, req *http.Request, enc encoding.Encoder, params martini.Params) string {
	var err error
	var w warranty.Warranty
	id := params["id"]
	w.Contact.ID, err = strconv.Atoi(id)

	ws, err := w.GetByContact()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return err.Error()
	}
	return encoding.Must(enc.Encode(ws))
}

func CreateWarranty(rw http.ResponseWriter, req *http.Request, enc encoding.Encoder, params martini.Params) string {
	contType := req.Header.Get("Content-Type")
	var w warranty.Warranty
	var err error

	contactTypeID, err := strconv.Atoi(params["contactReceiverTypeID"]) //to whom the emails go
	sendEmail, err := strconv.ParseBool(params["sendEmail"])

	if strings.Contains(contType, "application/json") {
		//json
		requestBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return encoding.Must(enc.Encode(false))
		}

		err = json.Unmarshal(requestBody, &w)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return encoding.Must(enc.Encode(false))
		}

	} else {
		//else, form
		w.PartNumber, err = strconv.Atoi(req.FormValue("part_number"))
		w.OldPartNumber = req.FormValue("old_part_number")
		date, err := time.Parse(timeFormat, req.FormValue("date"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return encoding.Must(enc.Encode(false))
		}
		w.Date = &date
		w.SerialNumber = req.FormValue("serial_number")

		w.Contact.FirstName = req.FormValue("first_name")
		w.Contact.LastName = req.FormValue("last_name")
		w.Contact.Email = req.FormValue("email")
		w.Contact.Phone = req.FormValue("phone")
		w.Contact.Type = req.FormValue("type")
		w.Contact.Address1 = req.FormValue("address1")
		w.Contact.Address2 = req.FormValue("address2")
		w.Contact.City = req.FormValue("city")
		w.Contact.State = req.FormValue("state")
		w.Contact.PostalCode = req.FormValue("postal_code")
		w.Contact.Country = req.FormValue("country")
	}
	err = w.Create()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return err.Error()
	}
	if sendEmail == true {
		//Send Email
		body :=
			"Name: " + w.Contact.FirstName + " " + w.Contact.LastName + "\n" +
				"Email: " + w.Contact.Email + "\n" +
				"Phone: " + w.Contact.Phone + "\n" +
				"Serial Number: " + w.SerialNumber + "\n" +
				"Date: " + w.Date.String() + "\n" +
				"Part Number: " + strconv.Itoa(w.PartNumber) + "\n"

		var ct contact.ContactType
		ct.ID = contactTypeID
		subject := "Email from Warranty Applications Form"
		err = contact.SendEmail(ct, subject, body) //contact type id, subject, techSupport
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return err.Error()
		}
	}
	//Return JSON
	return encoding.Must(enc.Encode(w))
}

func DeleteWarranty(rw http.ResponseWriter, req *http.Request, enc encoding.Encoder, params martini.Params) string {
	var err error
	var w warranty.Warranty
	id := params["id"]
	w.ID, err = strconv.Atoi(id)

	err = w.Delete()

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return err.Error()
	}

	//Return JSON
	return encoding.Must(enc.Encode(w))
}