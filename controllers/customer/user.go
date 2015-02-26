package customer_ctlr

import (
	"github.com/curt-labs/GoAPI/helpers/apicontext"
	"github.com/curt-labs/GoAPI/helpers/encoding"
	"github.com/curt-labs/GoAPI/helpers/error"
	"github.com/curt-labs/GoAPI/models/customer"
	"github.com/go-martini/martini"

	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

//Post - Form Authentication
func AuthenticateUser(rw http.ResponseWriter, r *http.Request, enc encoding.Encoder) string {
	var err error
	user := customer.CustomerUser{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err = user.AuthenticateUser(); err != nil {
		apierror.GenerateError("Trouble authenticating customer user", err, rw, r)
		return ""
	}

	if err = user.GetLocation(); err != nil {
		apierror.GenerateError("Trouble getting customer user location", err, rw, r)
		return ""
	}

	if err = user.GetKeys(); err != nil {
		apierror.GenerateError("Trouble getting customer user API keys", err, rw, r)
		return ""
	}

	var key string
	if len(user.Keys) != 0 {
		key = user.Keys[0].Key
	}

	cust, err := user.GetCustomer(key)
	if err != nil {
		apierror.GenerateError("Trouble getting customer user", err, rw, r)
		return ""
	}

	return encoding.Must(enc.Encode(cust))
}

//Get - Key (in params) Authentication
func KeyedUserAuthentication(rw http.ResponseWriter, r *http.Request, enc encoding.Encoder) string {
	var err error

	qs := r.URL.Query()
	key := qs.Get("key")

	dtx := &apicontext.DataContext{APIKey: key}
	if dtx.BrandArray, err = dtx.GetBrandsFromKey(); err != nil {
		apierror.GenerateError("Trouble getting brands from API key", err, rw, r)
		return ""
	}

	cust, err := customer.AuthenticateAndGetCustomer(key, dtx)
	if err != nil {
		apierror.GenerateError("Trouble getting customer while authenticating", err, rw, r)
		return ""
	}

	return encoding.Must(enc.Encode(cust))
}

func GetUserById(rw http.ResponseWriter, r *http.Request, enc encoding.Encoder, params martini.Params) string {
	var err error
	var user customer.CustomerUser

	qs := r.URL.Query()
	key := qs.Get("key")

	if params["id"] != "" {
		user.Id = params["id"]
	} else if r.FormValue("id") != "" {
		user.Id = r.FormValue("id")
	} else {
		err = errors.New("Trouble getting customer user ID")
		apierror.GenerateError("Trouble getting customer user ID", err, rw, r)
		return ""
	}

	if err = user.Get(key); err != nil {
		apierror.GenerateError("Trouble getting customer user", err, rw, r)
		return ""
	}

	return encoding.Must(enc.Encode(user))
}

func ResetPassword(rw http.ResponseWriter, r *http.Request, enc encoding.Encoder) string {
	var err error

	email := r.FormValue("email")
	custID := r.FormValue("customerID")

	if email == "" {
		err = errors.New("No email address provided")
		apierror.GenerateError("No email address provided", err, rw, r)
		return ""
	}
	if custID == "" {
		err = errors.New("Customer ID cannot be blank")
		apierror.GenerateError("Customer ID cannot be blank", err, rw, r)
		return ""
	}

	var user customer.CustomerUser
	user.Email = email

	resp, err := user.ResetPass()
	if err != nil || resp == "" {
		apierror.GenerateError("Trouble resetting user password", err, rw, r)
		return ""
	}

	return encoding.Must(enc.Encode(resp))
}

func ChangePassword(rw http.ResponseWriter, r *http.Request, enc encoding.Encoder, dtx *apicontext.DataContext) string {
	user := customer.CustomerUser{
		Email: r.FormValue("email"),
	}

	oldPass := r.FormValue("oldPass")
	newPass := r.FormValue("newPass")

	if err := user.ChangePass(oldPass, newPass, dtx); err != nil {
		apierror.GenerateError("Trouble changing password for customer user", err, rw, r)
		return ""
	}

	return encoding.Must(enc.Encode("Success"))
}

func GenerateApiKey(rw http.ResponseWriter, r *http.Request, params martini.Params, enc encoding.Encoder, dtx *apicontext.DataContext) string {
	var err error
	qs := r.URL.Query()
	key := qs.Get("key")
	if key == "" {
		key = r.FormValue("key")
	}

	user, err := customer.GetCustomerUserFromKey(key)
	if err != nil || user.Id == "" {
		apierror.GenerateError("Trouble getting customer user using this api key", err, rw, r)
		return ""
	}

	authed := false
	if user.Sudo == false {
		for _, k := range user.Keys {
			if k.Type == customer.PRIVATE_KEY_TYPE && k.Key == key {
				authed = true
				break
			}
		}
	} else {
		authed = true
	}

	if !authed {
		err = errors.New("You do not have sufficient permissions to perform this operation.")
		apierror.GenerateError("Unauthorized", err, rw, r, http.StatusUnauthorized)
		return ""
	}

	id := params["id"]
	if r.FormValue("id") != "" {
		id = r.FormValue("id")
	}

	generateType := params["type"]
	if r.FormValue("type") != "" {
		generateType = r.FormValue("type")
	}

	if id == "" {
		err = errors.New("You must provide a reference to the user whose key should be generated.")
		apierror.GenerateError("Invalid user reference", err, rw, r)
		return ""
	}
	if generateType == "" {
		err = errors.New("You must provide the type of key to be generated")
		apierror.GenerateError("Invalid API key type", err, rw, r)
		return ""
	}

	user.Id = id
	if err = user.Get(key); err != nil {
		apierror.GenerateError("Invalid user reference", err, rw, r)
		return ""
	}

	generated, err := user.GenerateAPIKey(generateType, dtx.BrandArray)
	if err != nil {
		err = errors.New("Failed to generate an API Key: ")
		apierror.GenerateError("Failed to generate an API Key", err, rw, r)
		return ""
	}

	return encoding.Must(enc.Encode(generated))
}

//a/k/a CreateUser
func RegisterUser(rw http.ResponseWriter, r *http.Request, enc encoding.Encoder) string {
	var err error

	name := r.FormValue("name")
	email := r.FormValue("email")
	pass := r.FormValue("pass")
	customerID, err := strconv.Atoi(r.FormValue("customerID"))
	isActive, err := strconv.ParseBool(r.FormValue("isActive"))
	locationID, err := strconv.Atoi(r.FormValue("locationID"))
	isSudo, err := strconv.ParseBool(r.FormValue("isSudo"))
	cust_ID, err := strconv.Atoi(r.FormValue("cust_ID"))
	notCustomer, err := strconv.ParseBool(r.FormValue("notCustomer"))

	if email == "" || pass == "" {
		err = errors.New("Email and password are required.")
		apierror.GenerateError("Email and password are required", err, rw, r)
		return ""
	}

	var user customer.CustomerUser
	user.Email = email
	user.Password = pass
	if name != "" {
		user.Name = name
	}
	if customerID != 0 {
		user.OldCustomerID = customerID
	}
	if locationID != 0 {
		user.Location.Id = locationID
	}
	if cust_ID != 0 {
		user.CustomerID = cust_ID
	}
	user.Active = isActive
	user.Sudo = isSudo
	user.Current = notCustomer

	if err = user.GetBrands(); err != nil {
		apierror.GenerateError("Trouble getting user brands.", err, rw, r)
		return ""
	}
	var brandIds []int
	for _, brand := range user.Brands {
		brandIds = append(brandIds, brand.ID)
	}

	// cu, err := user.Register(pass, customerID, isActive, locationID, isSudo, cust_ID, notCustomer)
	if err = user.Create(brandIds); err != nil {
		apierror.GenerateError("Trouble registering new customer user", err, rw, r)
		return ""
	}

	return encoding.Must(enc.Encode(user))
}

func DeleteCustomerUser(rw http.ResponseWriter, r *http.Request, enc encoding.Encoder, params martini.Params) string {
	var err error
	var cu customer.CustomerUser

	if params["id"] != "" {
		cu.Id = params["id"]
	} else if r.FormValue("id") != "" {
		cu.Id = r.FormValue("id")
	} else {
		err = errors.New("Trouble getting customer user ID")
		apierror.GenerateError("Trouble getting customer user ID", err, rw, r)
		return ""
	}

	if err = cu.Delete(); err != nil {
		apierror.GenerateError("Trouble deleting customer user", err, rw, r)
		return ""
	}

	return encoding.Must(enc.Encode(cu))
}

func DeleteCustomerUsersByCustomerID(rw http.ResponseWriter, r *http.Request, enc encoding.Encoder, params martini.Params) string {
	var err error
	var customerID int

	id := params["id"]
	if id == "" {
		id = r.FormValue("id")
	}

	if customerID, err = strconv.Atoi(id); err != nil {
		apierror.GenerateError("Trouble getting customer ID", err, rw, r)
		return ""
	}

	if err = customer.DeleteCustomerUsersByCustomerID(customerID); err != nil {
		apierror.GenerateError("Trouble deleting customer users", err, rw, r)
		return ""
	}

	return encoding.Must(enc.Encode("Success."))
}

func UpdateCustomerUser(rw http.ResponseWriter, r *http.Request, enc encoding.Encoder, params martini.Params) string {
	var err error
	var cu customer.CustomerUser

	qs := r.URL.Query()
	key := qs.Get("key")

	if params["id"] != "" {
		cu.Id = params["id"]
	} else if r.FormValue("id") != "" {
		cu.Id = r.FormValue("id")
	} else {
		err = errors.New("Trouble getting customer user ID")
		apierror.GenerateError("Trouble getting customer user ID", err, rw, r)
		return ""
	}

	if err = cu.Get(key); err != nil {
		apierror.GenerateError("Trouble getting customer user", err, rw, r)
		return ""
	}

	if strings.ToLower(r.Header.Get("Content-Type")) == "application/json" {
		var data []byte
		if data, err = ioutil.ReadAll(r.Body); err != nil {
			apierror.GenerateError("Trouble reading request body while updating customer user", err, rw, r)
			return ""
		}
		if err = json.Unmarshal(data, &cu); err != nil {
			apierror.GenerateError("Trouble unmarshalling json request body while updating customer user", err, rw, r)
			return ""
		}
	} else {
		name := r.FormValue("name")
		email := r.FormValue("email")
		isActive := r.FormValue("isActive")
		locationID := r.FormValue("locationID")
		isSudo := r.FormValue("isSudo")
		notCustomer := r.FormValue("notCustomer")
		if name != "" {
			cu.Name = name
		}
		if email != "" {
			cu.Email = email
		}
		if isActive != "" {
			if cu.Active, err = strconv.ParseBool(isActive); err != nil {
				cu.Active = false
			}
		}
		if locationID != "" {
			if cu.Location.Id, err = strconv.Atoi(locationID); err != nil {
				apierror.GenerateError("Trouble getting location ID", err, rw, r)
				return ""
			}
		}
		if isSudo != "" {
			if cu.Sudo, err = strconv.ParseBool(isSudo); err != nil {
				cu.Sudo = false
			}
		}
		if notCustomer != "" {
			if cu.NotCustomer, err = strconv.ParseBool(notCustomer); err != nil {
				cu.NotCustomer = false
			}
		}
	}

	if err = cu.UpdateCustomerUser(); err != nil {
		apierror.GenerateError("Trouble updating customer user", err, rw, r)
		return ""
	}

	return encoding.Must(enc.Encode(cu))
}
