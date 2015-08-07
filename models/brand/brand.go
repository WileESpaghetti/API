package brand

import (
	"database/sql"
	"errors"
	"net/url"

	"github.com/curt-labs/GoAPI/helpers/database"
	_ "github.com/go-sql-driver/mysql"
)

var (
	brandFields           = `ID, name, code, logo, logoAlt, formalName, longName, primaryColor`
	getAllBrandsStmt      = `select ` + brandFields + ` for from Brand`
	getBrandStmt          = `select ` + brandFields + ` from Brand where ID = ?`
	insertBrandStmt       = `insert into Brand(name, code) values (?,?)`
	updateBrandStmt       = `update Brand set name = ?, code = ? where ID = ?`
	deleteBrandStmt       = `delete from Brand where ID = ?`
	getCustomerUserBrands = `select b.ID, b.name, b.code, b.logo, b.logoAlt, b.formalName, b.longName, b.primaryColor
								from Brand as b
								join CustomerToBrand as ctb on ctb.BrandID = b.ID
								join Customer as c on c.cust_id = ctb.cust_id
								where c.cust_id = ?`
)

type Brands []Brand
type Brand struct {
	ID            int      `json:"id" xml:"id,attr"`
	Name          string   `json:"name" xml:"name,attr"`
	Code          string   `json:"code" xml:"code,attr"`
	Logo          *url.URL `json:"logo" xml:"logo,attr"`
	LogoAlternate *url.URL `json:"logo_alternate" xml:"logo_alternate,attr"`
	FormalName    string   `json:"formal_name" xml:"formal_name,attr"`
	LongName      string   `json:"long_name" xml:"long_name,attr"`
	PrimaryColor  string   `json:"primary_color" xml:"primary_color,attr"`
}

func GetAllBrands() (brands Brands, err error) {
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(getAllBrandsStmt)
	if err != nil {
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return
	}

	for rows.Next() {
		var b Brand
		var logo, logoAlt *string
		if err = rows.Scan(&b.ID, &b.Name, &b.Code, &logo, &logoAlt, &b.FormalName, &b.LongName, &b.PrimaryColor); err != nil {
			return
		}
		if logo != nil {
			b.Logo, _ = url.Parse(*logo)
		}
		if logoAlt != nil {
			b.LogoAlternate, _ = url.Parse(*logoAlt)
		}
		brands = append(brands, b)
	}
	defer rows.Close()

	return
}

func (b *Brand) Get() error {
	if b.ID == 0 {
		return errors.New("Invalid Brand ID")
	}

	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(getBrandStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var logo, logoAlt *string
	err = stmt.QueryRow(b.ID).Scan(&b.ID, &b.Name, &b.Code, &logo, &logoAlt, &b.FormalName, &b.LongName, &b.PrimaryColor)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Invalid Brand ID")
		}
		return err
	}

	if logo != nil {
		b.Logo, _ = url.Parse(*logo)
	}
	if logoAlt != nil {
		b.LogoAlternate, _ = url.Parse(*logoAlt)
	}

	return nil

}

func (b *Brand) Create() error {
	if b.Name == "" {
		return errors.New("Brand must have a name.")
	}
	if b.Code == "" {
		return errors.New("Brand must have a code.")
	}
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(insertBrandStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(b.Name, b.Code)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	b.ID = int(id)
	return err
}

func (b *Brand) Update() error {
	if b.Name == "" {
		return errors.New("Brand must have a name.")
	}
	if b.Code == "" {
		return errors.New("Brand must have a code.")
	}
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(updateBrandStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(b.Name, b.Code, b.ID)
	return err
}

func (b *Brand) Delete() error {
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(deleteBrandStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(b.ID)
	return err
}

func GetUserBrands(id int) ([]Brand, error) {
	brands := make([]Brand, 0)
	var err error

	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return brands, err
	}
	defer db.Close()

	stmt, err := db.Prepare(getCustomerUserBrands)
	if err != nil {
		return brands, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		return brands, err
	}

	for rows.Next() {
		var b Brand
		var logo, logoAlt *string
		if err = rows.Scan(&b.ID, &b.Name, &b.Code, &logo, &logoAlt, &b.FormalName, &b.LongName, &b.PrimaryColor); err != nil {
			continue
		}
		if logo != nil {
			b.Logo, _ = url.Parse(*logo)
		}
		if logoAlt != nil {
			b.LogoAlternate, _ = url.Parse(*logoAlt)
		}
		brands = append(brands, b)
	}

	return brands, nil
}
