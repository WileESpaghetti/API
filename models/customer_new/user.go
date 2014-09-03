package customer_new

import (
	"code.google.com/p/go.crypto/bcrypt"
	"database/sql"
	"errors"
	"fmt"
	"github.com/curt-labs/GoAPI/helpers/api"
	"github.com/curt-labs/GoAPI/helpers/database"
	"github.com/curt-labs/GoAPI/helpers/redis"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CustomerUser struct {
	Id                    string
	Name, Email           string
	DateAdded             time.Time
	Active, Sudo, Current bool
	Location              CustomerLocation
	Keys                  []ApiCredentials
}

type ApiCredentials struct {
	Key, Type, TypeId string
	DateAdded         time.Time
}

var (
	userCustomer = `select c.customerID, c.name, c.email, c.address, c.address2, c.city, c.phone, c.fax, c.contact_person,
						c.latitude, c.longitude, c.searchURL, c.logo, c.website,
						c.postal_code, s.stateID, s.state, s.abbr as state_abbr, cty.countryID, cty.name as country_name, cty.abbr as country_abbr,
						dt.dealer_type as typeID, dt.type as dealerType, dt.online as typeOnline, dt.show as typeShow, dt.label as typeLabel,
						dtr.ID as tierID, dtr.tier as tier, dtr.sort as tierSort,
						mi.ID as iconID, mi.mapicon, mi.mapiconshadow,
						mpx.code as mapix_code, mpx.description as mapic_desc,
						sr.name as rep_name, sr.code as rep_code, c.parentID
						from Customer as c
						join CustomerUser as cu on c.cust_id = cu.cust_ID
						left join States as s on c.stateID = s.stateID
						left join Country as cty on s.countryID = cty.countryID
						left join DealerTypes as dt on c.dealer_type = dt.dealer_type
						left join MapIcons as mi on dt.dealer_type = mi.dealer_type
						left join DealerTiers dtr on c.tier = dtr.ID
						left join MapixCode as mpx on c.mCodeID = mpx.mCodeID
						left join SalesRepresentative as sr on c.salesRepID = sr.salesRepID
						where cu.id = ?`

	customerUserAuth = `select password, id, name, email, date_added, active, isSudo, passwordConverted from CustomerUser
							where email = ?
							&& active = 1
							limit 1`
	updateCustomerUserPass = `update CustomerUser set password = ?, passwordConverted = 1
								where id = ? && active = 1`
	customerUserKeyAuth = `select cu.* from CustomerUser as cu
								join ApiKey as ak on cu.id = ak.user_id
								join ApiKeyType as akt on ak.type_id = akt.id
								where UPPER(akt.type) = ?
								&& ak.api_key = ?
								&& cu.active = 1 && ak.date_added >= ?`
	customerUserKeys = `select ak.api_key, akt.type, ak.date_added from ApiKey as ak
								join ApiKeyType as akt on ak.type_id = akt.id
								where user_id = ? && UPPER(akt.type) NOT IN (?)`
	userLocation = `select cl.locationID, cl.name, cl.email, cl.address, cl.city,
						cl.postalCode, cl.phone, cl.fax, cl.latitude, cl.longitude,
						cl.cust_id, cl.contact_person, cl.isprimary, cl.ShippingDefault,
						s.stateID, s.state, s.abbr as state_abbr, cty.countryID, cty.name as cty_name, cty.abbr as cty_abbr
						from CustomerLocations as cl
						left join States as s on cl.stateID = s.stateID
						left join Country as cty on s.countryID = cty.countryID
						join CustomerUser as cu on cl.locationID = cu.locationID
						where cu.id = ?`

	userAuthenticationKey = `select ak.api_key, akt.type, akt.id, ak.date_added from ApiKey as ak
									join ApiKeyType as akt on ak.type_id = akt.id
									where UPPER(akt.type) = ?
									&& ak.user_id = ?`

	resetUserAuthentication = `update ApiKey as ak
									set ak.date_added = ?
									where ak.type_id = ?
									&& ak.user_id = ?`
	customerIDFromKey = `select c.customerID from Customer as c
								join CustomerUser as cu on c.cust_id = cu.cust_ID
								join ApiKey as ak on cu.id = ak.user_id
								where ak.api_key = ?
								limit 1`
	customerUserFromKey = `select cu.* from CustomerUser as cu
								join ApiKey as ak on cu.id = ak.user_id
								join ApiKeyType as akt on ak.type_id = akt.id
								where akt.type != ? && ak.api_key = ?
								limit 1`

	customerUserFromId = `select cu.* from CustomerUser as cu
							join ApiKey as ak on cu.id = ak.user_id
							join ApiKeyType as akt on ak.type_id = akt.id
							where cu.id = ?
							limit 1`
)

func (u CustomerUser) UserAuthentication(password string) (cust Customer, err error) {

	err = u.AuthenticateUser(password)
	if err != nil {
		return
	}

	keyChan := make(chan int)
	locChan := make(chan int)

	go func() {
		if kErr := u.GetKeys(); kErr != nil {
			err = kErr
		}
		keyChan <- 1
	}()

	go func() {
		if lErr := u.GetLocation(); lErr != nil {
			err = lErr
		}
		locChan <- 1
	}()

	cust, err = u.GetCustomer()

	<-keyChan
	<-locChan

	cust.Users = append(cust.Users, u)

	return
}

func UserAuthenticationByKey(key string) (cust Customer, err error) {
	u, err := AuthenticateUserByKey(key)
	if err != nil {
		return
	}

	keyChan := make(chan int)
	locChan := make(chan int)

	go func() {
		if kErr := u.GetKeys(); kErr != nil {
			err = kErr
		}
		keyChan <- 1
	}()

	go func() {
		if lErr := u.GetLocation(); lErr != nil {
			err = lErr
		}
		locChan <- 1
	}()

	cust, err = u.GetCustomer()

	<-keyChan
	<-locChan

	cust.Users = append(cust.Users, u)

	return
}

func (u CustomerUser) GetCustomer() (c Customer, err error) {
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return c, err
	}
	defer db.Close()

	stmt, err := db.Prepare(userCustomer)
	if err != nil {
		return c, err
	}
	defer stmt.Close()

	var logo, web, lat, lon, url, icon, shadow, mapIconId []byte
	var stateId, state, stateAbbr, countryId, country, countryAbbr, parentId, postalCode, mapixCode, mapixDesc, rep, repCode []byte
	err = stmt.QueryRow(u.Id).Scan(
		&c.Id,            //c.customerID,
		&c.Name,          //c.name
		&c.Email,         //c.email
		&c.Address,       //c.address
		&c.Address2,      //c.address2
		&c.City,          //c.city,
		&c.Phone,         //phone,
		&c.Fax,           //c.fax
		&c.ContactPerson, //c.contact_person,
		&lat,             //c.latitude
		&lon,             //c.longitude
		&url,
		&logo,
		&web,
		&postalCode,          //c.postal_code
		&stateId,             //s.stateID
		&state,               //s.state
		&stateAbbr,           //s.abbr as state_abbr
		&countryId,           //cty.countryID,
		&country,             //cty.name as country_name
		&countryAbbr,         //cty.abbr as country_abbr,
		&c.DealerType.Id,     //dt.dealer_type as typeID
		&c.DealerType.Type,   // dt.type as dealerType
		&c.DealerType.Online, // dt.online as typeOnline,
		&c.DealerType.Show,   //dt.show as typeShow
		&c.DealerType.Label,  //dt.label as typeLabel,
		&c.DealerTier.Id,     //dtr.ID as tierID,
		&c.DealerTier.Tier,   //dtr.tier as tier
		&c.DealerTier.Sort,   //dtr.sort as tierSort
		&mapIconId,
		&icon,
		&shadow,    //mi.ID as iconID
		&mapixCode, //mpx.code as mapix_code
		&mapixDesc, //mpx.description as mapic_desc,
		&rep,       //sr.name as rep_name
		&repCode,   // sr.code as rep_code,
		&parentId,  //c.parentID
	)
	if err != nil {
		return c, err
	}
	c.Latitude, err = byteToFloat(lat)
	c.Longitude, err = byteToFloat(lon)
	c.SearchUrl, err = byteToUrl(url)
	c.Logo, err = byteToUrl(logo)
	c.Website, err = byteToUrl(web)
	c.DealerType.MapIcon.MapIcon, err = byteToUrl(icon)
	c.DealerType.MapIcon.MapIconShadow, err = byteToUrl(shadow)
	c.PostalCode, err = byteToString(postalCode)
	c.State.Id, err = byteToInt(stateId)
	c.State.State, err = byteToString(state)
	c.State.Abbreviation, err = byteToString(stateAbbr)
	c.State.Country.Id, err = byteToInt(countryId)
	c.State.Country.Country, err = byteToString(country)
	c.State.Country.Abbreviation, err = byteToString(countryAbbr)
	c.DealerType.MapIcon.Id, err = byteToInt(mapIconId)
	c.DealerType.MapIcon.MapIcon, err = byteToUrl(icon)
	c.DealerType.MapIcon.MapIconShadow, err = byteToUrl(shadow)
	c.MapixCode, err = byteToString(mapixCode)
	c.MapixDescription, err = byteToString(mapixDesc)
	c.SalesRepresentative, err = byteToString(rep)
	c.SalesRepresentativeCode, err = byteToString(repCode)

	parentInt, err := byteToInt(parentId)
	if err != nil {
		return c, err
	}
	if parentInt != 0 {
		par := Customer{Id: parentInt}
		par.GetCustomer()
		c.Parent = &par
	}
	return
}

