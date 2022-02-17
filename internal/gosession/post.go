package gosession

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/go-playground/validator.v9"
)

/*
	POST SYSTEM

	The post system is a way to display post data to the user.
	The users feed will consist of these posts with information provided by an admin.

*/

// DatabasePost is the post that is saved to the database
type DatabasePost struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title           string             `json:"title,omitempty" bson:"title,omitempty" validate:"required,min=2,max=32"`
	Description     string             `json:"description,omitempty" bson:"description,omitempty" validate:"min=1,max=4096"`
	CreatedAt       time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt       time.Time          `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	CreatedByUserID primitive.ObjectID `json:"createdByUserID,onitempty" bson:"createdByUserID,onitempty"`
	History         []PostHistory      `json:"history,omitempty" bson:"history,omitempty"`
	Type            string             `json:"type,omitempty" bson:"type,omitempty" validate:"required"`
	Mode            string             `json:"mode,omitempty" bson:"mode,omitempty" validate:"required"`
	Open            bool               `json:"open,omitempty" bson:"open,omitempty" validate:"required"`
}

// SubmittedPost is the post that is submitted by a user
type SubmittedPost struct {
	Ticker      string `json:"ticker,omitempty" bson:"ticker,omitempty" validate:"required,min=2,max=16"`
	Title       string `json:"title,omitempty" bson:"title,omitempty" validate:"required,min=2,max=32"`
	Description string `json:"description,omitempty" bson:"description,omitempty" validate:"min=1,max=4096"`
	Type        string `json:"type,omitempty" bson:"type,omitempty" validate:"required"`
	Mode        string `json:"mode,omitempty" bson:"mode,omitempty" validate:"required"`
}

// ReturnedPost is the post that is returned to the frontend
type ReturnedPost struct {
	ID          string        `json:"_id,omitempty" bson:"_id,omitempty" validate:"required"`
	Title       string        `json:"title,omitempty" bson:"title,omitempty" validate:"required,min=2,max=32"`
	Description string        `json:"description,omitempty" bson:"description,omitempty" validate:"min=1,max=4096"`
	CreatedAt   time.Time     `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt   time.Time     `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	CreatedBy   string        `json:"createdBy,onitempty" bson:"createdBy,onitempty"`
	History     []PostHistory `json:"history,omitempty" bson:"history,omitempty"`
	Type        string        `json:"type,omitempty" bson:"type,omitempty" validate:"required"`
	Mode        string        `json:"mode,omitempty" bson:"mode,omitempty" validate:"required"`
	Open        bool          `json:"open" bson:"open" validate:"required"`
}

// PostHistory struct
type PostHistory struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title           string             `json:"title,omitempty" bson:"title,omitempty" validate:"required,min=2,max=32"`
	Description     string             `json:"description,omitempty" bson:"description,omitempty" validate:"min=1,max=4096"`
	UpdatedAt       time.Time          `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	CreatedByUserID primitive.ObjectID `json:"createdByUserID,onitempty" bson:"createdByUserID,onitempty"`
	Type            string             `json:"type,omitempty" bson:"type,omitempty" validate:"required"`
	Mode            string             `json:"mode,omitempty" bson:"mode,omitempty" validate:"required"`
	Open            bool               `json:"open,omitempty" bson:"open,omitempty" validate:"required"`
}

// Validation Section --------------------------------------------

// PostValidator - This will validate the post using the structs annotations
type PostValidator struct {
	validator *validator.Validate
}

// Validate - This validates the request body
func (u *PostValidator) Validate(i interface{}) error {
	return u.validator.Struct(i)
}

