package gosession

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/go-playground/validator.v9"
)

// TODO
// ensure email field is unique through all new user registers

// DatabaseUser is a struct for an existing user that is binded from the database
// TODO max email size
type DatabaseUser struct {
	ID             primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	Email          string              `json:"email,omitempty" bson:"email,omitempty" validate:"required,min=3"`
	HashedPassword string              `json:"hashedPassword" bson:"hashedPassword" validate:"required,min=10,max=128"`
	CreatedAt      time.Time           `json:"createdAt,omitempty" bson:"createdAt,omitempty" validate:"required"`
	UpdatedAt      time.Time           `json:"updatedAt,omitempty" bson:"updatedAt,omitempty" validate:"required"`
	LastSignedIn   time.Time           `json:"lastSignedIn,omitempty" bson:"lastSignedIn,omitempty"`
	FirstName      string              `json:"firstName,omitempty" bson:"firstName,omitempty" validate:"min=1,max=64,alpha"`
	LastName       string              `json:"lastName,omitempty" bson:"lastName,omitempty" validate:"min=1,max=64,alpha"`
	Verified       bool                `json:"verified" bson:"verified" validate:"required"`
	Confirmations  []*UserConfirmation `json:"confirmations,omitempty" bson:"confirmations,omitempty" validate:"dive"`
	ProfileImage   string              `json:"profileImage,omitempty" bson:"profileImage,omitempty"`
	CoverImage     string              `json:"coverImage,omitempty" bson:"coverImage,omitempty"`
	AboutMe        string              `json:"aboutMe,omitempty" bson:"aboutMe,omitempty" validate:"min=1,max=4096"`
	Role           string              `json:"role,omitempty" bson:"role,omitempty"`
}

// ExistingUser is a struct for an sending back the user with password field removed
type ExistingUser struct {
	ID            primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	Email         string              `json:"email,omitempty" bson:"email,omitempty" validate:"required,min=3"`
	CreatedAt     time.Time           `json:"createdAt,omitempty" bson:"createdAt,omitempty" validate:"required"`
	UpdatedAt     time.Time           `json:"updatedAt,omitempty" bson:"updatedAt,omitempty" validate:"required"`
	LastSignedIn  time.Time           `json:"lastSignedIn,omitempty" bson:"lastSignedIn,omitempty"`
	FirstName     string              `json:"firstName,omitempty" bson:"firstName,omitempty" validate:"min=1,max=64,alpha"`
	LastName      string              `json:"lastName,omitempty" bson:"lastName,omitempty" validate:"min=1,max=64,alpha"`
	Verified      bool                `json:"verified" bson:"verified" validate:"required"`
	Confirmations []*UserConfirmation `json:"confirmations,omitempty" bson:"confirmations,omitempty" validate:"dive"`
	ProfileImage  string              `json:"profileImage,omitempty" bson:"profileImage,omitempty"`
	CoverImage    string              `json:"coverImage,omitempty" bson:"coverImage,omitempty"`
	AboutMe       string              `json:"aboutMe,omitempty" bson:"aboutMe,omitempty" validate:"min=1,max=4096"`
	Role          string              `json:"role,omitempty" bson:"role,omitempty"`
}

// Validation Section --------------------------------------------

// UserValidator - This will validate the user using the structs annotations
type UserValidator struct {
	validator *validator.Validate
}

// Validate - This validates the request body
func (u *UserValidator) Validate(i interface{}) error {
	return u.validator.Struct(i)
}

// Configuration Section --------------------------------------------

func configureUserRoutes() {

	e.GET("/user/:email", getUserByEmail, IPRateLimit(1, 2*time.Second))
	// Get a specific user from MongoDB
	// Docs: https://docs.mongodb.com/manual/reference/command/find/
	e.GET("/users/:id", getUser, IPRateLimit(1, 2*time.Second))

	// Get users from MongoDB
	// Docs: https://docs.mongodb.com/manual/reference/command/find/
	e.GET("/users", getUsers, IPRateLimit(1, 2*time.Second))

	// Update an user record in MongoDB
	// Docs: https://docs.mongodb.com/manual/reference/command/findAndModify/
	e.PUT("/users/:id", putUser, middleware.BodyLimit("1M"), IPRateLimit(1, 2*time.Second))

	// Delete a user from MongoDB with IDs
	// Docs: https://docs.mongodb.com/manual/reference/command/delete/
	e.DELETE("/users/:id", deleteUser, IPRateLimit(1, 2*time.Second))
}

func getUserByEmail(c echo.Context) error {
	collection := mg.Db.Collection("users")

	email := c.Param("email")
	if email == "" {
		return c.String(http.StatusNotFound, "This user email does not exist")
	}

	query := bson.M{"email": &email}

	var user ExistingUser

	err := collection.FindOne(c.Request().Context(), &query).Decode(&user)

	if err != nil {
		fmt.Println(err.Error())
		return c.String(http.StatusNotFound, err.Error())
	}

	// return user in JSON format
	return c.JSON(http.StatusOK, user)
}

