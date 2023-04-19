package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type sendRequest struct {
	Type string `json:"type"`
	Key  string `json:"key"`
	Msg  int    `json:"msg"`
}

type pollRequest struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}

type commitRequest struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}

type listRequest struct {
	Type string   `json:"type"`
	Keys []string `json:"keys"`
}

func main() {

	n := maelstrom.NewNode()
	kv := maelstrom.NewLinKV(n)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	n.Handle("send", func(msg maelstrom.Message) error {
		var body sendRequest
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		for {
			var offset int
			var oneLog [][2]int
			var ok bool
			log, err := kv.Read(ctx, body.Key)

			if err != nil {
				if maelstrom.ErrorCode(err) != 20 {
					return err
				}
			} else {
				oneLog, ok = log.([][2]int)
				if !ok {
					fmt.Printf("x does not hold a slice of [2]int; this is x: %v\n", log)
				}
			}

			offset = len(oneLog)

			err = kv.CompareAndSwap(ctx, body.Key, log, append(oneLog, [2]int{offset, body.Msg}), true)
			if err == nil {
				return n.Reply(msg, map[string]any{
					"type":   "send_ok",
					"offset": offset,
				})
			}

			if maelstrom.ErrorCode(err) == 22 {
				continue
			} else {
				return err
			}
		}
	})

	n.Handle("poll", func(msg maelstrom.Message) error {
		var body pollRequest
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		for {
			pollMsg := make(map[string][][2]int)

			for key, offset := range body.Offsets {

				log, err := kv.Read(ctx, key)
				if err != nil {
					if maelstrom.ErrorCode(err) != 20 {
						return err
					} else {
						continue
					}
				}

				oneLog, ok := log.([][2]int)

				if !ok {
					fmt.Println("x does not hold a slice of [2]int")
				}

				if offset <= len(oneLog) {
					for i := offset; i < len(oneLog); i++ {
						eachOffset := oneLog[i][0]
						eachMessage := oneLog[i][1]
						pollMsg[key] = append(pollMsg[key], [2]int{eachOffset, eachMessage})
					}
				}
			}
			return n.Reply(msg, map[string]any{
				"type": "poll_ok",
				"msgs": pollMsg,
			})
		}
	})

	n.Handle("commit_offsets", func(msg maelstrom.Message) error {

		var body commitRequest
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		for {
			readResponse, err := kv.Read(ctx, "committedOffsets")
			var committedOffsets map[string]int
			var ok bool
			if err != nil {
				if maelstrom.ErrorCode(err) != 20 {
					return err
				}
			} else {
				committedOffsets, ok = readResponse.(map[string]int)
				if !ok {
					fmt.Printf("x does not hold a map[string]int; this is x: %v\n", readResponse)
				}
			}

			for key, offset := range body.Offsets {
				committedOffsets[key] = offset
			}

			err = kv.CompareAndSwap(ctx, "committedOffsets", readResponse, committedOffsets, true)

			if err == nil {
				return n.Reply(msg, map[string]any{
					"type": "commit_offsets_ok",
				})
			}
			if maelstrom.ErrorCode(err) == 22 {
				continue
			} else {
				return err
			}
		}
	})

	n.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		var body listRequest

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		readResponse, err := kv.Read(ctx, "committedOffsets")
		if err != nil {
			return err
		}

		committedOffsets := readResponse.(map[string]int)

		offsets := make(map[string]int)
		for i := 0; i < len(body.Keys); i++ {
			offsets[body.Keys[i]] = committedOffsets[body.Keys[i]]
		}

		return n.Reply(msg, map[string]any{
			"type":    "list_committed_offsets_ok",
			"offsets": offsets,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
