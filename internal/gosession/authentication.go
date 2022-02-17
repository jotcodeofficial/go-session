package gosession

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// MODELS -------------------------------------------------------------------------------

// SignInUser is a struct for when a user tries to sign in
type SignInUser struct {
	Email    string `json:"email" bson:"email" validate:"required,email,min=3"`
	Password string `json:"password" bson:"password" validate:"required,min=10,max=128"`
}

// NewUser is a struct for a new user that was submitted by a user
type NewUser struct {
	Email         string             `json:"email" bson:"email" validate:"required,email,min=3"`
	FirstName     string             `json:"firstName,omitempty" bson:"firstName,omitempty" validate:"omitempty,min=1,max=64,alpha"`
	LastName      string             `json:"lastName,omitempty" bson:"lastName,omitempty" validate:"omitempty,min=1,max=64,alpha"`
	Password      string             `json:"password" bson:"password" validate:"required,min=10,max=128"`
	Confirmations []UserConfirmation `json:"confirmations,omitempty" bson:"confirmations,omitempty"`
}

// SubmitNewUser is a struct for a new user that is created by the system using NewUser
type SubmitNewUser struct {
	Email          string             `json:"email" bson:"email" validate:"required,email,min=3"`
	FirstName      string             `json:"firstName,omitempty" bson:"firstName,omitempty" validate:"omitempty,min=1,max=64,alpha"`
	LastName       string             `json:"lastName,omitempty" bson:"lastName,omitempty" validate:"omitempty,min=1,max=64,alpha"`
	HashedPassword string             `json:"hashedPassword" bson:"hashedPassword" validate:"required,min=10,max=128"`
	Confirmations  []UserConfirmation `json:"confirmations,omitempty" bson:"confirmations,omitempty"`
	CreatedAt      time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt      time.Time          `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	Role           string             `json:"role,omitempty" bson:"role,omitempty"`
}

// SendEmail is used to send to the email service to reset the password
type SendEmail struct {
	SenderName       string `json:"senderName,omitempty" bson:"senderName,omitempty" validate:"required,min=3,max=128"`
	SenderEmail      string `json:"senderEmail" bson:"senderEmail" validate:"required,min=10,max=128"`
	Subject          string `json:"subject" bson:"subject" validate:"required,min=10,max=128"`
	RecipientName    string `json:"recipientName" bson:"recipientName" validate:"required,min=3,max=128"`
	RecipientEmail   string `json:"recipientEmail" bson:"recipientEmail" validate:"required,min=10,max=128"`
	PlainTextContent string `json:"plainTextContent" bson:"plainTextContent" validate:"required,min=10,max=512"`
	HTMLContent      string `json:"htmlContent" bson:"htmlContent" validate:"min=10,max=512"`
	Template         string `json:"template" bson:"template"`
	Code             string `json:"code,omitempty" bson:"code,omitempty" validate:"omitempty,min=1,max=64"`
}

// EmailAuthToken is a struct to confirm the user owns an email address
type EmailAuthToken struct {
	Email     string    `json:"email" bson:"email" validate:"required,email,min=3"`
	Code      string    `json:"code,omitempty" bson:"code,omitempty" validate:"omitempty,min=1,max=64"`
	CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty" validate:"required"`
	ExpiresAt time.Time `json:"expiresAt,omitempty" bson:"expiresAt,omitempty" validate:"required"`
	Mode      string    `json:"mode" bson:"mode" validate:"required"`
}

// NewPassword this is the new password for when a user changes his password
type NewPassword struct {
	Password string `json:"password" bson:"password" validate:"required,min=10,max=128"`
}

// ROUTES --------------------------------------------------------------------------

