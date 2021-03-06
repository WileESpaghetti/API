package techSupport

import (
	"github.com/curt-labs/API/helpers/apicontextmock"
	"github.com/curt-labs/API/models/contact"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTechSupport(t *testing.T) {
	var err error
	var tc TechSupport
	tc.Issue = "This is an issue"
	//make contact
	tc.Contact.Email = "e@e.e"
	tc.Contact.LastName = "techSupport"
	tc.Contact.FirstName = "f"
	tc.Contact.Subject = "s"
	tc.Contact.Message = "m"

	MockedDTX, err := apicontextmock.Mock()
	if err != nil {
		return
	}

	Convey("Test Create TechSupport", t, func() {

		err = tc.Create()
		So(err, ShouldBeNil)
	})
	Convey("Test Get TechSupport", t, func() {

		err = tc.Get()
		So(err, ShouldBeNil)
	})
	Convey("Test Get TechSupport By Contact", t, func() {

		ts, err := tc.GetByContact(MockedDTX)
		So(err, ShouldBeNil)
		So(len(ts), ShouldBeGreaterThanOrEqualTo, 0)
	})
	Convey("Test Get All TechSupport", t, func() {

		allTs, err := GetAllTechSupport(MockedDTX)
		So(err, ShouldBeNil)
		So(len(allTs), ShouldBeGreaterThanOrEqualTo, 0)
	})
	Convey("Test Delete TechSupport", t, func() {

		err = tc.Delete()
		So(err, ShouldBeNil)

		err = tc.Contact.Delete()
		So(err, ShouldBeNil)
	})
	_ = apicontextmock.DeMock(MockedDTX)

}

func BenchmarkGetAllTechSupport(b *testing.B) {
	MockedDTX, err := apicontextmock.Mock()
	if err != nil {
		return
	}
	for i := 0; i < b.N; i++ {
		GetAllTechSupport(MockedDTX)
	}
	_ = apicontextmock.DeMock(MockedDTX)
}

func BenchmarkGetTechSupport(b *testing.B) {
	ts := setupDummyTechSupport()
	ts.Create()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.Get()
	}
	b.StopTimer()
	ts.Delete()
}

func BenchmarkGetTechSupportByContact(b *testing.B) {
	MockedDTX, err := apicontextmock.Mock()
	if err != nil {
		return
	}
	ts := setupDummyTechSupport()
	ts.Create()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.GetByContact(MockedDTX)
	}
	b.StopTimer()
	ts.Delete()
	_ = apicontextmock.DeMock(MockedDTX)
}

func BenchmarkCreateTechSupport(b *testing.B) {
	ts := setupDummyTechSupport()
	for i := 0; i < b.N; i++ {
		ts.Create()
		b.StopTimer()
		ts.Delete()
		b.StartTimer()
	}
}

func BenchmarkDeleteTechSupport(b *testing.B) {
	ts := setupDummyTechSupport()
	ts.Create()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.Delete()
	}
	b.StopTimer()
	ts.Delete()
}

func setupDummyTechSupport() *TechSupport {
	return &TechSupport{
		Contact: contact.Contact{
			Email:     "test@test.com",
			FirstName: "TESTER",
			LastName:  "TESTER",
			Subject:   "TESTER",
			Message:   "TESTER",
		},
		Issue: "TESTER",
	}
}
