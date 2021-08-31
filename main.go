package main

import (
	"fmt"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"encoding/json"

	"gitlab.com/amirkerroumi/my-graphql/schema"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to MyGraphQL 2.3 !"))

	mySchema := schema.GetSchema()

	defer r.Body.Close()
	bodyMapString := map[string]string{}
	decoder := json.NewDecoder(r.Body)
    decoder.Decode(&bodyMapString)
	query := bodyMapString["query"]

	params := graphql.Params{Schema: mySchema, RequestString: query}
	res := graphql.Do(params)
	if len(res.Errors) > 0 {
		log.Fatalf("failed to execute graphql operation, errors: %+v", res.Errors)
	}
	rJSON, _ := json.Marshal(res)
	fmt.Fprintf(w, "%s \n", rJSON)
}

func main() {
	// Create Server and Route Handlers
	r := mux.NewRouter()

	r.HandleFunc("/", indexHandler).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start Server
	go func() {
		log.Println("Starting Server")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful Shutdown
	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}
