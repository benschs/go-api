package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"

	"github.com/benschs/go-api/rest"
)

func init() {
	rest.AddTypeToRegistry(Message{})
}

type Module struct {
}

type Message struct {
	Message     string                 `json:"message"`
	Sender      Sender                 `json:"sender"`
	ListExample []string               `json:"listExample"`
	MapExample  map[string]interface{} `json:"mapExample"`
}

type Sender struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Verified bool   `json:"verified"`
}

func NewModule() *Module {
	return &Module{}
}

func (b *Module) ListMessages(headers map[string]string) ([]string, error) {
	fmt.Println("List messages called")
	spew.Dump(headers)
	return []string{"List", "messages", "called"}, nil
}

func (b *Module) GetMessage(id int) (string, error) {
	fmt.Printf("Get message %d called\n", id)
	return fmt.Sprintf("Message %d\n", id), nil
}

func (b *Module) CreateMessage(m Message) (string, error) {
	fmt.Println("Create message called")
	spew.Dump(m)
	return "Message created", nil
}

func (b *Module) UploadDocument(f *rest.FileInfo, n string) (string, error) {
	fmt.Println("Upload document called")
	spew.Dump(n)
	spew.Dump(string(f.File))
	return "Document uploaded", nil
}
