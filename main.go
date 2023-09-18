package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vivsoftorg/enbuild/backend/ml/dao"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Dataset structure
type Dataset struct {
	ID               int    `json:"_id,omitempty" bson:"_id,omitempty"`
	Name             string `json:"name,omitempty" bson:"name,omitempty"`
	Data             string `json:"data,omitempty" bson:"data,omitempty"`
	DownloadLink     string `json:"download_link,omitempty" bson:"download_link,omitempty"`
	ShortDescription string `json:"short_description,omitempty" bson:"short_description,omitempty"`
	LongDescription  string `json:"long_description,omitempty" bson:"long_description,omitempty"`
}

const (
	// environment variables
	mongoDBConnectionStringEnvVarName = "MONGODB_ENDPOINT"
	mongoDBDatabaseEnvVarName         = "MONGODB_DATABASE"
	mongoDBCollectionEnvVarName       = "MONGODB_COLLECTION"
	applicationPortNumber             = "APP_PORT"
)

var (
	appLogger   = log.New(os.Stdout, "[APP] ", log.LstdFlags|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile)
	database    string
	collection  string
)

// Function to get all the ml_datasets from MongoDB
func getAllDatasets(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	mlDatasets, err := connectAndRetrieveDatasets()
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Failed to retrieve datasets: %v\n", err)
		return
	}

	// Encode mlDatasets directly to JSON and write it to the response writer
	err = json.NewEncoder(res).Encode(mlDatasets)
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Failed to encode mlDatasets: %v\n", err)
	}
}

// Function to get a single ml_dataset from MongoDB
func getDataset(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	params := mux.Vars(req)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(res, "Invalid ID", http.StatusBadRequest)
		return
	}

	mlDataset, err := connectAndRetrieveDataset(id)
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Failed to retrieve dataset: %v\n", err)
		return
	}

	if mlDataset == nil {
		http.Error(res, "Not found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(res).Encode(mlDataset)
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Failed to encode mlDataset: %v\n", err)
	}
}

// Function to connect to MongoDB and retrieve all datasets
func connectAndRetrieveDatasets() ([]Dataset, error) {
	client, err := dao.ConnectDB()
	if err != nil {
		errorLogger.Printf("MongoDB connection check failed: %v\n", err)
		return nil, err
	}

	collection := client.Database("enbuild").Collection("MlDataset")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to query datasets from the database: %v", err)
	}
	defer cursor.Close(ctx)

	var datasets []Dataset

	for cursor.Next(ctx) {
		var dataset Dataset
		if err := cursor.Decode(&dataset); err != nil {
			return nil, fmt.Errorf("failed to decode dataset: %v", err)
		}
		datasets = append(datasets, dataset)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("error during cursor iteration: %v", err)
	}
	defer client.Disconnect(context.Background())

	return datasets, nil
}

// Function to connect to MongoDB and retrieve a single dataset by ID
func connectAndRetrieveDataset(id int) (*Dataset, error) {
	client, err := dao.ConnectDB()
	if err != nil {
		errorLogger.Printf("MongoDB connection check failed: %v\n", err)
		return nil, err
	}

	collection := client.Database("enbuild").Collection("MlDataset")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var dataset Dataset

	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&dataset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Dataset not found
		}
		return nil, fmt.Errorf("failed to query dataset from the database: %v", err)
	}
	defer client.Disconnect(context.Background())

	return &dataset, nil
}

// ------------------------------------------------------------------------
func handleOptions(w http.ResponseWriter, r *http.Request) {
	// Respond to pre-flight request with CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With, Authorization")
	w.WriteHeader(http.StatusOK)
}

// ------------------------------------------------------------------------
func main() {
	port := os.Getenv(applicationPortNumber)
	if port == "" {
		port = ":8081"
	}

	database = os.Getenv(mongoDBDatabaseEnvVarName)
	if database == "" {
		database = "enbuild"
		appLogger.Printf("missing environment variable: %s defaulting to %s", mongoDBDatabaseEnvVarName, database)

	}

	collection = os.Getenv(mongoDBCollectionEnvVarName)
	if collection == "" {
		collection = "MlDataset"
		appLogger.Printf("missing environment variable: %s defaulting to %s", mongoDBCollectionEnvVarName, collection)

	}

	mongoDBConnectionString := os.Getenv(mongoDBConnectionStringEnvVarName)
	if mongoDBConnectionString == "" {
		errorLogger.Fatal("Fatal: You need to set the environment variable: ", mongoDBConnectionStringEnvVarName)
	}

	// initilize
	err := dao.CreateCollectionIfNotExists(database, collection)
	if err != nil {
		errorLogger.Fatal("Fatal: Not able to initlize the MongoDB collection: ", collection)

	}

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "CORRELATIONID"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
	)

	router := mux.NewRouter()
	// Handle CORS pre-flight requests
	router.HandleFunc("/api/mlDataset", handleOptions).Methods("OPTIONS")
	router.HandleFunc("/api/mlDataset", getAllDatasets).Methods("GET")
	router.HandleFunc("/api/mlDataset/{id:[0-9]+}", getDataset).Methods("GET")
	router.HandleFunc("/api/healthz", func(res http.ResponseWriter, req *http.Request) {
		dao.ReportHealth(res, req, database)
	}).Methods("GET")
	router.Use(cors)

	appLogger.Println("Starting server at port ", port)

	log.Fatal(http.ListenAndServe(port, (router)))

}

// ------------------------------------------------------------------------
