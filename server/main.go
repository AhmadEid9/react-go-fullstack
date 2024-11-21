package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

type Todo struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title     string             `json:"title"`
	Body      string             `json:"body"`
	Completed bool               `json:"completed"`
}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		fmt.Println("Error while fetching todos", err)
		return err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo

		if err := cursor.Decode(&todo); err != nil {
			fmt.Println("Error while Decoding Todo", err)
			return err
		}

		todos = append(todos, todo)
	}

	return c.Status(200).JSON(todos)
}

func getTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		fmt.Println("Invalid Object Id")
		return c.Status(400).JSON(fiber.Map{"error": err})
	}

	filter := bson.M{"_id": objectId}

	var todo Todo
	err = collection.FindOne(context.Background(), filter).Decode(&todo)

	if err != nil {
		fmt.Println("Error while finding todo", err)
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	}

	return c.Status(200).JSON(todo)
}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)

	if err := c.BodyParser(&todo); err != nil {
		fmt.Println("Error while creating todo", err)
		return c.Status(400).JSON(fiber.Map{"error": "Error Parsing Request Body"})
	}

	if strings.TrimSpace(todo.Body) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Body cannot be empty"})
	}

	if strings.TrimSpace(todo.Title) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Body cannot be empty"})
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)

	if err != nil {
		fmt.Println("Error While Inserting Todo into the database", err)
		return c.Status(400).JSON(fiber.Map{"error": err})
	}

	todo.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(fiber.Map{"todo": todo})
}

func updateTodo(c *fiber.Ctx) error {
	todo := new(Todo)

	if err := c.BodyParser(&todo); err != nil {
		fmt.Println("Error while creating todo", err)
		return c.Status(400).JSON(fiber.Map{"error": "Error Parsing Request Body"})
	}

	if strings.TrimSpace(todo.Body) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Body cannot be empty"})
	}

	if strings.TrimSpace(todo.Title) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Body cannot be empty"})
	}

	id := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		fmt.Println("Error: ", err)
		return c.Status(400).JSON(fiber.Map{"error": err})
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"title": todo.Title, "body": todo.Body, "completed": todo.Completed}}
	_, err = collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		fmt.Println("Error While updating:", err)
		return c.Status(400).JSON(fiber.Map{"err": err})
	}

	return c.Status(201).JSON(fiber.Map{"message": "todo updated"})
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")

	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		fmt.Println("Invalid Object Id", err)
		return c.Status(400).JSON(fiber.Map{"error": err})
	}

	filter := bson.M{"_id": objectId}

	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": err})
	}

	return c.Status(200).JSON(fiber.Map{"success": true})
}

func main() {
	app := fiber.New()

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} | ${latency} | ${ip} | ${method} ${path}\n",
	}))

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	PORT := os.Getenv("PORT")
	MongoDB_URI := os.Getenv("MONGODB_URI")

	clientOptions := options.Client().ApplyURI(MongoDB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("Error creating mongo client", err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Error connecting to mongo client", err)
	}

	collection = client.Database("golang_db").Collection("todos")

	apiRoutes := app.Group("/api/todos")
	apiRoutes.Post("/", createTodo)
	apiRoutes.Get("/", getTodos)
	apiRoutes.Get("/:id", getTodo)
	apiRoutes.Put("/:id", updateTodo)
	apiRoutes.Delete("/:id", deleteTodo)

	fmt.Println("Connected to MongoDB")
	app.Listen(":" + PORT)
}
