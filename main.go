package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var db *sqlx.DB

const jwtSecret = "infinitas"

func main() {

	var err error
	db, err = sqlx.Open("mysql", "root:@tcp(127.0.0.1:3306)/users")

	if err != nil {
		panic(err)
	}

	app := fiber.New()

	app.Use("/hello", jwtware.New(jwtware.Config{
		SigningMethod: "HS256",
		SigningKey:    []byte(jwtSecret),
		SuccessHandler: func(c *fiber.Ctx) error {
			return c.Next()
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return fiber.ErrUnauthorized
		},
	}))

	app.Post("/signup", Signup)
	app.Post("/login", Login)
	app.Get("/hello", hello)

	app.Listen(":8000")
}

func Signup(c *fiber.Ctx) error {
	request := SignupRequest{}
	err := c.BodyParser(&request)
	if err != nil {
		return err
	}

	if request.Username == "" || request.Password == "" {
		return fiber.ErrUnprocessableEntity
	}

	password, err := bcrypt.GenerateFromPassword([]byte(request.Password), 10)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	query := "insert into user (username, password) values (?, ?)"
	result, err := db.Exec(query, request.Username, string(password))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	user := User{
		Id:       int(id),
		Username: request.Username,
		Password: string(password),
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func Login(c *fiber.Ctx) error {
	request := LoginRequest{}
	err := c.BodyParser(&request)
	if err != nil {
		return err
	}

	if request.Username == "" || request.Password == "" {
		return fiber.ErrUnprocessableEntity
	}

	user := User{}
	query := "select id, username, password from user where username=?"
	err = db.Get(&user, query, request.Username)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, " incorect username")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, " incorect password")
	}

	cliams := jwt.StandardClaims{
		Issuer:    strconv.Itoa(user.Id),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, cliams)
	token, err := jwtToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"jwtToken": token,
	})
}

func hello(c *fiber.Ctx) error {
	return c.SendString("Hello World")
}

type User struct {
	Id       int    `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
}

type SignupRequest struct {
	Username string `json:"username"`
	Password string `json:"passowrd"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"passowrd"`
}

func Fiber() {
	app := fiber.New()

	//MiddleWare
	// app.Use(func(ctx fiber.Ctx) error {
	// 	ctx.Locals("name", "korn") //ส่งค่าผ่าตัว local ไปให้ที่อื่น
	// 	fmt.Println("Before")
	// 	ctx.Next()
	// 	fmt.Println("After")
	// 	return nil
	// })

	// app.Use(requestid.New())

	// //MiddleWare เฉพาะ path
	// app.Use("/hello", func(ctx fiber.Ctx) error {
	// 	fmt.Println("Before Hello")
	// 	ctx.Next()
	// 	fmt.Println("After Hello")
	// 	return nil
	// })

	//Group
	v1 := app.Group("/v1")
	v1.Get("/hello", func(c *fiber.Ctx) error {
		return c.SendString("v1")
	})

	//GET
	app.Get("/hello", func(c *fiber.Ctx) error {
		name := c.Locals("name")
		return c.SendString(fmt.Sprintf("Hello %v", name))
	})

	//POST
	app.Post("/hello", func(c *fiber.Ctx) error {
		return c.SendString("POST : Hello")
	})

	//Parameters
	app.Get("/hello/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		return c.SendString(name)
	})

	//Parameters optional
	app.Get("/hello/:name/:phone?", func(c *fiber.Ctx) error {
		name := c.Params("name")
		phone := c.Params("phone")
		return c.SendString(name + "phone" + phone)
	})

	//Query
	app.Get("/query", func(c *fiber.Ctx) error {
		name := c.Query("name")
		return c.SendString("name :" + name)
	})
	type Person struct {
		Name  string `json:"id"`
		Phone string `json:"phone"`
	}

	//Query
	app.Get("/query2", func(c *fiber.Ctx) error {
		person := Person{}
		c.QueryParser(&person)
		return c.JSON(person)
	})

	//Wildcards
	app.Get("/wildcards/*", func(c *fiber.Ctx) error {
		wildcard := c.Params("*")
		return c.SendString(wildcard)
	})

	app.Static("/", "./root", fiber.Static{
		Index:         "index.html",
		CacheDuration: time.Second * 10,
	})

	//NewError
	app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.ErrBadRequest.Code, "not found")
	})

	//Mount
	// userApp := fiber.New()
	// userApp.Get("/login", func(c *fiber.Ctx) error {
	// 	return c.SendString("mount")
	// })

	// app.Mount("/user", userApp)

	app.Listen(":8000")
}