// getUserByID
func getUserByID(id string) (*ExistingUser, error) {

	collection := mg.Db.Collection("users")

	// Get the id from the paramaters
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("This user id is invalid")
	}
	query := bson.M{"_id": &userID}

	var user ExistingUser

	err = collection.FindOne(context.Background(), &query).Decode(&user)

	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New(err.Error())
	}

	// return user in JSON format
	return &user, nil
}

// getUser
func getUser(c echo.Context) error {

	collection := mg.Db.Collection("users")

	id := c.Param("id")
	if id == "" {
		return c.String(http.StatusNotFound, "This user id does not exist")
	}

	// Get the id from the paramaters
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.String(http.StatusNotFound, "This user id is invalid")
	}
	query := bson.M{"_id": &userID}

	var user ExistingUser

	err = collection.FindOne(c.Request().Context(), &query).Decode(&user)

	if err != nil {
		fmt.Println(err.Error())
		return c.String(http.StatusNotFound, err.Error())
	}

	// return user in JSON format
	return c.JSON(http.StatusOK, user)
}

// getUsers - This will get all users defined by supplied filter params
func getUsers(c echo.Context) error {

	ctx := c.Request().Context()

	collection := mg.Db.Collection("users")

	// create the pipleine for the different filters
	// "admin" = the role
	// "users" = the type of pipeline for the filters

	userRole := getUserRole(c)
	pipeline := addFiltersToPipeline(c, userRole, "users")

	// TODO ensure once they match an email the rest of the filters won't be on everything else
	// ensure that it matches just one via the email

	// ensure passed in params are sanitized

	// the big get users api is only for admin.
	// normal users will call public endpoints that will call this endpoint using pre-defined params (or not- depend on speed)

	// func addFilters(pipeline, context, role middleware)

	var users []*ExistingUser

	cur, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return c.String(http.StatusNotFound, "No users found")
	}

	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var user *ExistingUser
		if err = cur.Decode(&user); err != nil {
			fmt.Println(err)
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		return c.String(http.StatusNotFound, "No users found")
	}

	// return users list in JSON format
	return c.JSON(http.StatusOK, users)
}

// This is the route where a user can change their own settings
func putUser(c echo.Context) error {

	collection := mg.Db.Collection("users")

	id := c.Param("id")
	if id == "" {
		return c.String(http.StatusNotFound, "This user id does not exist")
	}

	// Get the id from the paramaters
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.String(http.StatusNotFound, "This user id is invalid")
	}

	user := ExistingUser{}
	c.Echo().Validator = &UserValidator{validator: v}

	// Parse body into struct
	if err := c.Bind(&user); err != nil {
		return c.String(http.StatusNotAcceptable, "This is not a valid user object")
	}

	if err := c.Validate(user); err != nil {
		log.Printf("Unable to validate the user %+v %v", user, err)
		// TODO implement: https://medium.com/@apzuk3/input-validation-in-golang-bc24cdec1835
		return c.JSON(http.StatusPartialContent, err.Error())
	}

	// Find the user and update its data
	query := bson.D{{Key: "_id", Value: &userID}}
	update := bson.D{
		{Key: "$set",
			Value: bson.D{
				{Key: "aboutMe", Value: &user.AboutMe},
				{Key: "updatedAt", Value: primitive.NewDateTimeFromTime(time.Now().UTC())},
				{Key: "firstName", Value: &user.FirstName},
				{Key: "lastName", Value: &user.LastName},
				{Key: "profileImage", Value: &user.ProfileImage},
				{Key: "coverImage", Value: &user.CoverImage},
			},
		},
	}
	err = collection.FindOneAndUpdate(c.Request().Context(), &query, &update).Err()

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			return c.String(http.StatusNotFound, "No users found")
		}
		return c.String(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, &user)
}

func deleteUser(c echo.Context) error {

	// TODO
	// get the user
	// 30 days before full deletion.
	// ensure the user has no active subscriptions

	collection := mg.Db.Collection("users")

	id := c.Param("id")
	if id == "" {
		return c.String(http.StatusNotFound, "This user id does not exist")
	}

	// Get the id from the paramaters
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.String(http.StatusNotFound, "This user id is invalid")
	}

	// find and delete the user with the given ID
	query := bson.D{{Key: "_id", Value: &userID}}
	result, err := collection.DeleteOne(c.Request().Context(), &query)

	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}

	// the users might not exist
	if result.DeletedCount < 1 {
		return c.String(http.StatusNotFound, "No user found")
	}

	// the record was deleted
	return c.JSON(http.StatusOK, "User deleted")
}
