package main

import (
	"io"
	"pickup/controllers"
	"github.com/rs/cors"
	"github.com/goods/httpbuf"
	"github.com/gorilla/pat"
	"encoding/gob"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"os"
	"log"
	"pickup/models"
	"pickup/helpers"
	"fmt"
	"encoding/xml"
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http/httputil"
	"pickup/settings"
	"gopkg.in/Graylog2/go-gelf.v1/gelf"
	jwt "github.com/dgrijalva/jwt-go"
	jwt_request "github.com/dgrijalva/jwt-go/request"
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

type handler func(http.ResponseWriter, *http.Request, *models.Context) (err error)

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	log.Println(string(dump))

	buf := new(httpbuf.Buffer)
	log.Println(buf)

	token, err := jwt_request.ParseFromRequest(r, jwt_request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		log.Println()
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		} else {
			log.Println("Token authenticated")
			return auth.PublicKey, nil
		}
	})

	 //create the context
	ctx, _ := models.NewContext(r, session, database, auth)
	defer ctx.Close()

    if err == nil && token.Valid {
    	fmt.Printf("%s", r.URL.Path)
		claims := token.Claims.(jwt.MapClaims)
		uid, _ := helpers.ObjectIdFromString(claims["sub"].(string))
		err = ctx.C("users").Find(bson.M{"_id":uid}).One(&ctx.User)
		
		if err != nil {
			log.Println("Fetch error")
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else {
			log.Println(ctx.User)
			log.Println("Auth:", claims["sub"])
		} 
	} else if !isLoginAttempt(r.URL.Path, r.Method) {
			log.Println("Unauth")
			w.WriteHeader(http.StatusUnauthorized)
			return
	}

	err = h(buf, r, ctx)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	buf.Apply(w)
}

func isLoginAttempt(path string, method string) bool {
	return (path == "/login" || isRegistrationAttempt(path, method))
}

func isRegistrationAttempt(path string, method string) bool {
	return ((path == "/users" || path == "/verify") && method == "POST")
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

	session, err = mgo.Dial("127.0.0.1")
	if err != nil {
		log.Println("Fatal error: ", err)
		panic(err)
	}

	settings.Init()

	auth = &models.JWTAuthenticationBackend{
		PrivateKey: getPrivateKey(),
		PublicKey:  getPublicKey(),
	}
	
	database = session.DB("pickup").Name
	err = session.DB("pickup").C("users").EnsureIndex(mgo.Index{
	        Key:    []string{"email"},
	        Unique: true,});

	helpers.CheckErr(err, "Error ensuring unique email index on users.")

    geoIndex := mgo.Index{
	    Key: []string{"$2d:location"},
	    Bits: 26,
	}

	if err := session.DB("pickup").C("courts").EnsureIndex(geoIndex); err != nil {
		log.Println("Ensuring unqiue index on court coordinates:", err.Error())
		return
	}

	var u []models.User
	query := session.DB("pickup").C("users").Find(bson.M{})
	if err = query.All(&u); err != nil {
		log.Println("Failed to query users for name/ID listings.", err.Error())
	} else {
		for _, user := range u {
			log.Println(user.GetIDEncoded(), user.Username)
		}
	}

	router = pat.New()
	controllers.Init(router)

	router.Add("POST", "/login", handler(controllers.Login)).Name("Login")

	router.Add("GET", "/users/me", handler(controllers.UserAccount))

	router.Add("GET", "/users/{id}", handler(controllers.UserInfo))

	router.Add("GET", "/users", handler(controllers.UsersIndex))
	router.Add("POST", "/users", handler(controllers.UsersNew))

	router.Add("GET", "/games/{gameId}", handler(controllers.GameInfo))

	router.Add("GET", "/games", handler(controllers.GameIndex))
	router.Add("POST", "/games", handler(controllers.GameNew))

	router.Add("GET", "/test", handler(controllers.TestName))

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	
	router.Add("GET", "/", handler(controllers.Index)).Name("index")

	handler := cors.Default().Handler(router)

	if err := http.ListenAndServe(":8077", handler); err != nil {
		log.Println("Fatal error -- panic: ", err.Error())
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