package main

import (
	"pickup/models"
	"github.com/manveru/faker"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"os"
	"log"
    "github.com/mitchellh/mapstructure"
)

const RADIUS = 10 // In miles

var session *mgo.Session
var database string

var emails []string
var courts []models.Court

var NEW_YORK = models.Location{Latitude: 40.761842, Longitude: -73.981626}

func addGames(count int) {

    fake, err := faker.New("en")
    if err != nil {
        log.Println("Failed to create faker.", err)
        return
    }

    games := make([]models.Game, count)
    for i := 0; i < count; i++ {
        teamPlayers := rand.Intn(10) + 1
        if (teamPlayers % 2 == 1) {
            teamPlayers += 1
        }

        users := make([]models.User, teamPlayers)
        randomPlayerSelect := session.DB("pickup").C("users").Pipe([]bson.M{{"$match": bson.M{}}})

        pipeIter := randomPlayerSelect.Iter()

        var playerCount int = 0
        var temp map[string]interface{}
        
        for pipeIter.Next(&temp) {
            if (playerCount >= teamPlayers) {
                break;
            }

            users[playerCount] = models.User{
                Email : temp["email"].(string),
                ID : temp["_id"].(bson.ObjectId),
                Username : temp["username"].(string),
            }

            playerCount += 1
        }

        var host models.User = users[0]
        
        var home []models.User = users[0:len(users)/2]
        var away []models.User = users[len(users)/2:len(users)]

        games[i] = models.Game{ID: bson.NewObjectId()}
        games[i].Name = fake.UserName()
        games[i].Owner = host

        var court models.Court
        randomCourtSelect := session.DB("pickup").C("courts").Pipe([]bson.M{{"$match": bson.M{}}})
        if err = randomCourtSelect.One(&court); err != nil {
            log.Println("Failed to query random court", err)
        }

        var tempCourt map[string]interface{}
        courtIter := randomCourtSelect.Iter()

        for courtIter.Next(&tempCourt) {
            if (tempCourt["name"].(string) == "") {
                continue
            } else {
                mapstructure.Decode(tempCourt, &court)
                break
            }
        }

        games[i].Court = court.ID
        games[i].HomeTeam = home
        games[i].AwayTeam = away

        games[i].HomeTeam[0] = games[i].Owner

        games[i].Result = models.Result {
            HomeScore: int32(rand.Intn(22)),
            AwayScore: int32(rand.Intn(22)),
        }

        games[i].BoxScore = make(map[string]interface{}, 0)
    }

    log.Println(games)

    for i := 0; i < count; i++ {
        if err := session.DB("pickup").C("games").Insert(&games[i]); err != nil {
            log.Println("Failed to create game:", err)
        }
    }
}


func addCourts(n int) {
    fake, err := faker.New("en")
    if err != nil {
        log.Println("Failed to create faker.", err)
        return
    }

    courts = make([]models.Court, n)
    for i := 0; i < n; i++ {
        courts[i] = models.Court{ID: bson.NewObjectId()}
        courts[i].Name = fake.CompanyName() + " Park"
        courts[i].Rating = rand.Float64() * 6
        courts[i].City = "New York"
        courts[i].State = "New York"
        courts[i].Location = scramble(NEW_YORK)
    }

    log.Println(courts)

    for i := 0; i < n; i++ {
        if err := session.DB("pickup").C("courts").Insert(&courts[i]); err != nil {
            log.Println("Failed to create court:", err)
        }
    }
}

func addUsers(n int) {
    fake, err := faker.New("en")
    if err != nil {
        log.Println("Failed to create faker.", err)
        return
    }

    emails = make([]string, n)
    users := make([]models.User, n)
    for i := 0; i < n; i++ {
        users[i] = models.User{ID: bson.NewObjectId()}
        first := fake.FirstName()
        last := fake.LastName()
        users[i].Email = first+last+"@fakemail.org"
        emails[i] = users[i].Email
        users[i].Username = fake.UserName()
        users[i].SetPassword("password")
        users[i].Verified = true
    }

    log.Println(users)

    for i := 0; i < n; i++ {
        if err := session.DB("pickup").C("users").Insert(&users[i]); err != nil {
            log.Println("Failed to create user:", err)
        }
    }
}

func main() {
	var err error
	session, err = mgo.Dial(os.Getenv("127.0.0.1"))
	if err != nil {
		panic(err)
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

    //addUsers(100)
    addCourts(100)
    addGames(200)
}

func scramble(l models.Location) (r models.Location) {
    r.Latitude = l.Latitude+float64(rand.Intn(2*RADIUS)-RADIUS)/60
    r.Longitude = l.Longitude+float64(rand.Intn(2*RADIUS)-RADIUS)/60
    return
}