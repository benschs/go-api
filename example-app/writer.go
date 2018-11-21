package main

import (
	"encoding/json"
	"net/http"
)

type CustomWriter struct {
}

func (c *CustomWriter) Write(w http.ResponseWriter, content interface{}) error {
	js, err := json.MarshalIndent(content, "", "   ")
	if err != nil {
		http.Error(w, "JSON Error: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Expires", "0")
	w.Write(js)
	return nil
}

func (c *CustomWriter) WriteError(w http.ResponseWriter, content interface{}) error {
	js, err := json.MarshalIndent(content, "", "   ")
	if err != nil {
		http.Error(w, "JSON Error: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Expires", "0")
	w.Write(js)
	return nil
}
