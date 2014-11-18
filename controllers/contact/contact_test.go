package contact

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/curt-labs/GoAPI/helpers/testThatHttp"
	"github.com/curt-labs/GoAPI/models/contact"
	. "github.com/smartystreets/goconvey/convey"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestContact(t *testing.T) {
	var c contact.Contact
	var ct contact.ContactType
	var cr contact.ContactReceiver
	var err error
	Convey("Testing Contact", t, func() {

		//test create contact type using form
		form := url.Values{"name": {"test type"}}
		v := form.Encode()
		body := strings.NewReader(v)
		testThatHttp.Request("post", "/contact/types", "", "", AddContactType, body, "application/x-www-form-urlencoded")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &ct)
		So(err, ShouldBeNil)
		So(ct, ShouldHaveSameTypeAs, contact.ContactType{})

		//test create contact receiver using form
		form = url.Values{"first_name": {"test name"}, "last_name": {"test last name"}, "email": {"test@test.com"}, "contact_types": {strconv.Itoa(ct.ID)}}
		v = form.Encode()
		body = strings.NewReader(v)
		testThatHttp.Request("post", "/contact/receivers", "", "", AddContactReceiver, body, "application/x-www-form-urlencoded")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &cr)
		So(err, ShouldBeNil)
		So(cr, ShouldHaveSameTypeAs, contact.ContactReceiver{})

		//test create contact using json
		flag.Set("noEmail", "true") //do not send email during tests
		c.LastName = "smith"
		c.FirstName = "fred"
		c.Type = ct.Name
		c.Email = "test@test.com"
		c.Message = "test mes"
		c.Subject = "test sub"
		bodyBytes, _ := json.Marshal(c)
		bodyJson := bytes.NewReader(bodyBytes)
		testThatHttp.Request("post", "/contact/", ":contactTypeID", strconv.Itoa(ct.ID), AddDealerContact, bodyJson, "application/json")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &c)
		So(err, ShouldBeNil)
		So(c, ShouldHaveSameTypeAs, contact.Contact{})

		//test update contact using form
		form = url.Values{"last_name": {"formLastName"}}
		v = form.Encode()
		body = strings.NewReader(v)
		testThatHttp.Request("put", "/contact/", ":id", strconv.Itoa(c.ID), UpdateContact, body, "application/x-www-form-urlencoded")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &c)
		So(err, ShouldBeNil)
		So(c, ShouldHaveSameTypeAs, contact.Contact{})

		//test update contact using json
		c.LastName = "jsonLastName"
		bodyBytes, _ = json.Marshal(c)
		bodyJson = bytes.NewReader(bodyBytes)
		testThatHttp.Request("put", "/contact/", ":id", strconv.Itoa(c.ID), UpdateContact, bodyJson, "application/json")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &c)
		So(err, ShouldBeNil)
		So(c, ShouldHaveSameTypeAs, contact.Contact{})

		//test update contact type using form
		form = url.Values{"name": {"formName"}, "show": {"true"}}
		v = form.Encode()
		body = strings.NewReader(v)
		testThatHttp.Request("put", "/contact/types/", ":id", strconv.Itoa(ct.ID), UpdateContactType, body, "application/x-www-form-urlencoded")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &ct)
		So(err, ShouldBeNil)
		So(ct, ShouldHaveSameTypeAs, contact.ContactType{})

		//test update contact receiver using form
		form = url.Values{"first_name": {"new test name"}, "last_name": {"new test last name"}}
		v = form.Encode()
		body = strings.NewReader(v)
		testThatHttp.Request("put", "/contact/receivers/", ":id", strconv.Itoa(cr.ID), UpdateContactReceiver, body, "application/x-www-form-urlencoded")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &cr)
		So(err, ShouldBeNil)
		So(cr, ShouldHaveSameTypeAs, contact.ContactReceiver{})

		//test get contact
		testThatHttp.Request("get", "/contact/", ":id", strconv.Itoa(c.ID), GetContact, nil, "")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &c)
		So(err, ShouldBeNil)

		//test get all contacts
		form = url.Values{"page": {"1"}, "count": {"1"}}
		v = form.Encode()
		body = strings.NewReader(v)
		testThatHttp.Request("get", "/contact", "", "", GetAllContacts, body, "application/x-www-form-urlencoded")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		var cs contact.Contacts
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &cs)
		So(err, ShouldBeNil)
		So(cs, ShouldHaveSameTypeAs, contact.Contacts{})
		So(len(cs), ShouldBeGreaterThan, 0)

		//test get contact type
		testThatHttp.Request("get", "/contact/types/", ":id", strconv.Itoa(ct.ID), GetContactType, nil, "")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &ct)
		So(err, ShouldBeNil)
		So(ct, ShouldHaveSameTypeAs, contact.ContactType{})

		//test get all contact type
		testThatHttp.Request("get", "/contact/types", "", "", GetAllContactTypes, nil, "")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		var cts contact.ContactTypes
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &cts)
		So(err, ShouldBeNil)
		So(cts, ShouldHaveSameTypeAs, contact.ContactTypes{})
		So(len(cts), ShouldBeGreaterThan, 0)

		//test get receivers by contact type
		testThatHttp.Request("get", "/contact/types/receivers/", ":id", strconv.Itoa(ct.ID), GetReceiversByContactType, nil, "")
		var crs contact.ContactReceivers
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &crs)
		So(err, ShouldBeNil)
		So(crs, ShouldHaveSameTypeAs, contact.ContactReceivers{})
		So(len(crs), ShouldBeGreaterThan, 0)

		//test get contact receiver
		testThatHttp.Request("get", "/contact/receiver/", ":id", strconv.Itoa(cr.ID), GetContactReceiver, nil, "")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &cr)
		So(err, ShouldBeNil)

		//test get all contact receiver
		testThatHttp.Request("get", "/contact/receiver", "", "", GetAllContactReceivers, nil, "")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &crs)
		So(err, ShouldBeNil)
		So(crs, ShouldHaveSameTypeAs, contact.ContactReceivers{})
		So(len(crs), ShouldBeGreaterThan, 0)

		//test delete contact
		testThatHttp.Request("delete", "/contact/", ":id", strconv.Itoa(c.ID), DeleteContact, nil, "")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &c)
		So(err, ShouldBeNil)

		//test delete contact type
		testThatHttp.Request("delete", "/contact/types/", ":id", strconv.Itoa(ct.ID), DeleteContactType, nil, "")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &ct)
		So(err, ShouldBeNil)

		//test delete contact receiver
		testThatHttp.Request("delete", "/contact/receiver/", ":id", strconv.Itoa(cr.ID), DeleteContactReceiver, nil, "")
		So(testThatHttp.Response.Code, ShouldEqual, 200)
		err = json.Unmarshal(testThatHttp.Response.Body.Bytes(), &cr)
		So(err, ShouldBeNil)

	})
}
