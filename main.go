package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    // Load .env file
    err := godotenv.Load()
    if err != nil {
        panic("Error loading .env file")
    }

    // Set MongoDB client options
    clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URL"))

    // Connect to MongoDB
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        panic(err)
    }

    // Check the connection
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        panic(err)
    }

    fmt.Println("Connected to MongoDB!")

    // Get collection
    collection := client.Database(os.Getenv("MONGO_DB")).Collection("submissions")

    // Get current time at the start of the day
    now := time.Now()
    midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

    // Set filter to match group "communication"
    filter := bson.M{"group": "communication"}

    // Get a cursor for the matched documents
    cursor, err := collection.Find(context.TODO(), filter)
    if err != nil {
        panic(err)
    }
    fmt.Println("Found documents!")

    // Iterate through each document
    for cursor.Next(context.TODO()) {
        var elem struct{
            Data struct{
                StatusLMD string `bson:"statusLMD"`
                InvoicingDateLMD string `bson:"invoicingDateLMD"`
                AlsoMarketingProjectNumberLMD string `bson:"alsoMarketingProjectNumberLMD"`
                SendToLMD string `bson:"sendToLMD"`
            } `bson:"data"`
        }
        err := cursor.Decode(&elem)
        if err != nil {
            panic(err)
        }

        // Parse date string
        if len(elem.Data.InvoicingDateLMD) > 33 {
            var datestring = elem.Data.InvoicingDateLMD[:33]
	
            layout := "Mon Jan 02 2006 15:04:05 MST-0700"
            
            t, err := time.Parse(layout, datestring)
            if err != nil {
                fmt.Println("Error:", err)
            }
          

            fmt.Printf("Before midnight: %t\n, Status: %t\n project:%t\n", t.After(midnight), elem.Data.StatusLMD == "FUTURE INVOICE", elem.Data.AlsoMarketingProjectNumberLMD == "6110CH232104")
            // If date is today or earlier and statusLMD is "FUTURE INVOICE"
            if t.After(midnight) && elem.Data.StatusLMD == "FUTURE INVOICE" && elem.Data.AlsoMarketingProjectNumberLMD == "6110CH232104"{
                // Update statusLMD to "OK FOR INVOICING"
                update := bson.D{{"$set", bson.D{{"data.statusLMD", "OK FOR INVOICING"}}}}
                _, err = collection.UpdateOne(context.TODO(), bson.M{"_id": cursor.Current.Lookup("_id")}, update)
                if err != nil {
                    panic(err)
                }
            }
        } 
    }

    if err = cursor.Err(); err != nil {
        panic(err)
    }

    cursor.Close(context.TODO())
    fmt.Println("Updates complete!")
}
