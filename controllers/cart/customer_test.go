package cart_ctlr

import (
	"encoding/json"
	"github.com/curt-labs/GoAPI/helpers/error"
	"github.com/curt-labs/GoAPI/helpers/httprunner"
	"github.com/curt-labs/GoAPI/models/cart"
	. "github.com/smartystreets/goconvey/convey"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

var (
	response *httptest.ResponseRecorder
)

func TestGetCustomers(t *testing.T) {
	Convey("Testing GetCustomers", t, func() {
		Convey("no shop identifier", func() {
			response = httprunner.Request("GET", "/shopify/customers", nil, GetCustomers)
			So(response.Code, ShouldEqual, 500)
			So(response.Body.String(), ShouldNotEqual, "[]")

			vals := make(url.Values, 0)
			vals.Add("shop", "testing")
			response = httprunner.Request("GET", "/shopify/customers", &vals, GetCustomers)
			So(response.Code, ShouldEqual, 500)
			So(response.Body.String(), ShouldNotEqual, "[]")

			vals.Add("since_id", "something")
			response = httprunner.Request("GET", "/shopify/customers", &vals, GetCustomers)
			So(response.Code, ShouldEqual, 500)
			So(response.Body.String(), ShouldNotEqual, "[]")
		})
		Convey("with shop identifier", func() {
			shopID := cart.InsertTestData()
			So(shopID, ShouldNotBeNil)

			val := shopID.Hex()
			qs := make(url.Values, 0)
			qs.Add("shop", val)

			response = httprunner.Request("GET", "/shopify/customers", &qs, GetCustomers)
			So(response.Code, ShouldEqual, 200)
			So(json.Unmarshal(response.Body.Bytes(), &[]cart.Customer{}), ShouldBeNil)

			qs.Add("since_id", "something")
			response = httprunner.Request("GET", "/shopify/customers", &qs, GetCustomers)
			So(response.Code, ShouldEqual, 200)
			So(json.Unmarshal(response.Body.Bytes(), &[]cart.Customer{}), ShouldBeNil)

			qs.Set("since_id", val)
			response = httprunner.Request("GET", "/shopify/customers", &qs, GetCustomers)
			So(response.Code, ShouldEqual, 200)
			var custs []cart.Customer
			So(json.Unmarshal(response.Body.Bytes(), &custs), ShouldBeNil)
			So(len(custs), ShouldEqual, 0)

			created_min := time.Now().AddDate(-1, 0, 0).Format(TimeLayout)
			created_max := time.Now().Format(TimeLayout)
			qs.Add("created_at_min", created_min)
			qs.Add("created_at_max", created_max)
			response = httprunner.Request("GET", "/shopify/customers", &qs, GetCustomers)
			So(response.Code, ShouldEqual, 200)
			So(json.Unmarshal(response.Body.Bytes(), &custs), ShouldBeNil)
			So(len(custs), ShouldEqual, 0)

			update_min := time.Now().AddDate(-1, 0, 0).Format(TimeLayout)
			update_max := time.Now().Format(TimeLayout)
			qs.Add("updated_at_min", update_min)
			qs.Add("updated_at_max", update_max)
			response = httprunner.Request("GET", "/shopify/customers", &qs, GetCustomers)
			So(response.Code, ShouldEqual, 200)
			So(json.Unmarshal(response.Body.Bytes(), &custs), ShouldBeNil)
			So(len(custs), ShouldEqual, 0)

			qs.Add("page", "bad")
			qs.Add("limit", "test")
			response = httprunner.Request("GET", "/shopify/customers", &qs, GetCustomers)
			So(response.Code, ShouldEqual, 200)
			So(json.Unmarshal(response.Body.Bytes(), &custs), ShouldBeNil)
			So(len(custs), ShouldEqual, 0)

			qs.Set("page", "1")
			qs.Set("limit", "1")
			response = httprunner.Request("GET", "/shopify/customers", &qs, GetCustomers)
			So(response.Code, ShouldEqual, 200)
			So(json.Unmarshal(response.Body.Bytes(), &custs), ShouldBeNil)
			So(len(custs), ShouldEqual, 0)
		})
	})
}

func BenchmarkGetCustomers(b *testing.B) {
	shopID := cart.InsertTestData()
	if shopID == nil {
		b.Error("shopID cannot be nil")
		b.Fail()
	}

	val := shopID.Hex()
	qs := make(url.Values, 0)
	qs.Add("shop", val)
	qs.Add("since_id", val)
	(&httprunner.BenchmarkOptions{
		Method:             "GET",
		Route:              "/shopify/customers",
		ParameterizedRoute: "/shopify/customers",
		Handler:            GetCustomer,
		QueryString:        &qs,
		JsonBody:           nil,
		Runs:               b.N,
	}).RequestBenchmark()

}

func TestAddCustomer(t *testing.T) {
	Convey("with no shop identifier", t, func() {
		qs := make(url.Values, 0)

		resp := httprunner.JsonRequest("POST", "/shopify/customers", &qs, cart.Shop{}, AddCustomer)
		So(resp.Code, ShouldEqual, 500)
		So(json.Unmarshal(resp.Body.Bytes(), &apierror.ApiErr{}), ShouldBeNil)
	})

	Convey("with shop identifier", t, func() {
		shopID := cart.InsertTestData()
		So(shopID, ShouldNotBeNil)
		val := shopID.Hex()
		qs := make(url.Values, 0)
		qs.Add("shop", val)

		response = httprunner.JsonRequest("POST", "/shopify/customers", &qs, cart.Shop{}, AddCustomer)
		So(response.Code, ShouldEqual, 500)
		So(json.Unmarshal(response.Body.Bytes(), &apierror.ApiErr{}), ShouldBeNil)

		cust := cart.Customer{
			ShopId:    *shopID,
			FirstName: "Alex",
			LastName:  "Ninneman",
		}

		response = httprunner.JsonRequest("POST", "/shopify/customers", &qs, cust, AddCustomer)
		So(response.Code, ShouldEqual, 500)
		So(json.Unmarshal(response.Body.Bytes(), &apierror.ApiErr{}), ShouldBeNil)

		cust.Email = "ninnemana@gmail.com"
		response = httprunner.JsonRequest("POST", "/shopify/customers", &qs, cust, AddCustomer)
		So(response.Code, ShouldEqual, 200)
		So(json.Unmarshal(response.Body.Bytes(), &cust), ShouldBeNil)
	})
}

func TestGetCustomer(t *testing.T) {
	Convey("no shop identifier", t, func() {
		response = httprunner.Request("GET", "/shopify/customers/1234", nil, GetCustomers)
		So(response.Code, ShouldEqual, 500)
		So(json.Unmarshal(response.Body.Bytes(), &apierror.ApiErr{}), ShouldBeNil)

		vals := make(url.Values, 0)
		vals.Add("shop", "testing")
		response = httprunner.Request("GET", "/shopify/customers/1234", &vals, GetCustomers)
		So(response.Code, ShouldEqual, 500)
		So(json.Unmarshal(response.Body.Bytes(), &apierror.ApiErr{}), ShouldBeNil)
	})
	Convey("with shop identifier", t, func() {
		shopID := cart.InsertTestData()
		So(shopID, ShouldNotBeNil)

		val := shopID.Hex()
		qs := make(url.Values, 0)
		qs.Add("shop", val)

		Convey("with bad customer reference", func() {
			response = httprunner.ParameterizedRequest("GET", "/shopify/customers/:id", "/shopify/customers/1234", &qs, GetCustomer)
			So(response.Code, ShouldEqual, 500)
			So(json.Unmarshal(response.Body.Bytes(), &apierror.ApiErr{}), ShouldBeNil)
		})

		Convey("with good customer reference", func() {
			cust := cart.Customer{
				ShopId:    *shopID,
				FirstName: "Alex",
				LastName:  "Ninneman",
				Email:     "ninnemana@gmail.com",
			}
			response = httprunner.JsonRequest("POST", "/shopify/customers", &qs, cust, AddCustomer)
			So(response.Code, ShouldEqual, 200)
			So(json.Unmarshal(response.Body.Bytes(), &cust), ShouldBeNil)

			response = httprunner.ParameterizedRequest("GET", "/shopify/customers/:id", "/shopify/customers/"+cust.Id.Hex(), &qs, GetCustomer)
			So(response.Code, ShouldEqual, 200)
			So(json.Unmarshal(response.Body.Bytes(), &cust), ShouldBeNil)
		})
	})
}

func BenchmarkAddCustomer(b *testing.B) {
	shopID := cart.InsertTestData()
	if shopID == nil {
		b.Error("failed to create a shop")
		b.Fail()
	}

	val := shopID.Hex()
	qs := make(url.Values, 0)
	qs.Add("shop", val)

	cust := cart.Customer{
		ShopId:    *shopID,
		FirstName: "Alex",
		LastName:  "Ninneman",
		Email:     "ninnemana@gmail.com",
	}

	cust.Email = "ninnemana@gmail.com"
	(&httprunner.BenchmarkOptions{
		Method:             "POST",
		Route:              "/shopify/customers",
		ParameterizedRoute: "/shopify/customers",
		Handler:            AddCustomer,
		QueryString:        &qs,
		JsonBody:           cust,
		Runs:               b.N,
	}).RequestBenchmark()
}

func BenchmarkGetCustomer(b *testing.B) {
	shopID := cart.InsertTestData()
	if shopID == nil {
		b.Error("failed to create a shop")
		b.Fail()
	}

	val := shopID.Hex()
	qs := make(url.Values, 0)
	qs.Add("shop", val)

	cust := cart.Customer{
		ShopId:    *shopID,
		FirstName: "Alex",
		LastName:  "Ninneman",
		Email:     "ninnemana@gmail.com",
	}
	if err := cust.Insert(); err != nil {
		b.Error(err.Error())
		b.Fail()
	}

	(&httprunner.BenchmarkOptions{
		Method:             "GET",
		Route:              "/shopify/customers/" + cust.Id.Hex(),
		ParameterizedRoute: "/shopify/customers/:id",
		Handler:            GetCustomers,
		QueryString:        &qs,
		JsonBody:           nil,
		Runs:               b.N,
	}).RequestBenchmark()
}