// ConfigureAuthenticationRoutes - Configure all the routes for authentication here
func configureAuthenticationRoutes() {

	e.POST("/users/:id/change-email", changeEmail, SessionMiddleware("user"))

	// Check if an email already exists on the system
	// TODO rate limit
	e.GET("/auth/emails/:email", getAccountExistViaEmailParam)

	// TODO rate limit
	e.POST("/auth/sign-in", signIn, middleware.BodyLimit("1K"))

	// signs the user out
	e.POST("/auth/sign-out", signOut)

	// registers a user account
	e.POST("/auth/register", register, middleware.BodyLimit("1M"))

	// confirms a user account
	e.POST("/auth/confirm-account/:email/:code", confirmAccount)

	// creates reset password auth token and sends a reset password email
	e.POST("/auth/reset-password/:email", resetPassword)

	// confirms reset password auth token
	// sets a fresh new password for user without needing old password
	// this should be called from frontend after /auth/reset-password/:email
	// this route needs a code from email to confirm if it exists in database
	e.POST("/auth/change-password/:email/:code", changePassword)

	// checks if the user exists in the redis session store
	e.GET("/auth/get-user-via-session", getUserViaSession)

}

// ROUTE FUNCTIONS --------------------------------------------------------------------------

// This route will register a user account and trigger the email service to send a confirmation
// email to confirm the accounts email.
// The email will contain a link, this link will have a unique code that can be typed
// in or a link clicked to run the confirm email route.
func register(c echo.Context) error {

	collection := mg.Db.Collection("users")

	var user NewUser

	c.Echo().Validator = &UserValidator{validator: v}

	if err := c.Bind(&user); err != nil {
		log.Printf("Unable to bind :%v", err)
		return err
	}

	if err := c.Validate(user); err != nil {
		log.Printf("Unable to validate the user %+v %v", user, err)
		return c.JSON(http.StatusPartialContent, err.Error())
	}

	err := isEmailValid(user.Email)
	if err != nil {
		s := fmt.Sprintf("Your email is not valid: %v", user.Email)
		return c.String(http.StatusNotAcceptable, s)
	}

	hashedPassword := hashAndSalt([]byte(user.Password))

	// create the submission user that will saved in the database
	var submitNewUser SubmitNewUser

	submitNewUser.FirstName = user.FirstName
	submitNewUser.LastName = user.LastName
	submitNewUser.Email = user.Email
	submitNewUser.HashedPassword = hashedPassword
	submitNewUser.CreatedAt = time.Now().UTC()
	submitNewUser.UpdatedAt = time.Now().UTC()
	submitNewUser.Confirmations = []UserConfirmation{
		{
			Version: 0,
			Message: "Terms & Conditions accepted",
		},
	}
	submitNewUser.Role = "user"

	if err := c.Validate(submitNewUser); err != nil {
		log.Printf("Unable to validate the user %+v %v", submitNewUser, err)
		// TODO implement: https://medium.com/@apzuk3/input-validation-in-golang-bc24cdec1835
		return c.JSON(http.StatusPartialContent, err.Error())
	}

	_, err = collection.InsertOne(c.Request().Context(), submitNewUser)
	if err != nil {
		log.Printf("Unable to insert new user :%v", err)
		fmt.Println(err)
		return c.JSON(http.StatusPartialContent, "Unable to register, please contact support")
	}

	// At this stage if it fails just tell the user to contact support
	// because their account is already created they just need to verify it.

	// start of the confirm account email process

	// generate password reset token with no expiry
	authToken, err := generateEmailAuthToken(submitNewUser.Email, 0, 0, 0, false, "CONFIRM_ACCOUNT")
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusPartialContent, "To confirm account please contact support")
	}
	// send reset token to the database
	err = addEmailAuthTokenToDatabase(authToken)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	data := SendEmail{
		RecipientEmail:   user.Email,
		RecipientName:    user.FirstName,
		SenderEmail:      "contact@domain.com",
		SenderName:       "Domain",
		Subject:          "Confirm Account",
		PlainTextContent: "Click this link to confirm your account, if you did not request this email then please ignore. ",
		HTMLContent:      "",
		Template:         "CONFIRM_ACCOUNT",
		Code:             authToken.Code,
	}

	// send email
	err = sendEmail(data)
	if err != nil {
		fmt.Println("error at send email")
		fmt.Println(err)
		return c.String(http.StatusNotFound, err.Error())
	}

	// return the created User in JSON format
	return c.JSON(http.StatusOK, "new user has been registered")

}

