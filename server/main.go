package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

type Todo struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Completed bool   `json:"completed"`
}

func mainHandler(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{
		"message": "Hello World",
	})
}

var todos = []Todo{}

func createTodo(c *fiber.Ctx) error {
	todo := Todo{}
	if err := c.BodyParser(&todo); err != nil {
		fmt.Println(string("\033[31m"), err, string("\033[0m"))
		return err
	}

	if todo.Body == "" || todo.Title == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Body add title are both required"})
	}

	todo.Id = len(todos) + 1

	todos = append(todos, todo)
	return c.Status(200).JSON(fiber.Map{"message": "Todo created", "todo": todo})
}

func getTodos(c *fiber.Ctx) error {
	return c.Status(200).JSON(todos)
}

func getTodo(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	for _, todo := range todos {
		if todo.Id == id {
			return c.Status(200).JSON(todo)
		}
	}
	return c.Status(404).JSON(fiber.Map{"message": "Todo not found"})
}

func updateTodo(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	todo := Todo{}
	if err := c.BodyParser(&todo); err != nil {
		return err
	}
	for i, t := range todos {
		if t.Id == id {
			todos[i].Title = todo.Title
			todos[i].Body = todo.Body
			todos[i].Completed = todo.Completed
			return c.Status(200).JSON(fiber.Map{"message": "Todo updated"})
		}
	}
	return c.Status(404).JSON(fiber.Map{"message": "Todo not found"})
}

func deleteTodo(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	for i, t := range todos {
		if t.Id == id {
			todos = append(todos[:i], todos[i+1:]...)
			return c.Status(200).JSON(fiber.Map{"message": "Todo deleted"})
		}
	}
	return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
}

func main() {

	fmt.Println("Hello World")
	app := fiber.New()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	PORT := os.Getenv("PORT")

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} | ${latency} | ${ip} | ${method} ${path}\n",
	}))

	app.Get("/", mainHandler)

	todoApi := app.Group("/api/")

	todoApi.Get("/todos", getTodos)
	todoApi.Get("/todos/:id", getTodo)

	todoApi.Post("/todos", createTodo)

	todoApi.Put("/todos/:id", updateTodo)

	todoApi.Delete("/todos/:id", deleteTodo)

	log.Fatal(app.Listen(":" + PORT))
}
