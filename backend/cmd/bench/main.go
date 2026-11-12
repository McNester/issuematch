package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"text/tabwriter"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"issuematch/internal/interactions"
	"issuematch/internal/issues"
	"issuematch/internal/platform"
	"issuematch/internal/users"
)

const resultsFile = "bench/results.csv"

var scenarios = []struct {
	name, path string
}{
	{"Search by text", "/api/issues?q=memory"},
	{"Search by filters", "/api/issues?category=bug&language=Go"},
	{"History", "/api/users/me/interactions"},
	{"Recommendations, baseline", "/api/recommendations?strategy=live&cache=off"},
	{"Recommendations, precomputed", "/api/recommendations?strategy=precomputed&cache=off"},
	{"Recommendations, precomputed and cached", "/api/recommendations"},
}

func main() {
	url := flag.String("url", "http://localhost:8080", "address of the API")
	requests := flag.Int("n", 200, "requests per scenario")
	dropIndexes := flag.Bool("drop-indexes", false, "drop all indexes before the run and recreate them after")
	flag.Parse()

	config, err := platform.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	db, err := platform.ConnectMongo(config.MongoURI)
	if err != nil {
		log.Fatal(err)
	}
	if err := run(db, *url, *requests, *dropIndexes); err != nil {
		log.Fatal(err)
	}
}

func run(db *mongo.Database, url string, requests int, dropIndexes bool) error {
	ctx := context.Background()
	datasetSize, err := describeDataset(ctx, db)
	if err != nil {
		return err
	}
	token, err := logIn(url)
	if err != nil {
		return err
	}

	indexes := "yes"
	if dropIndexes {
		indexes = "no"
		for _, collection := range []string{"users", "issues", "interactions"} {
			if err := db.Collection(collection).Indexes().DropAll(ctx); err != nil {
				return err
			}
		}
		defer recreateIndexes(ctx, db)
	}

	rows := [][]string{}
	for _, scenario := range scenarios {
		average, p95, err := measure(url+scenario.path, token, requests)
		if err != nil {
			return fmt.Errorf("%s: %w", scenario.name, err)
		}
		rows = append(rows, []string{
			scenario.name, datasetSize, indexes, strconv.Itoa(requests),
			fmt.Sprintf("%.2f", average), fmt.Sprintf("%.2f", p95),
		})
	}

	header := []string{"scenario", "dataset size", "indexes", "requests", "average ms", "p95 ms"}
	printTable(header, rows)
	return appendToCSV(header, rows)
}

func describeDataset(ctx context.Context, db *mongo.Database) (string, error) {
	description := ""
	for _, collection := range []string{"issues", "users", "interactions"} {
		count, err := db.Collection(collection).EstimatedDocumentCount(ctx)
		if err != nil {
			return "", err
		}
		description += fmt.Sprintf("%d %s ", count, collection)
	}
	return description[:len(description)-1], nil
}

func logIn(url string) (string, error) {
	credentials := `{"email":"demo@issuematch.local","password":"demo12345"}`
	response, err := http.Post(url+"/api/auth/login", "application/json", bytes.NewBufferString(credentials))
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login as the demo user answered %s; seed the database first", response.Status)
	}
	var session struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(response.Body).Decode(&session)
	return session.Token, err
}

// measure sends the requests one after another and returns the average and
// the 95th percentile of the response time in milliseconds.
func measure(url, token string, requests int) (average, p95 float64, err error) {
	times := make([]float64, requests)
	total := 0.0
	for i := range times {
		request, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return 0, 0, err
		}
		request.Header.Set("Authorization", "Bearer "+token)

		started := time.Now()
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return 0, 0, err
		}
		io.Copy(io.Discard, response.Body)
		response.Body.Close()
		times[i] = float64(time.Since(started).Microseconds()) / 1000
		total += times[i]

		if response.StatusCode != http.StatusOK {
			return 0, 0, fmt.Errorf("server answered %s", response.Status)
		}
	}
	slices.Sort(times)
	return total / float64(requests), times[(requests*95-1)/100], nil
}

func recreateIndexes(ctx context.Context, db *mongo.Database) {
	err := users.NewStore(db).EnsureIndexes(ctx)
	if err == nil {
		err = issues.NewStore(db).EnsureIndexes(ctx)
	}
	if err == nil {
		err = interactions.NewStore(db).EnsureIndexes(ctx)
	}
	if err != nil {
		log.Println("could not recreate the indexes, restart the server to create them:", err)
	}
}

func printTable(header []string, rows [][]string) {
	table := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, row := range append([][]string{header}, rows...) {
		for _, cell := range row {
			fmt.Fprint(table, cell, "\t")
		}
		fmt.Fprintln(table)
	}
	table.Flush()
}

// appendToCSV keeps earlier runs so that the results of all dataset sizes end up in one file.
func appendToCSV(header []string, rows [][]string) error {
	if err := os.MkdirAll("bench", 0o755); err != nil {
		return err
	}
	_, statError := os.Stat(resultsFile)
	file, err := os.OpenFile(resultsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if os.IsNotExist(statError) {
		writer.Write(header)
	}
	writer.WriteAll(rows)
	return writer.Error()
}
