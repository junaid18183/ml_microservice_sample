package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ------------------------------------------------------------------------
const (
	// environment variables
	mongoDBConnectionStringEnvVarName = "MONGODB_ENDPOINT"
	mongoDBDatabaseEnvVarName         = "MONGODB_DATABASE"
	mongoDBCollectionEnvVarName       = "MONGODB_COLLECTION"
	applicationPortNumber             = "APP_PORT"
)

// ------------------------------------------------------------------------
// connects to MongoDB
func ConnectDB() (client *mongo.Client, err error) {
	mongodb_endpoint := os.Getenv(mongoDBConnectionStringEnvVarName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	clientOptions := options.Client().ApplyURI(mongodb_endpoint).SetDirect(true)
	c, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	err = c.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB server: %v", err)
	}
	return c, err
}

// ------------------------------------------------------------------------
// Function to check the health of the application and MongoDB connection
func ReportHealth(res http.ResponseWriter, req *http.Request, dbName string) {
	res.Header().Set("Content-Type", "application/json")

	client, err := ConnectDB()
	ctx := context.Background()

	if err != nil {
		log.Printf("MongoDB connection check failed: %v\n", err)
		res.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(res).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
		return
	}
	// appLogger.Printf("Ping to MongoDB databse working")
	defer client.Disconnect(ctx)

	db, err := client.ListDatabaseNames(ctx, bson.D{{Key: "name", Value: dbName}})

	if err != nil {
		log.Printf("MongoDB databse check failed: %v\n", err)
		res.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(res).Encode(map[string]string{"status": "unhealthy", "error": err.Error()})
		return
	}

	if len(db) < 1 {
		log.Printf("MongoDB databse check failed: Can not find Database %s \n", dbName)
		res.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(res).Encode(map[string]string{"status": "unhealthy", "error": "MongoDB databse check failed: Can not find Database"})
		return
	}

	// If MongoDB connection is successful, report as healthy
	log.Printf("Health check passed. MongoDB databse %s Reporting as healthy.", db)
	json.NewEncoder(res).Encode(map[string]string{"status": "healthy"})
}

// ------------------------------------------------------------------------
// CreateCollectionIfNotExists creates a MongoDB collection with the specified name if it doesn't already exist.
func CreateCollectionIfNotExists(dbName string, collectionName string) error {
	client, err := ConnectDB()
	ctx := context.Background()

	if err != nil {
		log.Printf("MongoDB connection check failed: %v\n", err)
		return err
	}

	defer client.Disconnect(ctx)

	collection, err := client.Database(dbName).ListCollectionNames(ctx, bson.D{{Key: "name", Value: collectionName}})

	if err == nil && len(collection) == 1 {
		log.Printf("Collection %s already exists.\n", collection)
		return nil
	}

	log.Printf("Creating Collection %s \n", collectionName)
	collectionOptions := options.CreateCollection()
	collectionOptions.SetCapped(false) // Set collection options as needed

	err = client.Database(dbName).CreateCollection(context.Background(), collectionName, collectionOptions)
	if err != nil {
		return err
	}
	log.Printf("Collection %s Created Successfully in Database %s\n", collectionName, dbName)

	// Add sample data
	sampleData := bson.M{
		"_id":               1,
		"name":              "Feature Set 1",
		"download_link":     "linkFor Download",
		"short_description": "Acronym identification training and development sets for the acronym identification task at SDU@AAAI-21.",
		"long_description":  "Long description",
	}

	_, err = client.Database(dbName).Collection(collectionName).InsertOne(ctx, sampleData)
	if err != nil {
		return err
	}
	log.Println("Sample data added successfully")

	return nil
}

// ------------------------------------------------------------------------
