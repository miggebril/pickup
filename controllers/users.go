package controllers

import (
    "pickup/models"
    "pickup/helpers"
    "net/http"
    "gopkg.in/mgo.v2/bson"
    "log"
    "encoding/json"
)

func UsersNew(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
    log.Println("UsersNew")
    var form bson.M

    if r.Body == nil {
        log.Println(err.Error())
        http.Error(w, "Request required.", 400)
        return
    }

    err = json.NewDecoder(r.Body).Decode(&form)
    if err != nil {
        log.Println(err.Error())
        http.Error(w, err.Error(), 400)
        return
    }

    u := &models.User{
        Email:    form["Email"].(string),
        Username: form["Username"].(string),
        Verified: true,
        ID:       bson.NewObjectId(),
    }

    password := form["Password"].(string)

    u.SetPassword(password)

    if err != nil {
        log.Println("Error parsing new user form: ", err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return err
    }

    if err := ctx.C("users").Insert(u); err != nil {
        log.Println("Error inserting user to DB: ", err.Error())
        http.Error(w, err.Error(), 400)
        return err
    }

    log.Println("New user inserted to DB")

    token, err := models.Login(ctx, u.Email, password)
    if err != nil {
        log.Println("Failed to login", err)
        http.Error(w, err.Error(), 400)
        return err
    }

    log.Println("success")

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(token)
    return nil
}

func UserInfo(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
    log.Println("UserInfo")
    id, err := helpers.ObjectIdFromString(r.URL.Query().Get(":id"))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return nil
    }

    if r.Body == nil {
        log.Println(err.Error())
        http.Error(w, "Request required.", 400)
        return
    }

    var u models.User
    log.Println("Querying for user id: ", id)
    err = ctx.C("users").Find(bson.M{"_id":id}).One(&u)
    if err != nil {
        log.Println("Failed to query vendor index.", err)
        return nil
    }

    js, err := json.Marshal(&u)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return err
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(js)
    return nil
}

func UserAccount(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
    log.Println("UserAccount")
    js, err := json.Marshal(ctx.User)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return err
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(js)
    return nil
}

func UsersIndex(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
    log.Println("UsersIndex")
    queryParam := r.URL.Query().Get("q")

    var u []models.User
    query := ctx.C("users").Find(bson.M{"email": bson.RegEx{Pattern: queryParam, Options: "i"}}).Limit(30)
    if err = query.All(&u); err != nil {
        log.Println("Failed to query users index.", err)
        return nil
    }

    js, err := json.Marshal(u)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return err
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
    return nil
}
