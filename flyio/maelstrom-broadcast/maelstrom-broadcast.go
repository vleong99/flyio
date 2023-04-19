package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type TopologyMessageBody struct {
	Type     string              `json:"type"`
	Topology map[string][]string `json:"topology"`
}

func main() {
	seen := make(map[any]int)

	lock := sync.Mutex{}
	var seenValues []any

	var downstreamNodes []string

	n := maelstrom.NewNode()

	n.Handle("broadcast", func(msg maelstrom.Message) error {

		// Unmarshal the message body as an loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Add value to seen
		lock.Lock()
		_, ok := seen[body["message"]]
		if !ok {
			seen[body["message"]]++
			seenValues = append(seenValues, body["message"])
		}
		lock.Unlock()

		if !ok {
			// Propogate message
			go func() {
				for i := 0; i < len(downstreamNodes); i++ {
					if downstreamNodes[i] != msg.Src {
						retryCount := 0
						for {
							ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
							defer cancel()
							resp, err := n.SyncRPC(ctx, downstreamNodes[i], body)
							if err != nil {
								retryCount++
								if retryCount > 3 {
									time.Sleep(1 * time.Second)
									retryCount = 0
								}
								continue
							}

							var body map[string]any
							if err := json.Unmarshal(resp.Body, &body); err != nil {
								continue
							}

							if body["type"].(string) == "broadcast_ok" {
								break
							}
						}
					}
				}
			}()
		}

		// Echo the original message back with the updated message type.
		return n.Reply(msg, map[string]any{
			"type": "broadcast_ok",
		})

	})

	n.Handle("read", func(msg maelstrom.Message) error {
		//Unmarshal message
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		lock.Lock()
		defer lock.Unlock()
		// Return updated message
		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": seenValues,
		})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body TopologyMessageBody

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		downstreamNodes = body.Topology[msg.Dest]

		return n.Reply(msg, map[string]any{
			"type": "topology_ok",
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