// ConfigurePostRoutes - Configure all the routes for the posts here
func configurePostRoutes() {

	// Get a specific post
	// Docs: https://docs.mongodb.com/manual/reference/command/find/
	e.GET("/posts/:id", getPost)

	// Get a post page
	e.GET("/posts", getPosts)

	// Insert a new post for a user
	// Docs: https://docs.mongodb.com/manual/reference/command/insert/
	e.POST("/user/posts", postPosts)

	// Update an post record in MongoDB
	// Docs: https://docs.mongodb.com/manual/reference/command/findAndModify/
	//e.PUT("/posts/:id", putPosts)

	// Delete an post from MongoDB
	// Docs: https://docs.mongodb.com/manual/reference/command/delete/
	//e.DELETE("/posts/:id", deletePosts)
}

func getPost(c echo.Context) error {
	// get all records as a cursor

	if c.Param("id") == "" {
		return c.String(http.StatusNotAcceptable, "This is not a valid post id")
	}
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)
	query := bson.M{"_id": objID}

	//db.inventory.find( { qty: { $eq: 20 } } )

	// find all verified or unverified users
	//verified := strings.ToUpper(c.QueryParam("verified"))
	//if verified == "TRUE" || verified == "FALSE" {
	//	query.append(query, bson.M{"verified": verified})
	//}

	/*
		var findLogic Signal
		logicFilter := bson.M{
			"$and": bson.A{
				bson.M{"price": bson.M{"$gt": 100}},
				bson.M{"quantity": bson.M{"$gt": 30}},
			},
		}
	*/

	/* // if accessories exists on document and its not nil
	var findLogic Signal
	logicFilter := bson.M{
		"accessories": bson.M{"$exists: true"},
	}
	*/

	/*
		// QueryParams - example[?after=timestamp&before=timestamp]
		// find all posts before this datetime
		before := c.QueryParam("before")
		// find all posts after this datetime
		after := c.QueryParam("after")
		// find all delete posts
		deleted := c.QueryParam("deleted")
		// sort posts by ascending or decending ticker
		ascending := c.QueryParam("ascending")
		// find all posts that are free or paid
		premium := c.QueryParam("premium")
		// find all posts that are from specified user account name
		userAccount := c.QueryParam("userAccount")
		// find all posts that are active
		active := c.QueryParam("active")
		// find all posts that were profitable
		profitable := c.QueryParam("profitable")
	*/

	var post ReturnedPost

	err := mg.Db.Collection("posts").FindOne(c.Request().Context(), &query).Decode(&post)

	if err != nil {
		fmt.Println(err.Error())
		return c.String(http.StatusNotFound, err.Error())
	}

	// return post in JSON format
	return c.JSON(http.StatusOK, post)
}

func getPosts(c echo.Context) error {
	// get all records as a cursor

	afterID := c.QueryParam("afterID")
	fmt.Println(afterID)
	var query = bson.M{}
	opts := options.Find().
		SetSort(bson.M{"_id": 1}).
		SetLimit(3) // TODO

	if afterID != "" {
		convertedAfterID, err := primitive.ObjectIDFromHex(afterID)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(convertedAfterID)

		query = bson.M{"_id": bson.M{"$gt": convertedAfterID}}
	} else {
		fmt.Println("Should only send back first page")
	}

	//query := bson.M{}

	//db.inventory.find( { qty: { $eq: 20 } } )

	// find all verified or unverified users
	//verified := strings.ToUpper(c.QueryParam("verified"))
	//if verified == "TRUE" || verified == "FALSE" {
	//	query.append(query, bson.M{"verified": verified})
	//}

	/*
		var findLogic Post
		logicFilter := bson.M{
			"$and": bson.A{
				bson.M{"price": bson.M{"$gt": 100}},
				bson.M{"quantity": bson.M{"$gt": 30}},
			},
		}
	*/

	/* // if accessories exists on document and its not nil
	var findLogic Signal
	logicFilter := bson.M{
		"accessories": bson.M{"$exists: true"},
	}
	*/

	/*
		// QueryParams - example[?after=timestamp&before=timestamp]
		// find all posts before this datetime
		before := c.QueryParam("before")
		// find all posts after this datetime
		after := c.QueryParam("after")
		// find all delete posts
		deleted := c.QueryParam("deleted")
		// sort posts by ascending or decending ticker
		ascending := c.QueryParam("ascending")
		// find all posts that are free or paid
		premium := c.QueryParam("premium")
		// find all posts that are from specified trader account name
		traderAccount := c.QueryParam("traderAccount")
		// find all posts that are active
		active := c.QueryParam("active")
		// find all posts that were profitable
		profitable := c.QueryParam("profitable")
	*/

	// TODO dont returned createdBy, instead get the brand name to send back

	//var posts ReturnedPost
	var posts []ReturnedPost
	//.find(index).limit(amount)
	//query := bson.M{}
	//ctx, query, opts
	cur, err := mg.Db.Collection("posts").Find(c.Request().Context(), query, opts)
	if err != nil {
		fmt.Println(err)
	}
	if err = cur.All(c.Request().Context(), &posts); err != nil {
		// Handle error
	}

	// return posts list in JSON format
	return c.JSON(http.StatusOK, posts)
}