//TODO - does this method work the way the original author wanted it? Seems to reset a password when there is not a match. Odd.
func (u *CustomerUser) AuthenticateUser(pass string) error {

	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(customerUserAuth)
	if err != nil {
		return err
	}
	defer stmt.Close()
	var dbPass string
	var passConversion bool
	err = stmt.QueryRow(u.Email).Scan(
		&dbPass,
		&u.Id,
		&u.Name,
		&u.Email,
		&u.DateAdded,
		&u.Active,
		&u.Sudo,
		&passConversion,
	)
	if err == nil {
		err = errors.New("No user found that matches: " + u.Email)
	}

	// Attempt to compare bcrypt strings
	if bcrypt.CompareHashAndPassword([]byte(dbPass), []byte(pass)) != nil {
		// Compare unsuccessful
		enc_pass, err := api_helpers.Md5Encrypt(pass)
		if err != nil {
			return err
		}
		if len(enc_pass) != len(dbPass) || passConversion { //bool
			return errors.New("Invalid password")
		}

		hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			return errors.New("Failed to encode the password")
		}

		stmtPass, err := db.Prepare(updateCustomerUserPass)
		if err != nil {
			return err
		}
		_, err = stmtPass.Exec(hashedPass, u.Id)
	}

	resetChan := make(chan int)
	go func() {
		if resetErr := u.ResetAuthentication(); resetErr != nil {
			err = resetErr
		}
		resetChan <- 1
	}()

	<-resetChan
	return nil
}

