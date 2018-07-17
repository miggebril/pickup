package main

import (
	"pickup/controllers"
	"github.com/goods/httpbuf"
	"github.com/gorilla/pat"
	"encoding/gob"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"os"
	"io"
	"log"
	"pickup/models"
	"pickup/helpers"
	"fmt"
	"encoding/xml"
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"pickup/settings"
	jwt "github.com/dgrijalva/jwt-go"
	jwt_request "github.com/dgrijalva/jwt-go/request"
	"gopkg.in/Graylog2/go-gelf.v1/gelf"
)

var session *mgo.Session
var database string
var router *pat.Router
var auth *models.JWTAuthenticationBackend

type Settings struct {
	XMLName xml.Name       `xml:"settings"`
	AppId string         `xml:"appid"`
	AppSecret string     `xml:"secretkey"`
}

type handler func(http.ResponseWriter, *http.Request, *models.Context) error

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
	log.Println("Auth:", r.Header.Get("Authorization"))

	buf := new(httpbuf.Buffer)
	log.Println(buf)

	token, err := jwt_request.ParseFromRequest(r, jwt_request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		} else {
			return auth.PublicKey, nil
		}
	})

	  //create the context
	  ctx, _ := models.NewContext(r, session, database, auth)
	  //if err != nil {
	  	  //http.Error(w, err.Error(), http.StatusInternalServerError)
	  //}
	  defer ctx.Close()

	  if err == nil && token.Valid {//&& !auth.IsInBlacklist(r.Header.Get("Authorization")) {
		//ctx, _ = models.NewContext(r, dbmap, database, auth)
		//defer ctx.Close()
		claims := token.Claims.(jwt.MapClaims)
		uid, _ := helpers.ObjectIdFromString(claims["sub"].(string))

		err = ctx.C("users").Find(bson.M{"_id":uid}).Sort("-timestamp").One(&ctx.User)
		log.Println(ctx.User)
		if err != nil {
			log.Println("Fetch error")
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else {
			log.Println("Auth:", claims["sub"])
			uid, _ := helpers.ObjectIdFromString(claims["sub"].(string))

			err = ctx.C("users").Find(bson.M{"_id":uid}).Sort("-timestamp").One(&ctx.User)
			log.Println(ctx.User)
			if err != nil {
				log.Println("Fetch error")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
	} else if r.URL.Path != "/login" && !((r.URL.Path == "/users" || r.URL.Path == "/verify") && r.Method == "POST") {
		log.Println("Unauth")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}


	err = h(buf, r, ctx)
	buf.Apply(w)
}

func init() {
  gob.Register(bson.ObjectId(""))
}

func main() {
	gelfWriter, err := gelf.NewWriter("127.0.0.1:12202")
	if err != nil {
		log.Fatalf("gelf.NewWriter: %s", err)
	}
	// log to both stderr and graylog2
	log.SetOutput(io.MultiWriter(os.Stderr, gelfWriter))
	log.Printf("logging to stderr & graylog2@")

	session, err = mgo.Dial(os.Getenv("127.0.0.1"))
	if err != nil {
		panic(err)
	}

	settings.Init()

	auth = &models.JWTAuthenticationBackend{
		PrivateKey: getPrivateKey(),
		PublicKey:  getPublicKey(),
	}
	
	database = session.DB("pickup").Name
	if err := session.DB("pickup").C("users").EnsureIndex(mgo.Index{
	        Key:    []string{"email"},
	        Unique: true,
	    }); err != nil {
	        log.Println("Ensuring unqiue index on users:", err)
    }

    geoIndex := mgo.Index{
	    Key: []string{"$2d:location"},
	    Bits: 26,
	}

	if err := session.DB("pickup").C("courts").EnsureIndex(geoIndex); err != nil {
		log.Println("Ensuring unqiue index on court coordinates:", err)
		return
	}

	var u []models.User
	query := session.DB("pickup").C("users").Find(bson.M{"verified":true})
	if err = query.All(&u); err != nil {
		log.Println("Failed to query users for name/ID listings.", err)
	} else {
		for _, user := range u {
			log.Println(user.GetIDEncoded(), user.Username)
		}
	}

	router = pat.New()
	controllers.Init(router)

	router.Add("POST", "/login", handler(controllers.Login)).Name("Login")

	router.Add("GET", "/users/me", handler(controllers.UserAccount))
	// router.Add("POST", "/users/me", handler(controllers.UserUpdate))


	router.Add("GET", "/users/{id}", handler(controllers.UserInfo))
	router.Add("GET", "/users", handler(controllers.UsersIndex))

	router.Add("POST", "/users", handler(controllers.UsersNew))
	// router.Add("POST", "/verify", handler(controllers.UsersVerify))

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	router.Add("GET", "/", handler(controllers.Index)).Name("index")

	if err := http.ListenAndServe(":8077", router); err != nil {
		panic(err)
	}
}

func getPrivateKey() *rsa.PrivateKey {
	privateKeyFile, err := os.Open(settings.Get().PrivateKeyPath)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := privateKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	privateKeyFile.Close()

	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	return privateKeyImported
}

func getPublicKey() *rsa.PublicKey {
	publicKeyFile, err := os.Open(settings.Get().PublicKeyPath)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := publicKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(publicKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	publicKeyFile.Close()

	publicKeyImported, err := x509.ParsePKIXPublicKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	rsaPub, ok := publicKeyImported.(*rsa.PublicKey)

	if !ok {
		panic(err)
	}

	return rsaPub
}