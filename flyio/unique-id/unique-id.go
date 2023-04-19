package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	uuid "github.com/satori/go.uuid"
)

func main() {

	n := maelstrom.NewNode()

	n.Handle("generate", func(msg maelstrom.Message) error {

		//Unmarshal message
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		//update message type
		body["type"] = "generate_ok"

		id := uuid.NewV4()

		body["id"] = id

		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