// TODO return post after submitted
func postPosts(c echo.Context) error {

	fmt.Println("STARTING post posts")

	postsCollection := mg.Db.Collection("posts")

	store := redisSessionInstance.Store

	session, err := store.Get(c.Request(), "session_")
	if err != nil {
		return c.String(http.StatusNotAcceptable, "failed getting session")
	}

	userID := session.Values["userID"]
	if err != nil {
		return c.String(http.StatusNotFound, "This user id is invalid")
	}

	// Get the user object ID from provided hex
	fmt.Println(userID.(string))
	userObjectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		return c.String(http.StatusNotFound, "This user id is invalid")
	}

	// New post struct
	var submittedPost SubmittedPost

	c.Echo().Validator = &PostValidator{validator: v}

	if err := c.Bind(&submittedPost); err != nil {
		log.Printf("Unable to bind :%v", err)
		return err
	}

	// create the post object to be saved to the database
	var databasePost DatabasePost

	// force MongoDB to always set its own generated ObjectIDs
	databasePost.ID = primitive.NilObjectID

	// force MongoDB to always set time.Now
	databasePost.CreatedAt = time.Now().UTC()
	databasePost.UpdatedAt = time.Now().UTC()

	databasePost.CreatedByUserID = userObjectID

	databasePost.Description = submittedPost.Description
	databasePost.Mode = submittedPost.Mode
	databasePost.Title = submittedPost.Title
	databasePost.Type = submittedPost.Type

	// ensure post history is empty
	databasePost.History = []PostHistory{}

	if err := c.Validate(submittedPost); err != nil {
		log.Printf("Unable to validate the user %+v %v", submittedPost, err)
		// TODO implement: https://medium.com/@apzuk3/input-validation-in-golang-bc24cdec1835
		return c.JSON(http.StatusPartialContent, err.Error())
	}

	insertedPost, err := postsCollection.InsertOne(c.Request().Context(), databasePost)
	if err != nil {
		// TODO - Unable to register, please contact support.
		log.Printf("Unable to insert new post :%v", err)
		return err
	}
	fmt.Println("post created by user: " + " with the id: ")
	fmt.Println(insertedPost.InsertedID)

	// convert databasePost to ReturnedPost then return that

	var returnedPost ReturnedPost
	returnedPost.Title = databasePost.Title
	returnedPost.Description = databasePost.Description
	returnedPost.CreatedAt = databasePost.CreatedAt
	returnedPost.UpdatedAt = databasePost.UpdatedAt
	returnedPost.History = databasePost.History
	returnedPost.Type = databasePost.Type
	returnedPost.Mode = databasePost.Mode
	returnedPost.Open = databasePost.Open

	// return the created User in JSON format
	// TODO remove this so we do not send back the id because that would be bad
	return c.JSON(http.StatusOK, returnedPost)
}