// This route will confirm the account
func confirmAccount(c echo.Context) error {

	email := c.Param("email")
	code := c.Param("code")

	err := doesAccountExistViaEmailString(email)
	if err != nil {
		return c.String(http.StatusNotFound, "This email does not exist.")
	}
	if code == "" {
		return c.String(http.StatusNotFound, "You have not supplied a valid confirmation code")
	}

	// TODO check if user is already verified

	err = doesEmailAuthTokenExistAndNotExpired(code, email)
	if err != nil {
		return c.String(http.StatusNotFound, "This email auth token is invalid")
	}

	// change the verified to true
	err = verifyUserAccount(email)
	if err != nil {
		return c.String(http.StatusNotFound, "This account could not be verified")
	}

	// delete the token now that it was found
	err = deleteEmailAuthTokensBasedOnMode(code, email, "CONFIRM_ACCOUNT")
	if err != nil {
		return c.String(http.StatusNotFound, "This auth token could not be deleted")
	}

	return c.JSON(http.StatusOK, "User account has been verified")
}

// after the user has clicked the reset password button in the email it will bring them to
// the application with a query param code and let them change their password.
// The application should send the new password with the code to the backend here.
func changePassword(c echo.Context) error {

	// check if the code exists for this email in the emailAuth collection.
	// if it does  and is not expired then update the password for this user and delete the auth from the collection.
	// if it does not or is expired then and delete the expired auth tokens from the collection.
	email := c.Param("email")
	code := c.Param("code")

	err := doesAccountExistViaEmailString(email)
	if err != nil {
		return c.String(http.StatusNotFound, "This email does not exist.")
	}
	if code == "" {
		return c.String(http.StatusNotFound, "You have not supplied a valid confirmation code")
	}

	err = doesEmailAuthTokenExistAndNotExpired(code, email)
	if err != nil {
		return c.String(http.StatusNotFound, "This email auth token is invalid")
	}

	var newPassword NewPassword

	c.Echo().Validator = &UserValidator{validator: v}

	if err := c.Bind(&newPassword); err != nil {
		log.Printf("Unable to bind :%v", err)
		return err
	}

	if err := c.Validate(newPassword); err != nil {
		log.Printf("Unable to validate the newPassword %+v %v", newPassword, err)
		return c.JSON(http.StatusPartialContent, err.Error())
	}

	hashedPassword := hashAndSalt([]byte(newPassword.Password))

	// change the users password to the new hashed and salted one
	err = changeUserPassword(email, hashedPassword)
	if err != nil {
		fmt.Println(err)
		fmt.Println("This account could not be verified")
		return c.String(http.StatusNotFound, "This account could not be verified")
	}

	// delete the token now that it was found
	err = deleteEmailAuthTokensBasedOnMode(code, email, "RESET_PASSWORD")
	if err != nil {
		fmt.Println(err)
		fmt.Println("This auth token could not be deleted")
		return c.String(http.StatusNotFound, "This auth token could not be deleted")
	}

	return c.JSON(http.StatusOK, "User password has been changed")
}

