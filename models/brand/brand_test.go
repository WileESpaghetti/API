package brand

import (
	"github.com/curt-labs/API/helpers/apicontextmock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestBrands(t *testing.T) {
	var err error
	b := setupDummyBrand()
	dtx, err := apicontextmock.Mock()
	if err != nil {
		return
	}

	Convey("Testing GetAll", t, func() {
		brands, err := GetAllBrands()
		So(err, ShouldBeNil)
		So(len(brands), ShouldBeGreaterThanOrEqualTo, 0)
	})

	Convey("Testing Brands - CRUD", t, func() {
		Convey("Testing Create", func() {
			err = b.Create()
			So(err, ShouldBeNil)
			So(b.ID, ShouldNotEqual, 0)

			err = b.Get()
			So(err, ShouldBeNil)
			So(b.ID, ShouldBeGreaterThan, 0)
			So(b.Name, ShouldEqual, "TESTER")

			b.Name = "TESTING"
			err = b.Update()
			So(err, ShouldBeNil)
			So(b.Name, ShouldEqual, "TESTING")

			sites, err := getWebsites(b.ID)
			So(err, ShouldBeNil)
			So(sites, ShouldHaveSameTypeAs, []Website{})

			brands, err := GetUserBrands(dtx.CustomerID)
			So(err, ShouldBeNil)
			So(brands, ShouldHaveSameTypeAs, []Brand{})

			err = b.Delete()
			So(err, ShouldBeNil)

		})
		Convey("Testing Get - Bad ID", func() {
			br := Brand{}
			err = br.Get()
			So(err, ShouldNotBeNil)
		})
	})
	apicontextmock.DeMock(dtx)

}

func BenchmarkGetAllBrands(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetAllBrands()
	}
}

func BenchmarkGetBrand(b *testing.B) {
	br := setupDummyBrand()
	br.Create()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		br.Get()
	}
	b.StopTimer()
	br.Delete()
}

func BenchmarkCreateBrand(b *testing.B) {
	br := setupDummyBrand()
	for i := 0; i < b.N; i++ {
		br.Create()
		b.StopTimer()
		br.Delete()
		b.StartTimer()
	}
}

func BenchmarkUpdateBrand(b *testing.B) {
	br := setupDummyBrand()
	br.Create()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		br.Name = "TESTING"
		br.Code = "TEST"
		br.Update()
	}
	b.StopTimer()
	br.Delete()
}

func BenchmarkDeleteBrand(b *testing.B) {
	br := setupDummyBrand()
	br.Create()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		br.Delete()
	}
	b.StopTimer()
	br.Delete()
}

func setupDummyBrand() *Brand {
	return &Brand{
		Name: "TESTER",
		Code: "TESTER",
	}
}
