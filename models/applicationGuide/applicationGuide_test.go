package applicationGuide

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bmizerany/assert"
	"github.com/curt-labs/API/helpers"
	"github.com/curt-labs/API/helpers/apicontext"
	"github.com/curt-labs/API/helpers/apicontextmock"
	"github.com/curt-labs/API/models/products"
	"github.com/curt-labs/API/models/site"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAppGuides(t *testing.T) {
	MockedDTX := &apicontext.DataContext{}
	var err error
	if MockedDTX, err = apicontextmock.Mock(); err != nil {
		return
	}
	Convey("Test Create AppGuide", t, func() {
		var err error
		var ag ApplicationGuide

		//create
		ag.FileType = "pdf"
		ag.Url = "test.com"
		ag.Website.ID = MockedDTX.WebsiteID
		err = ag.Create(MockedDTX)
		So(err, ShouldBeNil)

		//get
		//err = ag.Get(MockedDTX)
		//So(err, ShouldBeNil)

		//get by site
		ags, err := ag.GetBySite(MockedDTX)

		So(err, ShouldBeNil)
		So(len(ags), ShouldBeGreaterThanOrEqualTo, 1)

		//delete
		err = ag.Delete()
		So(err, ShouldBeNil)

	})
	_ = apicontextmock.DeMock(MockedDTX)
}

func BenchmarkGetAppGuide(b *testing.B) {
	MockedDTX := &apicontext.DataContext{}
	var err error
	if MockedDTX, err = apicontextmock.Mock(); err != nil {
		return
	}
	var ag ApplicationGuide
	ag.FileType = "pdf"
	ag.Url = "http://google.com"
	ag.Website.ID = 1

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ag.Create(MockedDTX)
		//b.StartTimer()
		//ag.Get(MockedDTX)
		//b.StopTimer()
		ag.Delete()
	}
	_ = apicontextmock.DeMock(MockedDTX)
}

func BenchmarkGetBySite(b *testing.B) {
	MockedDTX := &apicontext.DataContext{}
	var err error
	if MockedDTX, err = apicontextmock.Mock(); err != nil {
		return
	}
	var ag ApplicationGuide
	ag.FileType = "pdf"
	ag.Url = "http://google.com"
	ag.Website.ID = 1

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ag.Create(MockedDTX)
		b.StartTimer()
		ag.GetBySite(MockedDTX)
		b.StopTimer()
		ag.Delete()
	}
	_ = apicontextmock.DeMock(MockedDTX)
}

func BenchmarkDeleteAppGuide(b *testing.B) {
	MockedDTX := &apicontext.DataContext{}
	var err error
	if MockedDTX, err = apicontextmock.Mock(); err != nil {
		return
	}
	var ag ApplicationGuide
	ag.FileType = "pdf"
	ag.Url = "http://google.com"
	ag.Website.ID = 1

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ag.Create(MockedDTX)
		b.StartTimer()
		ag.Delete()
	}
	_ = apicontextmock.DeMock(MockedDTX)
}

func TestApplicationGuide_Get(t *testing.T) {
	var columns = []string{"id", "url", "websiteID", "fileType", "catID", "icon", "catTitle"}
	var getTests = []struct {
		name   string
		in     *ApplicationGuide
		rows   *sqlmock.Rows
		outErr error
		outAg  *ApplicationGuide
	}{
		{
			name:   "no application guide found",
			in:     &ApplicationGuide{ID: 1},
			outErr: sql.ErrNoRows,
			outAg:  &ApplicationGuide{ID: 1},
		},
		{
			name: "application guide found",
			in:   &ApplicationGuide{ID: 1},
			rows: sqlmock.NewRows(columns).
				AddRow(1, "http://www.example.com", 2, "xls", 3, "www.example.com/icon.png", "example category"),
			outErr: nil,
			outAg:  &ApplicationGuide{ID: 1, Url: "http://www.example.com", Website: site.Website{ID: 2}, FileType: "xls", Category: products.Category{CategoryID: 3, Title: "example category"}, Icon: "www.example.com/icon.png"},
		},
	}

	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := helpers.NewMock()
			defer db.Close()

			query := getApplicationGuide

			var rows *sqlmock.Rows
			if tt.rows != nil {
				rows = tt.rows
			} else {
				rows = sqlmock.NewRows(columns)
			}

			mock.ExpectQuery(query).WithArgs(tt.in.ID).WillReturnRows(rows)

			err := tt.in.Get(db)
			assert.Equal(t, tt.outErr, err)
			assert.Equal(t, tt.outAg, tt.in)
		})
	}
}
