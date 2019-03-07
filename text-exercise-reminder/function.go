package exerciseReminder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

// Main is the entry point for GCP cloud function
func Main(w http.ResponseWriter, req *http.Request) {
	result, err := sendText()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v", err)
	} else {
		fmt.Fprintf(w, "Result of exercise reminder: '%v'", result)
	}
}

func sendText() (string, error) {
	accountSid := "AC9865dc0f114ce520d508c8d5d3ac4d2d"
	authToken := "3fb03bcc1334a0699bd057cac48e29d5"
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + accountSid + "/Messages.json"

	projectID := os.Getenv("GCP_PROJECT")
	if projectID == "" {
		projectID = "personal-223023"
	}

	completedTextBaseURL := os.Getenv("COMPLETED_TEXT_URL")
	if completedTextBaseURL == "" {
		completedTextBaseURL = "https://us-central1-personal-223023.cloudfunctions.net/completed-exercise"
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
	exerciseConfig := &ExerciseConfig{}
	key := datastore.NameKey("ExerciseReminderConfig", "is_enabled", nil)
	err = client.Get(ctx, key, exerciseConfig)
	if err != nil {
		return "Did not send text, could not get flag from datastore.", nil
	}

	if !exerciseConfig.Enabled {
		fmt.Printf("IsEnabled config was set to %v.\n", exerciseConfig)
		return "Did not send text, exercise reminder is currently disabled.", nil
	}

	type Exercise struct {
		Text string
		Type string
	}
	exercises := [4]Exercise{
		Exercise{"Max push ups or 2 sets of 60% of max.", "Upperbody"},
		Exercise{"Ab work out of choice for 2-3 minutes.", "Abs"},
		Exercise{"15 Body weight squats, 16 Lunges, 15 calf raises. Twice for extra credit.", "Lowerbody"},
		Exercise{"Stretch for 5 minutes.", "Misc"},
	}

	rand.Seed(time.Now().Unix())
	if rand.Intn(2) == 1 {
		exercise := exercises[rand.Intn(len(exercises))]

		var completedExerciseTextBuilder strings.Builder
		completedExerciseTextBuilder.WriteString(exercise.Text)
		completedExerciseTextBuilder.WriteString("\nCompleted? Hit: ")
		completedExerciseTextBuilder.WriteString(completedTextBaseURL)
		completedExerciseTextBuilder.WriteString("?type=")
		completedExerciseTextBuilder.WriteString(exercise.Type)
		completedExerciseTextBuilder.WriteString("&id=")
		completedExerciseTextBuilder.WriteString(strconv.Itoa(rand.Int()))
		completedExerciseText := completedExerciseTextBuilder.String()

		fmt.Println("Sending text for exercise: " + completedExerciseText)

		msgData := url.Values{}
		msgData.Set("To", "+19526490887")
		msgData.Set("From", "+16123244132")
		msgData.Set("Body", completedExerciseText)
		msgDataReader := *strings.NewReader(msgData.Encode())

		client := &http.Client{}
		req, _ := http.NewRequest("POST", urlStr, &msgDataReader)
		req.SetBasicAuth(accountSid, authToken)
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, _ := client.Do(req)
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var data map[string]interface{}
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&data)
			fmt.Printf("Response from twilio - status: %s - data: %v.\n", resp.Status, data)
			return "Did send text.", err
		}
		err = fmt.Errorf("Bad or no Response from twilio: %s", resp.Status)
		return "Did not send text, invalid twilio response encountered.", err
	}

	return "Will not send a exercise reminder text for this invocation.", nil
}

func main() {
	result, err := sendText()
	if err != nil {
		log.Fatalf("Could not complete sendText successfully - error %v", err)
	}

	fmt.Printf("Result of exercise reminder: '%v'", result)
}
