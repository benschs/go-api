package rest

import (
	"encoding/json"
	"net/http"
)

type IResponseWriter interface {
	Write(w http.ResponseWriter, content interface{}) error
	WriteError(w http.ResponseWriter, content interface{}) error
}

// StdJSONWriter converts the content to JSON and writes it.
type StdJSONWriter struct {
	settings JSONSettings
}

type JSONSettings struct {
	UseIndent bool
	Prefix    string
	Indent    string
}

func (r *StdJSONWriter) Write(w http.ResponseWriter, content interface{}) error {
	var bytes []byte
	var err error
	if r.settings.UseIndent {
		bytes, err = json.MarshalIndent(content, r.settings.Prefix, r.settings.Indent)
	} else {
		bytes, err = json.Marshal(content)
	}
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytes)
	return err
}

func (r *StdJSONWriter) WriteError(w http.ResponseWriter, content interface{}) error {
	return r.Write(w, content)
}