func AuthenticateUserByKey(key string) (u CustomerUser, err error) {
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return u, err
	}
	defer db.Close()

	stmt, err := db.Prepare(customerUserKeyAuth)
	if err != nil {
		return u, err
	}
	defer stmt.Close()
	t := time.Now()
	t1 := t.Add(time.Duration(-6) * time.Hour) //6 hours ago
	Timer := t1.String()
	KeyType := api_helpers.AUTH_KEY_TYPE
	params := []interface{}{
		KeyType,
		key,
		Timer,
	}
	var dbPass, custId, customerId string
	var passConversion, notCustomer []byte //bools
	err = stmt.QueryRow(params...).Scan(
		&u.Id,
		&u.Name,
		&u.Email,
		&dbPass,     //Not Used
		&customerId, //Not Used
		&u.DateAdded,
		&u.Active,
		&u.Location.Id,
		&u.Sudo,
		&custId,         //Not Used
		&notCustomer,    //Not Used
		&passConversion, //Not Used
	)
	if err != nil {
		return u, err
		// err = errors.New("Invalid password")
	}

	// DISABLED: See RenewAuthentication() below
	//
	// resetChan := make(chan int)
	// go func() {
	// 	if resetErr := u.RenewAuthentication(); resetErr != nil {
	// 		err = resetErr
	// 	}
	// 	resetChan <- 1
	// }()
	return
}

func (u *CustomerUser) GetKeys() error {
	//ak.api_key, akt.type, ak.date_added
	var keys []ApiCredentials
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(customerUserKeys)
	if err != nil {
		return err
	}
	defer stmt.Close()

	params := []interface{}{
		u.Id,
		strings.Join([]string{api_helpers.AUTH_KEY_TYPE}, ","),
	}
	res, err := stmt.Query(params...)
	for res.Next() {
		var a ApiCredentials
		res.Scan(&a.Key, &a.Type, &a.DateAdded)
		keys = append(keys, a)
	}
	u.Keys = keys
	return nil
}

