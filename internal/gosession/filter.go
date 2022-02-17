package gosession

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
)

/*
	FILTERS SYSTEM
	v0.1
*/

func addFiltersToPipeline(c echo.Context, role string, pipelineType string) []bson.M {

	pipeline := []bson.M{}

	if role == "admin" {
		switch pipelineType {
		case "users":
			pipeline = addFiltersToAdminUsersPipeline(c, pipeline)
		case "user":
			fmt.Println("Not created yet")
		default:
			fmt.Println("This pipeline type does not exist")
		}
	}

	return pipeline
}

func addFiltersToAdminUsersPipeline(c echo.Context, pipeline []bson.M) []bson.M {

	// query to find all users with this email TODO limit 1
	email := c.QueryParam("email")
	if email != "" {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"email": email}})
	}

	// query to find all users that are verified
	verified := c.QueryParam("verified")
	if verified != "" {

		verified = strings.ToUpper(verified)
		if verified != "TRUE" && verified != "FALSE" {
			fmt.Println("verified query param is an invalid value")
		}
		res, err := strconv.ParseBool(verified)
		if err != nil {
			fmt.Println("verified query param could not be converted into a bool")
		}
		pipeline = append(pipeline, bson.M{"$match": bson.M{"verified": res}})

	}

	// query to find all the users that have subscriptions
	hasSubscriptions := c.QueryParam("hasSubscriptions")
	if hasSubscriptions != "" {

		hasSubscriptions = strings.ToUpper(hasSubscriptions)
		if hasSubscriptions != "TRUE" && hasSubscriptions != "FALSE" {
			fmt.Println("hasSubscriptions query param is an invalid value")
		}
		res, err := strconv.ParseBool(hasSubscriptions)
		if err != nil {
			fmt.Println(err)
		}

		pipeline = append(pipeline, bson.M{
			"$match": bson.M{"subscriptions.0": bson.M{"$exists": res}},
		})

	}

	// query to find all the users accounts that were created before this datetime
	createdBefore := c.QueryParam("createdBefore")
	if createdBefore != "" {

		layout := "2006-01-02T15:04:05.000Z"
		datetimer, err := time.Parse(layout, createdBefore)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(datetimer)
		pipeline = append(pipeline, bson.M{
			"$match": bson.M{"$and": []interface{}{
				bson.M{"createdAt": bson.M{"$ne": nil}},
				bson.M{"createdAt": bson.M{"$lt": datetimer}},
			}},
		})

	}

	// query to find all the users accounts that were created after this datetime
	createdAfter := c.QueryParam("createdAfter")
	if createdAfter != "" {

		layout := "2006-01-02T15:04:05.000Z"
		datetimer, err := time.Parse(layout, createdAfter)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(datetimer)
		pipeline = append(pipeline, bson.M{
			"$match": bson.M{"$and": []interface{}{
				bson.M{"createdAt": bson.M{"$ne": nil}},
				bson.M{"createdAt": bson.M{"$gt": datetimer}},
			}},
		})

	}

	// query to find all the users accounts that last signed in before this datetime
	lastSignedInBefore := c.QueryParam("lastSignedInBefore")
	if lastSignedInBefore != "" {

		layout := "2006-01-02T15:04:05.000Z"
		datetimer, err := time.Parse(layout, lastSignedInBefore)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(datetimer)
		pipeline = append(pipeline, bson.M{
			"$match": bson.M{"$and": []interface{}{
				bson.M{"lastSignedIn": bson.M{"$ne": nil}},
				bson.M{"lastSignedIn": bson.M{"$lt": datetimer}},
			}},
		})
	}

	// query to find all the users accounts that have purchased premium subscriptions
	hasPurchasedSubscriptions := c.QueryParam("hasPurchasedSubscriptions")
	if hasPurchasedSubscriptions != "" {

		hasPurchasedSubscriptions = strings.ToUpper(hasPurchasedSubscriptions)
		if hasPurchasedSubscriptions != "TRUE" && hasPurchasedSubscriptions != "FALSE" {
			fmt.Println("hasPurchasedSubscriptions query param is an invalid value")
		}
		res, err := strconv.ParseBool(hasPurchasedSubscriptions)
		if err != nil {
			fmt.Println(err)
		}

		if res {
			pipeline = append(pipeline, bson.M{
				"$match": bson.M{
					"subscriptions": bson.M{"$elemMatch": bson.M{"premium": true}},
				},
			})
		} else {
			pipeline = append(pipeline, bson.M{
				"$match": bson.M{
					"subscriptions": bson.M{"$elemMatch": bson.M{"premium": bson.M{"$ne": true}}},
				},
			})
		}
	}

	// query to find all the users accounts that were updated before this datetime
	updatedBefore := c.QueryParam("updatedBefore")
	if updatedBefore != "" {

		layout := "2006-01-02T15:04:05.000Z"
		datetimer, err := time.Parse(layout, updatedBefore)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(datetimer)
		pipeline = append(pipeline, bson.M{
			"$match": bson.M{"$and": []interface{}{
				bson.M{"updatedAt": bson.M{"$ne": nil}},
				bson.M{"updatedAt": bson.M{"$lt": datetimer}},
			}},
		})

	}

	// query to find all the users accounts that were updated after this datetime
	updatedAfter := c.QueryParam("updatedAfter")
	if updatedAfter != "" {

		layout := "2006-01-02T15:04:05.000Z"
		datetimer, err := time.Parse(layout, updatedAfter)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(datetimer)
		pipeline = append(pipeline, bson.M{
			"$match": bson.M{"$and": []interface{}{
				bson.M{"updatedAt": bson.M{"$ne": nil}},
				bson.M{"updatedAt": bson.M{"$gt": datetimer}},
			}},
		})

	}

	// query to find all the users accounts that have active subscriptions
	hasActiveSubscriptions := c.QueryParam("hasActiveSubscriptions")
	if hasActiveSubscriptions != "" {

		hasActiveSubscriptions = strings.ToUpper(hasActiveSubscriptions)
		if hasActiveSubscriptions != "TRUE" && hasActiveSubscriptions != "FALSE" {
			fmt.Println("hasActiveSubscriptions query param is an invalid value")
		}
		res, err := strconv.ParseBool(hasActiveSubscriptions)
		if err != nil {
			fmt.Println(err)
		}

		datetimer := time.Now().UTC()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(datetimer)

		if res {
			pipeline = append(pipeline, bson.M{
				"$match": bson.M{"$and": []interface{}{
					bson.M{"subscriptions": bson.M{"$elemMatch": bson.M{"expiryDate": bson.M{"$lt": datetimer}}}},
				}},
			})
		} else {
			pipeline = append(pipeline, bson.M{
				"$match": bson.M{"$and": []interface{}{
					bson.M{"subscriptions": bson.M{"$elemMatch": bson.M{"expiryDate": bson.M{"$gt": datetimer}}}},
				}},
			})
		}
	}

	// query to find all the users accounts that have denied confirmations
	hasDeniedConfirmations := c.QueryParam("hasDeniedConfirmations")
	if hasDeniedConfirmations != "" {

		hasDeniedConfirmations = strings.ToUpper(hasDeniedConfirmations)
		if hasDeniedConfirmations != "TRUE" && hasDeniedConfirmations != "FALSE" {
			fmt.Println("hasDeniedConfirmations query param is an invalid value")
		}
		res, err := strconv.ParseBool(hasDeniedConfirmations)
		if err != nil {
			fmt.Println(err)
		}

		if res {
			pipeline = append(pipeline, bson.M{
				"$match": bson.M{
					"confirmations": bson.M{"$elemMatch": bson.M{"accepted": false}},
				},
			})
		}
	}

	return pipeline
}
