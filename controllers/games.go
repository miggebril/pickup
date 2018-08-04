package controllers

import (
    "pickup/models"
    "pickup/helpers"
    "net/http"
    "gopkg.in/mgo.v2/bson"
    "log"
    "encoding/json"
    //"googlemaps.github.io/maps"
)

func GameNew(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	log.Println("GameNew")
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

    courtId, err := helpers.ObjectIdFromString(form["Court"].(string))
    if err != nil {
    	log.Println(err.Error())
        http.Error(w, err.Error(), 400)
        return
    }

    var homeCourt models.Court

    log.Println("Querying for game id: ", courtId)
    if err = ctx.C("courts").Find(bson.M{"_id":courtId}).One(&homeCourt); err != nil {
    	log.Println("Failed to query courts index.", err)
        return nil
    }

    game := &models.Game{
    	HomeCourt: homeCourt,
    	Owner: 	  *ctx.User,
        Court:    courtId,
        Name: 	  form["Name"].(string),
        ID:       bson.NewObjectId(),
    }

    if err := ctx.C("games").Insert(game); err != nil {
        log.Println("Error inserting new game to DB: ", err.Error())
        http.Error(w, err.Error(), 400)
        return err
    }

    response, err := json.Marshal(game)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return err
    }

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	return nil
}

func GameInfo(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	log.Println("GameInfo")
    id, err := helpers.ObjectIdFromString(r.URL.Query().Get(":gameId"))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return nil
    }

    if r.Body == nil {
        log.Println(err.Error())
        http.Error(w, "Request required.", 400)
        return
    }

    var game models.Game
    log.Println("Querying for game id: ", id)
    if err = ctx.C("games").Find(bson.M{"_id":id}).One(&game); err != nil {
        log.Println("Failed to query games index.", err)
        return nil
    }

    response, err := json.Marshal(&game)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return err
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(response)
    return nil
}

func GameIndex(w http.ResponseWriter, r *http.Request, ctx *models.Context) (err error) {
	log.Println("GameIndex")
	var games []models.Game

	query := ctx.C("games").Find(bson.M{"$and" : []bson.M{ 
		bson.M{"result.homescore" : bson.M{"$lt" : 16}}, 
		bson.M{"result.awayscore" : bson.M{"$lt" : 16}},
	}}).Limit(30)

    if err = query.All(&games); err != nil {
        log.Println("Failed to query games index.", err)
        return err
    }
    
    response, err := json.Marshal(games)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return err
    }

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	return nil
}
