package applicationGuide

import (
	"database/sql"
	"errors"
	"fmt"
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

var ctx = &apicontext.DataContext{
	APIKey:  "99900000-0000-0000-0000-000000000000",
	BrandID: 1,
}

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
		//ag.FileType = "pdf"
		//ag.Url = "test.com"
		//ag.Website.ID = MockedDTX.WebsiteID
		//err = ag.Create(MockedDTX)
		//So(err, ShouldBeNil)

		//get
		//err = ag.Get(MockedDTX)
		//So(err, ShouldBeNil)

		//get by site
		//ags, err := ag.GetBySite(MockedDTX)

		//So(err, ShouldBeNil)
		//So(len(ags), ShouldBeGreaterThanOrEqualTo, 1)

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
		//ag.Create(MockedDTX)
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
		//ag.Create(MockedDTX)
		//b.StartTimer()
		//ag.GetBySite(MockedDTX)
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
		//ag.Create(MockedDTX)
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

func TestApplicationGuide_GetBySite(t *testing.T) {
	var columns = []string{"id", "url", "websiteID", "fileType", "catID", "icon", "catTitle"}
	var getBySiteTests = []struct {
		name     string
		in       ApplicationGuide
		rows     *sqlmock.Rows
		queryErr error
		outErr   error
		outAgs   []ApplicationGuide
	}{
		{
			name:   "no application guides found",
			in:     ApplicationGuide{Website: site.Website{ID: 1}},
			rows:   sqlmock.NewRows(columns),
			outErr: nil,
			outAgs: []ApplicationGuide{},
		},
		{
			name:     "query failed",
			rows:     sqlmock.NewRows(columns),
			queryErr: errors.New("query error"),
			outErr:   errors.New("query error"),
			outAgs:   nil,
		},
		{
			name: "application guides found",
			rows: sqlmock.NewRows(columns).
				AddRow(1, "http://www.example.com", 1, "xls", 2, "www.example.com/icon.png", "example title").
				AddRow(2, "http://www.example.org", 1, "xls", 3, "www.example.org/icon.png", "example2 title").
				AddRow(3, "http://www.example.net", 1, "xls", 4, "www.example.net/icon.png", "example3 title"),
			outErr: nil,
			outAgs: []ApplicationGuide{
				{ID: 1, Url: "http://www.example.com", Website: site.Website{ID: 1}, FileType: "xls", Category: products.Category{CategoryID: 2, Title: "example title"}, Icon: "www.example.com/icon.png"},
				{ID: 2, Url: "http://www.example.org", Website: site.Website{ID: 1}, FileType: "xls", Category: products.Category{CategoryID: 3, Title: "example2 title"}, Icon: "www.example.org/icon.png"},
				{ID: 3, Url: "http://www.example.net", Website: site.Website{ID: 1}, FileType: "xls", Category: products.Category{CategoryID: 4, Title: "example3 title"}, Icon: "www.example.net/icon.png"},
			},
		},
		{
			name: "scan error on application guide",
			rows: sqlmock.NewRows(columns).
				AddRow(1, "http://www.example.com", 1, "xls", 2, "www.example.com/icon.png", "example title").
				AddRow(2, "http://www.example.org", 1, "xls", "asdf", "www.example.org/icon.png", "example2 title"). // causes scan error
				AddRow(3, "http://www.example.net", 1, "xls", 4, "www.example.net/icon.png", "example3 title"),
			outErr: fmt.Errorf(`sql: Scan error on column index %d, name %q: %w`, 4, columns[4], errors.New("converting driver.Value type string (\"asdf\") to a int: invalid syntax")),
			outAgs: []ApplicationGuide{
				{ID: 1, Url: "http://www.example.com", Website: site.Website{ID: 1}, FileType: "xls", Category: products.Category{CategoryID: 2, Title: "example title"}, Icon: "www.example.com/icon.png"},
			},
		},
		{
			name: "row error",
			rows: sqlmock.NewRows(columns).
				AddRow(1, "http://www.example.com", 1, "xls", 2, "www.example.com/icon.png", "example title").
				AddRow(2, "http://www.example.org", 1, "xls", "asdf", "www.example.org/icon.png", "example2 title"). // causes scan error
				AddRow(3, "http://www.example.net", 1, "xls", 4, "www.example.net/icon.png", "example3 title").
				RowError(1, errors.New("scan error")),
			outErr: errors.New("scan error"),
			outAgs: []ApplicationGuide{
				{ID: 1, Url: "http://www.example.com", Website: site.Website{ID: 1}, FileType: "xls", Category: products.Category{CategoryID: 2, Title: "example title"}, Icon: "www.example.com/icon.png"},
			},
		},
	}

	for _, tt := range getBySiteTests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := helpers.NewMock()
			defer db.Close()

			query := getApplicationGuidesBySite

			var rows *sqlmock.Rows
			if tt.rows != nil {
				rows = tt.rows
			} else {
				rows = sqlmock.NewRows(columns)
			}

			mock.ExpectQuery(query).WillReturnRows(rows).WillReturnError(tt.queryErr)

			ags, err := tt.in.GetBySite(db, ctx)

			assert.Equal(t, tt.outErr, err)
			assert.Equal(t, tt.outAgs, ags)
		})
	}
}

func TestApiKeyType_Create(t *testing.T) {
	var getTests = []struct {
		name    string
		in      *ApplicationGuide
		result  sql.Result
		execErr error
		rows    *sqlmock.Rows
		outErr  error
		outAg   *ApplicationGuide
	}{
		{
			name:    "insert API key type failed",
			in:      &ApplicationGuide{Url: "http://www.example.com", Website: site.Website{ID: 1}, FileType: "pdf", Category: products.Category{CategoryID: 2, Title: "new category"}, Icon: "http://www.example.com/icon.png"},
			execErr: errors.New("exec error"),
			outErr:  errors.New("exec error"),
			outAg:   &ApplicationGuide{Url: "http://www.example.com", Website: site.Website{ID: 1}, FileType: "pdf", Category: products.Category{CategoryID: 2, Title: "new category"}, Icon: "http://www.example.com/icon.png"},
		},
		{
			name:    "successful insert API key type",
			in:      &ApplicationGuide{Url: "http://www.example.com", Website: site.Website{ID: 1}, FileType: "pdf", Category: products.Category{CategoryID: 2, Title: "new category"}, Icon: "http://www.example.com/icon.png"},
			execErr: nil,
			outErr:  nil,
			result:  sqlmock.NewResult(1, 1),
			rows:    sqlmock.NewRows([]string{"id"}).AddRow("99990000-0000-0000-0000-000000000000"),
			outAg:   &ApplicationGuide{ID: 1, Url: "http://www.example.com", Website: site.Website{ID: 1}, FileType: "pdf", Category: products.Category{CategoryID: 2, Title: "new category"}, Icon: "http://www.example.com/icon.png"},
		},
	}

	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := helpers.NewMock()
			defer db.Close()

			query := createApplicationGuide
			mock.ExpectExec(query).
				WithArgs(tt.in.Url, tt.in.Website.ID, tt.in.FileType, tt.in.Category.CategoryID, tt.in.Icon, ctx.BrandID).
				WillReturnError(tt.execErr).
				WillReturnResult(tt.result)

			err := tt.in.Create(db, ctx)
			assert.Equal(t, tt.outErr, err)
			assert.Equal(t, tt.outAg, tt.in)
		})
	}
}
