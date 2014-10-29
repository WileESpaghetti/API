package vinLookup

import (
	"database/sql"
	"github.com/curt-labs/GoAPI/models/products"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestVinLookup(t *testing.T) {
	buickVin := "1g4ha5em2au000001"
	taurusVin := "1fahp2fw5ag100583"
	caddyVin := "1g6da5egxa0100211"
	bogusVin := "123456789"

	Convey("Testing VinPartLookup", t, func() {
		vs, err := VinPartLookup(buickVin)
		if err != sql.ErrNoRows {
			So(err, ShouldBeNil)
			So(len(vs.Parts), ShouldBeGreaterThanOrEqualTo, 1)

		}
		if err != sql.ErrNoRows {
			vs, err = VinPartLookup(taurusVin)
			So(err, ShouldBeNil)
			So(len(vs.Parts), ShouldBeGreaterThanOrEqualTo, 1)

			//Make sure it's a Taurus - VINs should be constant
			So(vs.Vehicle.Base.Model, ShouldEqual, "Taurus")

			//We have 2010 Taurus Hitches
			So(len(vs.Parts), ShouldBeGreaterThanOrEqualTo, 1)
		}

	})
	Convey("Testing GetVehicleConfigs->GetParts", t, func() {
		v, err := GetVehicleConfigs(caddyVin)
		if err != sql.ErrNoRows {
			So(err, ShouldBeNil)
			So(len(v.Configurations), ShouldBeGreaterThanOrEqualTo, 1)

			//get random vehicleConfig
			// i := rand.Intn(len(v.Configurations))

			//get parts
			// parts, err := v.
			// if err != sql.ErrNoRows {
			// 	So(err, ShouldBeNil)
			// 	So(len(parts), ShouldBeGreaterThanOrEqualTo, 1)
			// }
		}

	})
	Convey("Testing Bad Vin", t, func() {
		vs, err := VinPartLookup(bogusVin)
		So(err, ShouldNotBeNil)
		So(vs, ShouldHaveSameTypeAs, products.Lookup{})

		vcs, err := GetVehicleConfigs(bogusVin)
		So(err, ShouldNotBeNil)
		So(vcs, ShouldHaveSameTypeAs, products.Lookup{})
	})
}
