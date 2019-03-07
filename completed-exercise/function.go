package completedExercise

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
)

// Main is the entry point for the GCP cloud function
func Main(w http.ResponseWriter, req *http.Request) {
	exerciseType := req.URL.Query().Get("type")
	idStr := req.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Invalid id '%s' from query parameters provided, could not complete function.", idStr)
		log.Fatalf("Invalid id from query parameters, %v", err)
	}

	err = saveCompletedExercise(exerciseType, int64(id))
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Could not save the completed exercise. Error: %v", err)
		log.Fatalf("Could not save the completed exercise, %v", err)
	}

	fmt.Fprintf(w, "Successfully saved completed exercise: %s", exerciseType)
}

func saveCompletedExercise(exerciseType string, id int64) error {
	projectID := os.Getenv("GCP_PROJECT")
	if projectID == "" {
		projectID = "personal-223023"
	}

	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	defer client.Close()
	if err != nil {
		log.Fatal(err)
	}

	type CompletedExercise struct {
		Type        string
		CompletedAt time.Time
	}
	completedExercise := &CompletedExercise{exerciseType, time.Now()}
	key := datastore.IDKey("CompletedExercises", id, nil)
	if _, err = client.Put(ctx, key, completedExercise); err != nil {
		return err
	}

	return nil
}
