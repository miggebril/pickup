package models

import (
	"gopkg.in/mgo.v2/bson"
	"golang.org/x/crypto/bcrypt"
	"encoding/json"
	"encoding/base64"
	"fmt"
)

type TokenAuthentication struct {
	Token string `json:"token" form:"token"`
}

type User struct {
	ID       bson.ObjectId `json:"-" bson:"_id,omitempty" col:"users"` 
	Email string
	Username string 			
	Password []byte			`json:"-"`

	Verified bool
}

func (u *User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		ID 				string `json:"ID"`
		*Alias
	}{
		ID: u.GetIDEncoded(),
		Alias: (*Alias)(u),
	})
}

func (u User) GetIDEncoded() string {
	return base64.URLEncoding.EncodeToString([]byte(u.ID))
}

//SetPassword takes a plaintext password and hashes it with bcrypt and sets the
//password field to the hash.
func (u *User) SetPassword(password string) {
	hpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	fmt.Println(hpass)
	if err != nil {
		panic(err) //this is a panic because bcrypt errors on invalid costs
	}
	u.Password = hpass
}

// Login validates and returns a user object if they exist in the database.
func Login(ctx *Context, email, password string) ([]byte, error) {
	var u User
	err := ctx.C("users").Find(bson.M{"email": email}).One(&u)
	if err != nil {
		return []byte(""), err
	}

	if err = bcrypt.CompareHashAndPassword(u.Password, []byte(password)); err == nil {
		token, err := ctx.Auth.GenerateToken(u.ID)
		if err != nil {
			return []byte(""), err
		} else {
			response, _ := json.Marshal(TokenAuthentication{token})
			return response, nil
		}
	}

	return []byte(""), err
}

func LoginFB(ctx *Context, fbid string) (u *User, err error) {
	err = ctx.C("users").Find(bson.M{"facebook": fbid}).One(&u)
	if err != nil {
		return
	}

	return
}

func LoginTwitter(ctx *Context, twtid string) (u *User, err error) {
	err = ctx.C("users").Find(bson.M{"twitter": twtid}).One(&u)
	if err != nil {
		return
	}

	return
}