// when a user cannot log in and needs to reset their password they click the reset password
// button from the application which will call this route. This will generate an email auth token
// and store it in the email auth collection. It will then send the password reset email to the user
func resetPassword(c echo.Context) error {

	// QueryParams - example[?oldPassword=oldPassword&newPassword=newPassword]
	// TODO - Implement this route, do it via a form post not a url query parameter
	// send the reset password token via
	// Use the oldPassword to send the reset email
	//oldPassword := c.QueryParam("oldPassword")
	email := c.Param("email")

	if email == "" {
		return c.String(http.StatusNotFound, "You have not supplied a valid email")
	}

	err := doesAccountExistViaEmailString(email)
	if err != nil {
		return c.String(http.StatusNotFound, "This account does not exist")
	}

	// reset password auth token expires after 1 day
	authToken, err := generateEmailAuthToken(email, 0, 0, 1, true, "RESET_PASSWORD")
	if err != nil {
		// custom error code TODO and say contact support
		// TODO
		fmt.Println(err)
		return c.JSON(http.StatusPartialContent, "To confirm account please contact support")
	}
	// send reset token to the database
	err = addEmailAuthTokenToDatabase(authToken)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	// in email send link with that reset token as the query param
	// in mongo have a collection called resetPasswordTokens which stores email and token
	// this way we can delete them after a certain period via continous loop checker
	// https://domain.com/auth/change-password?token=dj845hi48h4h58945h

	data := SendEmail{
		RecipientEmail:   email,
		RecipientName:    email,
		SenderEmail:      "contact@domain.com",
		SenderName:       "Domain",
		Subject:          "Reset Password",
		PlainTextContent: "Click this link to reset your password, if you did not request this email then please ignore. ",
		HTMLContent:      "",
		Template:         "RESET_PASSWORD",
		Code:             authToken.Code,
	}

	// send email
	err = sendEmail(data)
	if err != nil {
		fmt.Println("error at send reset passwordemail")
		fmt.Println(err)
		return c.String(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, "Reset password email has been sent")
}

// getAccountExists - This will check if an account exists on the system via email address
func getAccountExistViaEmailParam(c echo.Context) error {

	collection := mg.Db.Collection("users")

	email := c.Param("email")
	if email == "" {
		return c.String(http.StatusNotFound, "You have not supplied a valid email")
	}

	var user ExistingUser
	query := bson.M{"email": email}
	err := collection.FindOne(c.Request().Context(), &query).Decode(&user)
	if err != nil {
		return c.String(http.StatusNotFound, "This account does not exist on our system")
	}

	return c.String(http.StatusAccepted, "This account already exists")
}

func signOut(c echo.Context) error {
	/*
		sess, err := session.Get("session", c)
		if err != nil {
			return c.JSON(http.StatusForbidden, "Access Denied")
		}

		if sess.Values["id"] == nil || sess.Values["id"] == "" {
			return c.JSON(http.StatusForbidden, "you are already signed out")
		}

		// delete the session
		sess.Options.MaxAge = -1
		sess.Save(c.Request(), c.Response())
	*/
	store := redisSessionInstance.Store
	session, err := store.Get(c.Request(), "session_")
	if err != nil {
		log.Fatal("failed getting session: ", err)
	}
	// Delete session (MaxAge <= 0)
	session.Options.MaxAge = -1
	if err = session.Save(c.Request(), c.Response()); err != nil {
		log.Fatal("failed deleting session: ", err)
	}

	return c.JSON(http.StatusOK, "signed out")
}

// sign in
func signIn(c echo.Context) error {
	ctx := c.Request().Context()
	collection := mg.Db.Collection("users")

	var signInUser SignInUser

	c.Echo().Validator = &UserValidator{validator: v}

	if err := c.Bind(&signInUser); err != nil {
		log.Printf("Unable to bind :%v", err)
		return err
	}

	if err := c.Validate(signInUser); err != nil {
		log.Printf("Unable to validate the user %+v %v", signInUser, err)
		// TODO implement: https://medium.com/@apzuk3/input-validation-in-golang-bc24cdec1835
		return c.JSON(http.StatusPartialContent, err.Error())
	}

	var databaseUser DatabaseUser

	filter := bson.M{"email": signInUser.Email}
	err := collection.FindOne(ctx, filter).Decode(&databaseUser)
	if err != nil {
		fmt.Println(err)
		return c.String(http.StatusNotAcceptable, "This user account does not exist")
	}

	// here we compare the password for this users hashedpassword in the collection for the matching email for this passed in password
	err = bcrypt.CompareHashAndPassword([]byte(databaseUser.HashedPassword), []byte(signInUser.Password))
	if err != nil {
		fmt.Println(err)
		return c.String(http.StatusNotAcceptable, "Incorrect password")
	}

	store := redisSessionInstance.Store

	session, err := store.Get(c.Request(), "session_")
	if err != nil {
		return c.String(http.StatusNotAcceptable, "failed getting session")
	}

	// checks if session exists or is new one
	// only need to set data to its values if it's new
	if session.IsNew {
		// check here if too many sessions exist TODO - set a max limit of sessions per user
		//session.Values["id"] = existingUser.ID.String()

		session.Values["role"] = databaseUser.Role
		session.Values["userID"] = databaseUser.ID.Hex()

		// TODO when checking if too many sessions for this user check the usersessionid exists more than certain times in Redis
		// session.Values["userSessionID"] = existingUser.userSessionID
		session.IsNew = false
		// Save session
		if err = session.Save(c.Request(), c.Response()); err != nil {
			fmt.Println("failed saving session: ", err)
			return c.String(http.StatusNotAcceptable, "failed saving session")
		}
	}

	/*
	   To prevent XSS attack, use HTTP only cookie. HttpOnly is another directive/flag that you can send when setting up cookie. HttpOnly cookies are not accessible to document.cookie API; they are only sent to the server.
	   You should note that by doing so, your own scripts also lose the ability to read cookies:
	*/

	// TODO (for subdomains) Domain: "domain.com"
	// TODO ensure  no more than 3 sessions for this user id

	// set the id value to the request to pass into the getUser route
	//c.SetParamNames("id")
	//c.SetParamValues(existingUser.ID.Hex())

	//return getUser(c)
	user, err := getUserByID(databaseUser.ID.Hex())
	if err != nil {
		return c.String(http.StatusNotAcceptable, "failed getting new user")
	}
	fmt.Println("po")
	return c.JSON(http.StatusOK, user)
	//return c.JSON(http.StatusOK, json.NewEncoder(w).Encode(res))
}

// This will return the user by the sessions stored id,
// if the session is valid and the user exists in the collection
func getUserViaSession(c echo.Context) error {

	fmt.Println("starting get user by session")

	collection := mg.Db.Collection("users")

	store := redisSessionInstance.Store

	session, err := store.Get(c.Request(), "session_")
	if err != nil {
		return c.String(http.StatusNotAcceptable, "failed getting session")
	}

	// in collection:
	// _id = ObjectId("5fb15136c36043c315aec107")

	id := session.Values["userID"]
	if id == nil {
		return c.String(http.StatusNotFound, "This user id is invalid")
	}

	converted, err := primitive.ObjectIDFromHex(id.(string))
	if err != nil {
		return c.String(http.StatusNotFound, "This user id is invalid")
	}
	query := bson.M{"_id": converted}

	var user ExistingUser

	err = collection.FindOne(c.Request().Context(), &query).Decode(&user)

	if err != nil {
		fmt.Println(err.Error())
		return c.String(http.StatusNotFound, err.Error())
	}

	// return user in JSON format
	return c.JSON(http.StatusOK, user)
}

// END ROUTE FUNCTIONS --------------------------------------------------------------------------

// INTERNAL FUNCTIONS --------------------------------------------------------------------------

// this will send an email request to the email service
func sendEmail(email SendEmail) error {

	var endpoint = ""
	if email.Template == "CONFIRM_ACCOUNT" {
		// TODO change in production
		endpoint = "http://127.0.0.1:8081/auth/confirm-account"
	} else if email.Template == "RESET_PASSWORD" {
		// TODO change in production
		endpoint = "http://127.0.0.1:8081/auth/reset-password"
	} else {
		return errors.New("No valid template option supplied")
	}

	byteInfo, err := json.Marshal(email)
	if err != nil {
		fmt.Println(err)
		return errors.New(err.Error())
	}

	resp, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(byteInfo))
	if err != nil {
		fmt.Println(err)
		return errors.New(err.Error())
	}

	client := &http.Client{}
	newResp, err := client.Do(resp)
	if err != nil {
		fmt.Println(err)
		return errors.New(err.Error())
	}

	defer newResp.Body.Close()
	body, err := ioutil.ReadAll(newResp.Body)
	if err != nil {
		fmt.Println(string(body))
		fmt.Println(err)
		return errors.New(err.Error())
	}
	return nil

}

