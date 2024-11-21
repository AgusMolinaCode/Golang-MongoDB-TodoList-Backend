package handlers

import (
	"agustin/internal/config"
	"agustin/internal/models"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var user models.User
    if err := c.BodyParser(&user); err != nil {
        return c.Status(400).SendString(err.Error())
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        return c.Status(500).SendString(err.Error())
    }
    user.Password = string(hashedPassword)

    _, err = config.UserCollection.InsertOne(ctx, user)
    if err != nil {
        return c.Status(500).SendString(err.Error())
    }

    return c.SendStatus(201)
}

func Login(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var input models.User
    if err := c.BodyParser(&input); err != nil {
        return c.Status(400).SendString(err.Error())
    }

    var user models.User
    err := config.UserCollection.FindOne(ctx, bson.M{"username": input.Username}).Decode(&user)
    if err != nil {
        return c.Status(401).SendString("Invalid username or password")
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
        return c.Status(401).SendString("Invalid username or password")
    }

    token := jwt.New(jwt.SigningMethodHS256)
    claims := token.Claims.(jwt.MapClaims)
    claims["id"] = user.ID.Hex()
    claims["username"] = user.Username
    claims["exp"] = time.Now().Add(time.Hour * 5).Unix()

    t, err := token.SignedString([]byte("secret"))
    if err != nil {
        return c.Status(500).SendString(err.Error())
    }

    return c.JSON(fiber.Map{"token": t})
}