func (u *CustomerUser) GetLocation() error {
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(userLocation)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(u.Id).Scan(
		&u.Location.Id,
		&u.Name,
		&u.Email,
		&u.Location.Address,
		&u.Location.City,
		&u.Location.PostalCode,
		&u.Location.Phone,
		&u.Location.Fax,
		&u.Location.Latitude,
		&u.Location.Longitude,
		&u.Location.CustomerId,
		&u.Location.ContactPerson,
		&u.Location.IsPrimary,
		&u.Location.ShippingDefault,
		&u.Location.State.Id,
		&u.Location.State.State,
		&u.Location.State.Abbreviation,
		&u.Location.State.Country.Id,
		&u.Location.State.Country.Country,
		&u.Location.State.Country.Abbreviation,
	)
	if err != nil {
		return err
	}
	return nil
}

func (u *CustomerUser) ResetAuthentication() error {
	var err error
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(userAuthenticationKey)
	if err != nil {
		return err
	}
	defer stmt.Close()

	params := []interface{}{
		api_helpers.AUTH_KEY_TYPE,
		u.Id,
	}
	var a ApiCredentials
	err = stmt.QueryRow(params...).Scan(&a.Key, &a.Type, &a.TypeId, &a.DateAdded)
	if err != nil {
		return err
	} else {
		paramsNew := []interface{}{
			time.Now().String(),
			a.TypeId,
			u.Id,
		}

		stmtNew, err := db.Prepare(resetUserAuthentication)
		_, err = stmtNew.Exec(paramsNew...)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetCustomerIdFromKey(key string) (id int, err error) {
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return id, err
	}
	defer db.Close()

	stmt, err := db.Prepare(customerIDFromKey)
	if err != nil {
		return id, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(key).Scan(&id)
	if err != nil {
		return id, err
	}
	return id, err
}

func GetCustomerUserFromKey(key string) (u CustomerUser, err error) {
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return u, err
	}
	defer db.Close()

	stmt, err := db.Prepare(customerUserFromKey)
	if err != nil {
		return u, err
	}
	defer stmt.Close()

	params := []interface{}{
		api_helpers.AUTH_KEY_TYPE,
		key,
	}
	var dbPass, custId, notCustomer, passConversion, customerId string
	err = stmt.QueryRow(params...).Scan(
		&u.Id,
		&u.Name,
		&u.Email,
		&dbPass, //Not Used
		&custId, //Not Used
		&u.DateAdded,
		&u.Active,
		&u.Location.Id,
		&u.Sudo,
		&customerId,     //Not Used
		&notCustomer,    //Not Used
		&passConversion, //Not Used
	)
	if err != nil {
		err = errors.New("Invalid key")
		return
	}
	return
}

func GetCustomerUserById(id string) (u CustomerUser, err error) {
	db, err := sql.Open("mysql", database.ConnectionString())
	if err != nil {
		return u, err
	}
	defer db.Close()

	stmt, err := db.Prepare(customerUserFromId)
	if err != nil {
		return u, err
	}
	defer stmt.Close()

	var dbPass, custId, customerId, notCustomer, passConversion string
	err = stmt.QueryRow(id).Scan(
		&u.Id,
		&u.Name,
		&u.Email,
		&dbPass, //Not Used
		&custId, //Not User
		&u.DateAdded,
		&u.Active,
		&u.Location.Id,
		&u.Sudo,
		&customerId,     //Not Used
		&notCustomer,    //Not Used
		&passConversion, //Not Used
	)
	if err != nil {
		return u, err
		// err = errors.New("Invalid key")
		return
	}
	return
}

type ApiRequest struct {
	User        CustomerUser
	RequestTime time.Time
	Url         *url.URL
	Query       url.Values
	Form        url.Values
}

func (u *CustomerUser) LogApiRequest(r *http.Request) {
	var ar ApiRequest
	ar.User = *u
	ar.RequestTime = time.Now()
	ar.Url = r.URL
	ar.Query = r.URL.Query()
	ar.Form = r.Form

	redis.Lpush(fmt.Sprintf("log:%s", u.Id), ar)
}

// The disabling of the triggers is failing in this method.
//
// I'm going to disable the call to it completely and expand
// the time limit of the authentication key to 6 hours.
//
// TODO: This will need to be fixed at some point in time. **Important

// func (u *CustomerUser) RenewAuthentication() error {
// 	log.Println("renewing authentication key")
// 	t := time.Now()

// 	log.Printf(renewUserAuthenticationStmt, t.String(), AUTH_KEY_TYPE, u.Id)

// 	// Excecute the update statement
// 	_, _, err := database.Db.Query(disableTriggerStmt)
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}
// 	_, _, err = database.Db.Query(renewUserAuthenticationStmt, t.String(), AUTH_KEY_TYPE, u.Id)
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}
// 	_, _, err = database.Db.Query(enableTriggerStmt)
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}
// 	return nil
// }