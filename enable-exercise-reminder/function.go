package enableExercise

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

// Main prints the JSON encoded "message" field in the body
// of the request or "Hello, World!" if there isn't one.
func Main(w http.ResponseWriter, r *http.Request) {
	eventType := r.URL.Query().Get("type")
	enabled := false
	if strings.EqualFold(eventType, "entered") {
		enabled = true
	}

	err := updateIsEnabledFlag(enabled)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v", err)
	} else {
		fmt.Fprintf(w, "Ok")
	}
}

func updateIsEnabledFlag(isEnabled bool) (err error) {
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

	type ExerciseConfig struct {
		Enabled   bool
		UpdatedAt time.Time
	}
	data := &ExerciseConfig{isEnabled, time.Now()}
	key := datastore.NameKey("ExerciseReminderConfig", "is_enabled", nil)
	if _, err = client.Put(ctx, key, data); err != nil {
		return
	}

	fmt.Println("Updated")
	return
}

func main() {
	err := updateIsEnabledFlag(true)
	if err != nil {
		log.Fatalf("Could not update is enabled flag successfully - error %v", err)
	}
}