// this will add the supplied email auth token to the email auth collection
func addEmailAuthTokenToDatabase(emailAuthToken EmailAuthToken) error {
	// TODO create this collection in the mongo db
	collection := mg.Db.Collection("emailAuthTokens")

	// TODO verify emailAuthToken is validated here or above
	result, err := collection.InsertOne(context.Background(), emailAuthToken)
	if err != nil {
		log.Printf("Unable to insert new email auth token :%v", err)
		fmt.Println("Unable to insert new email auth token: ")
		fmt.Println(err)
		return err
	}

	fmt.Println("addEmailAuthTokenToDatabase")
	fmt.Println(emailAuthToken)
	fmt.Println("result: ")
	fmt.Println(result)
	return nil
}

// this is used to determine if a email address exists on the system
func doesAccountExistViaEmailString(email string) error {
	collection := mg.Db.Collection("users")

	if email == "" {
		return errors.New("You have not supplied a valid email")
	}

	var user ExistingUser
	query := bson.M{"email": email}
	err := collection.FindOne(context.Background(), &query).Decode(&user)
	if err != nil {
		return errors.New("This email account does not exist on the system")
	}

	return nil
}

// email auth token functions

// this is used to determine if an email auth token exists and is not invalid
func doesEmailAuthTokenExistAndNotExpired(code string, email string) error {
	collection := mg.Db.Collection("emailAuthTokens")

	// we assume the account exists at this stage to save on database operations
	// TODO check if the email is a valid one here
	if code == "" {
		return errors.New("You have not supplied a valid code")
	}
	if email == "" {
		return errors.New("You have not supplied a valid email")
	}

	var token EmailAuthToken
	query := bson.M{"code": code}
	err := collection.FindOne(context.Background(), &query).Decode(&token)
	if err != nil {
		return errors.New("This email auth token does not exist on the system")
	}

	if token.ExpiresAt.IsZero() {
		return nil
	}

	if time.Now().After(token.ExpiresAt) {
		return errors.New("This code is expired")
	}

	return nil
}

