package main

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/google/uuid"
	cfg "github.com/pobyzaarif/go-config"
)

type Config struct {
	ElasticHost     string `env:"ELASTIC_HOST"`
	ElasticUser     string `env:"ELASTIC_USER"`
	ElasticPassword string `env:"ELASTIC_PASSWORD"`
}

var esIndex = "default-index"

func main() {
	// Init config
	config := Config{}
	err := cfg.LoadConfig(&config)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfgElastic := elasticsearch.Config{
		Addresses: []string{
			config.ElasticHost,
		},
		Username: config.ElasticUser,
		Password: config.ElasticPassword,
	}
	es, err := elasticsearch.NewClient(cfgElastic)
	if err != nil {
		panic(err)
	}

	i, err := es.Info()
	if err != nil {
		panic(err)
	}

	spew.Dump(i)

	// Create index
	err = createIndex(es, esIndex)
	if err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)

	// Insert new document
	docID := uuid.NewString()
	data := map[string]interface{}{
		"id":      docID,
		"title":   "Test Title",
		"content": "This is a test content",
	}
	err = insertDocument(es, esIndex, data)
	if err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)

	// Get document by ID
	getRes, err := getDocument(es, esIndex, docID)
	if err != nil {
		panic(err)
	}
	spew.Dump(getRes)

	time.Sleep(1 * time.Second)

	// Update document by ID
	updatedFields := map[string]interface{}{
		"title": "Updated Test Title",
	}
	err = updateDocument(es, esIndex, docID, updatedFields)
	if err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)

	// Get updated document by ID
	getRes, err = getDocument(es, esIndex, docID)
	if err != nil {
		panic(err)
	}
	spew.Dump(getRes)

	time.Sleep(1 * time.Second)

	// Delete document by ID
	err = deleteDocument(es, esIndex, docID)
	if err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)

	// Try to get deleted document by ID
	getRes, err = getDocument(es, esIndex, docID)
	if err != nil {
		panic(err)
	}
	spew.Dump(getRes)

}

func createIndex(es *elasticsearch.Client, indexName string) error {
	res, err := es.Indices.Create(indexName)
	if err != nil {
		return err
	}
	return res.Body.Close()
}

func insertDocument(es *elasticsearch.Client, indexName string, document map[string]interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(document); err != nil {
		return err
	}

	res, err := es.Index(
		indexName,
		&buf,
		es.Index.WithRefresh("true"),
		es.Index.WithDocumentID(document["id"].(string)),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func getDocument(es *elasticsearch.Client, indexName string, documentID string) (map[string]interface{}, error) {
	var getRes map[string]interface{}
	res, err := es.Get(indexName, documentID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&getRes); err != nil {
		return nil, err
	}

	return getRes, nil
}

func updateDocument(es *elasticsearch.Client, indexName string, documentID string, updatedFields map[string]interface{}) error {
	updateBody := map[string]interface{}{
		"doc": updatedFields,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(updateBody); err != nil {
		return err
	}

	res, err := es.Update(
		indexName,
		documentID,
		&buf,
		es.Update.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func deleteDocument(es *elasticsearch.Client, indexName string, documentID string) error {
	res, err := es.Delete(
		indexName,
		documentID,
		es.Delete.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
