package models

import (
  "net/http"
  "crypto/rsa"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "log"
)

type Context struct {
  Database *mgo.Database
  User     *User
  Auth     *JWTAuthenticationBackend
}

func (c *Context) Close() {
  c.Database.Session.Close()
}

//C is a convenience function to return a collection from the context database.
func (c *Context) C(name string) *mgo.Collection {
  return c.Database.C(name)
}

func NewContext(req *http.Request, session *mgo.Session, database string, auth *JWTAuthenticationBackend) (*Context, error) {
  ctx := &Context{
      Database: session.Clone().DB(database),
      Auth: auth,
  }
  
  return ctx, nil
}

func (c *Context) GetUser(id bson.ObjectId) (User, error) {
  coll := c.C("users")
  query := coll.Find(bson.M{"_id":id}).Sort("-timestamp")
  log.Println("GetUser:", id)
  var u User
  err := query.One(&u)
  return u, err
}

func (c *Context) SetUser(id bson.ObjectId) (error) {
  coll := c.C("users")
  query := coll.Find(bson.M{"_id":id})
  log.Println("GetUser:", id)
  return query.One(&c.User)
}

func (c *Context) GetOneUser() (User, error) {
  coll := c.C("users")
  query := coll.Find(bson.M{})
  var u User
  err := query.One(&u)
  return u, err
}

type JWTAuthenticationBackend struct {
  PrivateKey *rsa.PrivateKey
  PublicKey  *rsa.PublicKey
}

const (
  tokenDuration = 72
  expireOffset  = 3600
)