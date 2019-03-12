package main

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"labix.org/v2/mgo/txn"
)

// waitBeforeRetry is the time hub waits before trying to reconnect
// to mongo in case of failure
const (
	waitBeforeRetry = 5 * time.Second

	userCollectionName = "users"

	defaultMongoAddress = "localhost"
	defaultMongoDBName  = "test"
)

type User struct {
	Email string `json:"email"`
}

type mongoStorage struct {
	session *mgo.Session
	dbName  string
}

func newMongoStorage(address string, dbName string) *mongoStorage {
	var (
		session *mgo.Session
		err     error
	)

	for {
		session, err = mgo.Dial(address)
		if err != nil {
			log.Infof("cannot connect to mongo: %v, retry in %+v\n", err, waitBeforeRetry)
			time.Sleep(waitBeforeRetry)
		} else {
			break
		}
	}

	return &mongoStorage{
		session: session,
		dbName:  dbName,
	}
}

func main() {
	ms := newMongoStorage(defaultMongoAddress, defaultMongoDBName)

	session := ms.session.Copy()
	defer session.Close()

	users := session.DB(ms.dbName).C(userCollectionName)

	alice := &User{
		Email: "alice@gmail.com",
	}
	aliceId := bson.NewObjectId()

	bob := &User{
		Email: "bob@gmail.com",
	}
	bobId := bson.NewObjectId()

	runner := txn.NewRunner(users)
	ops := []txn.Op{
		{
			C:      userCollectionName,
			Id:     aliceId,
			Insert: alice,
		},
		{
			C:      userCollectionName,
			Id:     bobId,
			Insert: bob,
		},
	}
	mongoTxId := bson.NewObjectId() // Optional
	if err := runner.Run(ops, mongoTxId, nil); err != nil {
		log.Fatal(err)
	}

	var allUsers []*User
	if err := users.Find(bson.M{}).All(&allUsers); err != nil {
		log.Fatal(err)
	}
	for _, user := range allUsers {
		fmt.Println(user)
	}

	//var allUsersHelper []map[string]*User
	//if err := users.Find(bson.M{}).All(&allUsersHelper); err != nil {
	//	log.Fatal(err)
	//}
	//
	//allUsers := make([]*User, 0)
	//for _, user := range allUsersHelper {
	//	if value, ok := user[defaultKey]; ok {
	//		allUsers = append(allUsers, value)
	//	}
	//}
	//
	//fmt.Println(len(allUsers))
	//for _, user := range allUsers {
	//	fmt.Println(user)
	//}
}