func deleteEmailAuthToken(code string, email string) error {

	// we assume this account exists at this point to save on database operations
	if code == "" {
		return errors.New("You have not supplied a valid code")
	}
	if email == "" {
		return errors.New("You have not supplied a valid email")
	}

	// find and delete the signal with the given ID
	query := bson.D{
		{Key: "email", Value: email},
		{Key: "code", Value: code},
	}
	result, err := mg.Db.Collection("emailAuthTokens").DeleteOne(context.Background(), &query)

	if err != nil {
		return errors.New(err.Error())
	}

	if result.DeletedCount < 1 {
		return errors.New("No email auth token found")
	}

	// the record was deleted
	fmt.Println("The auth token was deleted")
	return nil
}

func deleteEmailAuthTokensBasedOnMode(code string, email string, mode string) error {

	// we assume this account exists at this point to save on database operations
	if code == "" {
		return errors.New("You have not supplied a valid code")
	}
	if email == "" {
		return errors.New("You have not supplied a valid email")
	}
	if mode == "" {
		return errors.New("You have not supplied a valid mode")
	}

	// if mode == "CONFIRM_ACCOUNT" - delete all from email with CONFIRM_ACCOUNT
	// if mode == "RESET_PASSWORD" - delete all for email with RESET_PASSWORD

	// find and delete the signal with the given ID
	query := bson.D{
		{Key: "email", Value: email},
		{Key: "mode", Value: mode},
	}
	result, err := mg.Db.Collection("emailAuthTokens").DeleteMany(context.Background(), &query)

	if err != nil {
		return errors.New(err.Error())
	}

	if result.DeletedCount < 1 {
		return errors.New("No email auth tokens found")
	}

	// the record was deleted
	fmt.Println("The auth tokens were deleted")
	return nil

}

