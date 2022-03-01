package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/soramon0/distro/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandler struct {
	ctx context.Context
	collection *mongo.Collection
	cache *redis.Client
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, cache *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		ctx: ctx,
		collection: collection,
		cache: cache,
	}
}

// swagger:operation POST /recipes recipes newRecipe
//
// Create new recipe
//
// ---
// produces:
// - application/json
// responses:
//    '200':
//         description: Successful operation
//    '400':
//         description: Invalid input
//		'500':
//				 description: Interanl server error
func (h *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()

	_, err := h.collection.InsertOne(h.ctx, recipe)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new recipe"})
		return
	}

	h.cache.Del(h.ctx, "recipes")

	c.JSON(http.StatusOK, recipe)
}

// swagger:operation GET /recipes recipes listRecipes
//
// Returns list of recipes
//
// ---
// porduces:
// - application/json
// responses:
//    '200':
//         description: Successful operation
//		'500':
//				 description: Interanl server error
func (h *RecipesHandler)  ListRecipesHandler(c *gin.Context) {
	val, err := h.cache.Get(h.ctx, "recipes").Result()
	if err == redis.Nil {
		cur, err := h.collection.Find(h.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cur.Close(h.ctx)

		recipes := make([]models.Recipe, 0)
		for cur.Next(h.ctx) {
			var recipe models.Recipe
			cur.Decode(&recipe)
			recipes = append(recipes, recipe)
		}

		data, _ := json.Marshal(recipes)
		h.cache.Set(h.ctx, "recipes", string(data), 0)

		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}
}

// swagger:operation PUT /recipes/{id} recipes updateRecipe
//
// Update an existing recipe
//
// ---
// parameters:
// - name: id
//   in: path
//   description: ID of the recipe
//   required: true
//   type: string
// produces:
// - application/json
// responses:
//    '200':
//         description: Successful operation
//    '400':
//         description: Invalid input
//		'500':
//				 description: Interanl server error
func (h *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse id"})
		return
	}

	_, err = h.collection.UpdateOne(h.ctx, bson.M{"_id": objectId}, bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "name", Value: recipe.Name},
			primitive.E{Key: "instructions", Value: recipe.Instructions},
			primitive.E{Key: "ingredients", Value: recipe.Ingredients},
			primitive.E{Key: "tags", Value: recipe.Tags},
		}},
	})
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cache.Del(h.ctx, "recipes")

	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been updated"})
}

// swagger:operation DELETE /recipes/{id} recipes deleteRecipe
//
// Delete an existing recipe
//
// ---
// parameters:
// - name: id
//   in: path
//   description: ID of the recipe
//   required: true
//   type: string
// produces:
// - application/json
// responses:
//    '200':
//         description: Successful operation
//    '400':
//         description: Invalid input
//		'500':
//				 description: Interanl server error
func (h *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Could not parse id"})
		return
	}

	_, err = h.collection.DeleteOne(h.ctx, bson.M{"_id": objectId})
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cache.Del(h.ctx, "recipes")

	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been deleted"})
}

// swagger:operation GET /recipes/search recipes searchRecipe
//
// Search recipes by recipe tag
//
// ---
// parameters:
// - name: tag
//   in: query
//   description: recipe tag to search by
//   required: true
//   type: string
// produces:
// - application/json
// responses:
//    '200':
//         description: Successful operation
func (h *RecipesHandler)  SearchRecipesHandler(c *gin.Context) {
	var recipes []models.Recipe

	tag := c.Query("tag")
	listOfRecipes := make([]models.Recipe, 0)

	for i := 0; i < len(recipes); i++ {
		found := false

		for _, t := range recipes[i].Tags {
			if strings.EqualFold(t, tag) {
				found = true
			}
		}

		if found {
			listOfRecipes = append(listOfRecipes, recipes[i])
		}
	}

	c.JSON(http.StatusOK, listOfRecipes)
}