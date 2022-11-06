package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	app := NewApplication(os.Args[0], "Demo Application")

	app.Setup = func(ctx context.Context) {
		// Set client options
		clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")

		// Connect to MongoDB
		// TODO: w przykładach jest ctx := context.TODO()
		// poza tym, czy Mongo powinno dostac context z tej funkcji, czy moze powinienem stworzyc "potomny"
		// np. withCancel ?
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatal(err)
		}

		// TODO: gdzie ten defer dac w takim przypadku?
		//defer func() {
		//	if err := client.Disconnect(ctx); err != nil {
		//		log.Fatal(err)
		//	}
		//}()

		// Ping the primary
		if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
			panic(err)
		}

		log.Printf("Successfully connected and pinged.")

		handleRoot := func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello"))
		}

		app.Router.Get("/", handleRoot)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("%v failed to run: %v", app.Name, err)
	}
}
