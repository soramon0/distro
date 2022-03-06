package models

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// swagger:parameters recipes newRecipe
type Recipe struct {
	//swagger:ignore
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

func SeedRecipes(collection *mongo.Collection, cache *redis.Client) {
	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		log.Println(err)
		return
	}

	if count > 0 {
		log.Println("Recipes already seeded")
		return
	}

	var recipes []interface{}
	file, err  := ioutil.ReadFile("recipes.json")
	if err != nil {
		log.Println(err)
		return
	}

	err = json.Unmarshal(file, &recipes)
	if err != nil {
		log.Println(err)
		return
	}

	result, err := collection.InsertMany(context.Background(), recipes)
	if err != nil {
		log.Println(err)
		return
	}

	cache.Del(context.Background(), "recipes")

	log.Printf("Seeded %d recipes\n", len(result.InsertedIDs))
}
