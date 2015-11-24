package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

// Save JSON and create the parent directory if it doesn't exist
func saveJSON(filename string, data interface{}) error {
	directory := path.Dir(filename)

	// Create directory if not exist
	if _, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(directory, os.ModeDir|os.ModePerm); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Create JSON data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Save data to file
	if err := ioutil.WriteFile(filename, jsonData, 0600); err != nil {
		return err
	}

	return nil
}

// Load JSON
func loadJSON(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(data); err != nil {
		return err
	}

	return nil
}
