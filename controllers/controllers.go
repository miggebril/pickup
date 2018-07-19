package controllers

import (
	"encoding/json"
	"pickup/models"
	"pickup/helpers"
  	"net/http"
  	"github.com/gorilla/pat"
  	"html/template"
  	"path/filepath"
  	"gopkg.in/mgo.v2/bson"
  	"sync"
  	"fmt"
  	"errors"
)

var cachedTemplates = map[string]*template.Template{}
var cachedMutex sync.Mutex

var router *pat.Router

func Init(r *pat.Router) {
	router = r
}

func reverse(name string, things ...interface{}) string {
	//convert the things to strings
	strs := make([]string, len(things))
	for i, th := range things {
		strs[i] = fmt.Sprint(th)
	}
	//grab the route
	u, err := router.GetRoute(name).URL(strs...)
	if err != nil {
		panic(err)
	}
	return u.Path
}

var funcs = template.FuncMap{
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
	"reverse": reverse,
	"RenderPercent": helpers.RenderPercent,
	"RenderTimestamp": helpers.RenderTimestamp,
	"GetIDEncoded": helpers.GetIDEncoded,
	"RenderTimeDetails": helpers.RenderTimeDetails,
}

func T(name string, pjax string) *template.Template {
	cachedMutex.Lock()
	defer cachedMutex.Unlock()

	t := template.Must(template.New("_base.html").ParseFiles(
		"templates/_base.html",
		filepath.Join("templates", name),
	))

	return t
}

func TestName(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	u := make(map[string]string)

	u["name"] = ctx.User.Username
	js, err := json.Marshal(u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return nil
}

func LoginForm(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	return T("login.html", r.Header.Get("X-PJAX")).Execute(w, map[string]interface{}{
		"ctx": ctx,
	})
}

func Login(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	var form bson.M

    if r.Body == nil {
        fmt.Println(err.Error())
        http.Error(w, "Request required.", 400)
        return
    }

    err = json.NewDecoder(r.Body).Decode(&form)
    if err != nil {
        fmt.Println(err.Error())
        http.Error(w, err.Error(), 400)
        return
    }

    email := form["Email"].(string)
    password := form["Password"].(string)

	token, err := models.Login(ctx, email, password)
	if err != nil {
		http.Error(w, "Invalid password.", http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(token)
	return nil
}

func RegisterForm(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	return T("register.html", r.Header.Get("X-PJAX")).Execute(w, map[string]interface{}{
		"ctx": ctx,
		"email": r.FormValue("email"),
		"username": r.FormValue("username"),
	})
}

func Register(w http.ResponseWriter, r *http.Request, ctx *models.Context) error {
	username, email, password, password_confirm := r.FormValue("username"), r.FormValue("email"), r.FormValue("password"), r.FormValue("password_confirm")
	

	if len(password) < 8 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}

	if password != password_confirm {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	u := &models.User{
		Email: 	  email,
		Username: username,
		ID:       bson.NewObjectId(),
	}
	u.SetPassword(password)

	if err := ctx.C("users").Insert(u); err != nil {
		r.Form.Del("email")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	token, err := models.Login(ctx, email, password)
	if err != nil {
		http.Error(w, "Invalid password.", http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(token)
	return nil
}

func Index(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	return T("index.html", "na").Execute(w, map[string]interface{}{
		"ctx":     ctx,
	})
}
