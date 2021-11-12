package models

import (
	"auth-api/db"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/matthewhartstonge/argon2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username,omitempty"`
	Password  []byte             `bson:"password,omitempty"`
	Image     string             `bson:"img,omitempty"`
	CreatedAt time.Time          `bson:"createdAt,ompitempty"`
}

func getCollection() *mongo.Collection {
	return db.GetDb().Database("apiAuth").Collection("users")
}

func GetAllUser(page int) ([]User, error) {
	collection := getCollection()

	var users []User

	findOptions := options.Find()
	findOptions.SetLimit(20)
	findOptions.SetSkip(int64(page * 20))
	findOptions.SetProjection(bson.M{"password": 0})

	f, err := collection.Find(context.TODO(), bson.D{}, findOptions)

	if err != nil {
		return nil, err
	}

	err = f.All(context.TODO(), &users)

	if err != nil {
		return nil, err
	}

	return users, nil
}

func CreateUser(user User) (*mongo.InsertOneResult, error) {
	collection := getCollection()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	insertResult, err := collection.InsertOne(ctx, user)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return insertResult, nil
}

func FindByUsername(username string) (*User, error) {
	collection := getCollection()
	u := new(User)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{"username": username})

	err := res.Decode(&u)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func FindUserByIdString(id string) *User {
	collection := getCollection()
	u := new(User)

	objId, _ := primitive.ObjectIDFromHex(id)
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	res := collection.FindOne(ctx, bson.M{"_id": objId}, options.FindOne().SetProjection(bson.M{"password": 0}))
	res.Decode(&u)

	return u
}

func FindUserById(id string) *User {
	collection := getCollection()
	u := new(User)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{"_id": id}, options.FindOne().SetProjection(bson.M{"password": 0}))
	res.Decode(&u)

	return u
}

func UpdateUser(user *User, body map[string]string) (*User, error) {
	collection := getCollection()

	if body["username"] != "" {
		u, _ := FindByUsername(body["username"])
		if u != nil {
			return nil, errors.New("username already in use")
		}
		user.Username = body["username"]
	}
	if body["password"] != "" {
		argon := argon2.DefaultConfig()
		hash, err := argon.HashEncoded([]byte(body["password"]))
		if err != nil {
			return nil, errors.New("internal server error")
		}
		user.Password = hash
	}
	if body["img"] != "" {
		user.Image = body["img"]
	}

	filter := bson.M{"_id": user.ID}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	_, err := collection.UpdateOne(ctx, filter, bson.M{"$set": user})

	if err != nil {
		return nil, errors.New("internal server error")
	}

	return user, nil
}

func DeleteUser(user *User) error {
	collection := getCollection()

	filter := bson.M{"_id": user.ID}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	_, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	return nil
}
