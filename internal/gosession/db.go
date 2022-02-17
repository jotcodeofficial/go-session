package gosession

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/sessions"
	redisSessions "github.com/rbcervilla/redisstore/v8"
	"github.com/ulule/limiter/v3"
	redisLimiter "github.com/ulule/limiter/v3/drivers/store/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoInstance contains the Mongo client and database objects
type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

// RedisLimiterInstance contains the Redis limiter client and store objects
type RedisLimiterInstance struct {
	Client redisLimiter.Client
	Store  limiter.Store
}

// RedisSessionInstance contains the Redis session client and store objects
type RedisSessionInstance struct {
	Client *redis.Client
	Store  *redisSessions.RedisStore
}

var mg MongoInstance
var redisLimiterInstance RedisLimiterInstance
var redisSessionInstance RedisSessionInstance

// TODO set password here
func connectToRedisLimiterDatabase() error {
	redisLimiterClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	store, err := redisLimiter.NewStoreWithOptions(redisLimiterClient, limiter.StoreOptions{
		Prefix:   "domain_rate_limiter_",
		MaxRetry: 3,
	})
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// panic: failed to load incr lua script: EOF

	redisLimiterInstance = RedisLimiterInstance{
		Client: redisLimiterClient,
		Store:  store,
	}
	return nil
}

func connectToRedisSessionDatabase() error {
	redisSessionClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// New default RedisStore
	store, err := redisSessions.NewRedisStore(context.Background(), redisSessionClient)
	if err != nil {
		log.Fatal("failed to create redis store: ", err)
	}

	store.KeyPrefix("session_")
	store.Options(sessions.Options{
		Path: "/",
		//Domain: "example.com",
		MaxAge:   86400 * 7,
		HttpOnly: false, // set to httponly false TODO
		//SameSite: true,
		//Secure: true,
	})

	redisSessionInstance = RedisSessionInstance{
		Client: redisSessionClient,
		Store:  store,
	}
	return nil
}

//dbPort := os.Getenv("DOMAIN_API_PORT")

// Connect configures the MongoDB client and initializes the database connection.
// Source: https://www.mongodb.com/blog/post/quick-start-golang--mongodb--starting-and-setup
func connectToDatabase() error {

	// TODO this should be ENV variable
	var mongoURI = "mongodb://" + config.DBUsername + ":" + config.DBPassword + "@" + config.DBHost + ":" + config.DBPort + "/" + config.DBName

	fmt.Println(mongoURI)

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Println("Cannot connect 1...")
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	db := client.Database(config.DBName)
	if err != nil {
		fmt.Println("Cannot connect...")
		return err
	}

	err = client.Ping(context.TODO(), nil)
	// Check the connection
	if err != nil {
		fmt.Println("Cannot ping...")
		return err
	}
	fmt.Printf("Database listening on port:%s", config.DBPort)
	fmt.Println("finished connecting to mongo db")
	mg = MongoInstance{
		Client: client,
		Db:     db,
	}

	return nil
}
