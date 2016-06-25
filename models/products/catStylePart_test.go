package products

import (
	"os"
	"testing"

	"github.com/curt-labs/API/helpers/database"
	"gopkg.in/mgo.v2"
)

func TestCategoryStyleParts(t *testing.T) {
	v := NoSqlVehicle{
		Year:  "2010",
		Make:  "Chevrolet",
		Model: "Silverado 1500",
	}

	if err := database.Init(); err != nil {
		t.Error(err)
	}

	session := database.ProductMongoSession

	csp, err := CategoryStyleParts(v, []int{3}, session, true)
	if err != nil {
		t.Error(err)
	}
	t.Log(len(csp))
}