func verifyUserAccount(email string) error {

	// we assume the email is valid at this point to save on database operations
	upsert := true
	opts := options.FindOneAndUpdateOptions{
		Upsert: &upsert,
	}

	filter := bson.D{{Key: "email", Value: email}}
	query := bson.D{
		{Key: "$set",
			Value: bson.D{
				{Key: "verified", Value: true},
			},
		},
	}

	err := mg.Db.Collection("users").FindOneAndUpdate(context.Background(), &filter, &query, &opts).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("No account found")
		}
		return errors.New(err.Error())
	}

	fmt.Println("The account should be verified now")
	return nil
}

func changeUserPassword(email string, hashedPassword string) error {

	// we assume the email is valid at this point to save on database operations

	upsert := true
	opts := options.FindOneAndUpdateOptions{
		Upsert: &upsert,
	}

	filter := bson.D{{Key: "email", Value: email}}
	query := bson.D{
		{Key: "$set",
			Value: bson.D{
				{Key: "hashedPassword", Value: hashedPassword},
			},
		},
	}

	err := mg.Db.Collection("users").FindOneAndUpdate(context.Background(), &filter, &query, &opts).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("No account found")
		}
		return errors.New(err.Error())
	}

	fmt.Println("The accounts password should be changed now")
	return nil
}

func generateEmailAuthToken(email string, expiresAddYear int, expiresAddMonth int, expiresAddDay int, expires bool, mode string) (EmailAuthToken, error) {
	var emailAuthToken EmailAuthToken
	randomString, err := generateRandomAuthString()
	if err != nil {
		fmt.Println("cannot assign generated reset auth string to emailAuthToken")
		fmt.Println(err)
	}
	emailAuthToken.Code = randomString
	emailAuthToken.Email = email
	emailAuthToken.CreatedAt = time.Now()
	emailAuthToken.Mode = mode
	if expires {
		emailAuthToken.ExpiresAt = emailAuthToken.CreatedAt.AddDate(expiresAddYear, expiresAddMonth, expiresAddDay)

	} else {
		emailAuthToken.ExpiresAt = time.Time{}
	}

	return emailAuthToken, nil
}

func generateRandomAuthString() (string, error) {
	b := make([]byte, 30)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func getUserRole(c echo.Context) string {

	store := redisSessionInstance.Store

	session, err := store.Get(c.Request(), "session_")
	if err != nil {
		return "anonymous"
	}

	if session.Values["userID"] == nil || session.Values["userID"] == "" {
		return "anonymous"
	}

	// pass in min role to use this route here.
	userRole := session.Values["role"].(string)
	return userRole

}

// hash and salt a password
func hashAndSalt(pass []byte) string {
	hashed, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error when hashing and salting: ")
		fmt.Println(err)
	}
	// Comparing the password with the hash
	err = bcrypt.CompareHashAndPassword(hashed, pass)
	if err != nil {
		fmt.Println("Error when comparing hashed/salted with pass")
		fmt.Println(err)
	}
	return string(hashed)
}

// END INTERNAL FUNCTIONS --------------------------------------------------------------------------

// TODO ROUTES TO BE IMPLEMENTED ------------------------------------------------------------------------

// func deleteAllEmailAuthTokensForEmail(email string) error
// func deleteAllEmailAuthTokensForEmailAndTemplate(email string, template string) error
// func deleteAllEmailAuthTokensForEmail(email string) error
// func deleteAllEmailAuthTokensForEmailAndTemplate(email string, template string) error

// This route will update a users email, but they must be logged in first and confirm new email link before it takes affect
func changeEmail(c echo.Context) error {
	return c.String(http.StatusNotAcceptable, "This route is not configured yet!")
}

// END ROUTES TO BE IMPLEMENTED ------------------------------------------------------------------------
