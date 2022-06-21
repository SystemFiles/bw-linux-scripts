package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/digitalocean/godo"
	"github.com/google/uuid"
	"github.com/systemfiles/bw-linux-scripts/pkg/secrets"
)

const CreateHTTPAddr = "https://api.digitalocean.com/v2/volumes/%s/snapshots"
const OldestDays = "2160h"

func stringToTime(dateString string) time.Time {
	re, _ := regexp.Compile("[0-9]{4}-[0-9]{2}-[0-9]{2}")
	extractedCreationDate := re.FindString(dateString)

	snapshotCreated, err := time.Parse("2006-01-02", extractedCreationDate)
	if err != nil {
		log.Fatal(err.Error())
	}

	return snapshotCreated
}

func getLatestSnapshot(snapshots []godo.Snapshot) godo.Snapshot {
	latest := snapshots[0]
	for _, s := range snapshots[1:] {
		if stringToTime(s.Created).After(stringToTime(latest.Created)) {
			latest = s
		}
	}

	return latest
}

func deleteOldSnapshots(c *godo.Client, ctx context.Context) error {
	opts := &godo.ListOptions{
		Page: 1,
		PerPage: 200,
	}

	snapshots, _, err := c.Snapshots.ListVolume(ctx, opts)
	if err != nil {
		return err
	}

	latestSnapshot := getLatestSnapshot(snapshots)

	t := time.Now()
	d, err := time.ParseDuration(OldestDays)
	if err != nil {
		return err
	}

	var deleted int
	for _, s := range snapshots {
		snapshotCreated := stringToTime(s.Created)

		if t.Sub(snapshotCreated) >= d && s.ID != latestSnapshot.ID {
			log.Printf("deleting old snapshot %s (%s)\n", s.Name, s.Created)
			c.Snapshots.Delete(ctx, s.ID)
			deleted += 1
		}
	}

	log.Printf("Deleted %d snapshots!", deleted)
	return nil
}

type SnapshotCreateCustomOpts struct {
	Name string `json:"name"`
	Tags []string `json:"tags"`
}

func (s *SnapshotCreateCustomOpts) Reader() io.Reader {
	dataRaw, _ := json.Marshal(s)
	return bytes.NewReader(dataRaw)
}

func createNewSnapshot(c *godo.Client, ctx context.Context, volumeUUID, apiKey string) error {
	backup_name := fmt.Sprintf("bw-data-snapshot-%s", strings.Split(uuid.NewString(), "-")[0])
	data := &SnapshotCreateCustomOpts{
		Name: backup_name,
		Tags: []string{
			"backup",
			"automation",
			"bwdata",
		},
	}

	// GODO SDK does not seem to have snapshot creation capabilities
	// Create it using HTTP
	client := &http.Client{
		CheckRedirect: nil,
	}

	req, err := http.NewRequest("POST", fmt.Sprintf(CreateHTTPAddr, volumeUUID), data.Reader())
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !strings.Contains(resp.Status, "200") || !strings.Contains(resp.Status, "201") {
		log.Fatalf("failed to create new snapshot: %s", resp.Status)
	}

	log.Printf("Volume Snapshot created (%s)!", backup_name)
	return nil
}

func main() {
	s := secrets.NewSecrets()

	doClient := godo.NewFromToken(s.ApiKey)
	ctx := context.TODO()

	// delete old snapshots first
	if err := deleteOldSnapshots(doClient, ctx); err != nil {
		log.Fatalf("error occurred when deleting old snapshots. %v", err)
	}

	// create a new snapshot
	if err := createNewSnapshot(doClient, ctx, s.VolumeUUID, s.ApiKey); err != nil {
		log.Fatalf("error occurred when creating new snapshot. %v", err)
	}

	log.Println("Complete!")
}