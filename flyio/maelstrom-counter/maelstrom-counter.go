package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type addBody struct {
	Type  string `json:"type"`
	Delta int    `json:"delta"`
}

func main() {

	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	n.Handle("add", func(msg maelstrom.Message) error {

		var body addBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		for {
			var value int
			_, err := kv.Read(ctx, "globalCounter")
			if err != nil {
				if maelstrom.ErrorCode(err) != 20 {
					return err
				}
				if err = kv.Write(ctx, "globalCounter", 0); err != nil {
					return err
				}
				continue
			} else {
				value, err = kv.ReadInt(ctx, "globalCounter")
				if err != nil {
					return ctx.Err()
				}
			}
			err = kv.CompareAndSwap(ctx, "globalCounter", value, value+body.Delta, false)
			if err == nil {
				return n.Reply(msg, map[string]any{
					"type": "add_ok",
				})
			}
			if maelstrom.ErrorCode(err) == 22 {
				continue
			} else {
				return err
			}
		}

	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var value int

		_, err := kv.Read(ctx, "globalCounter")

		if err != nil {
			if maelstrom.ErrorCode(err) == 20 {
				return n.Reply(msg, map[string]any{
					"type":  "read_ok",
					"value": 0,
				})
			} else {
				return err
			}
		} else {
			value, err = kv.ReadInt(ctx, "globalCounter")
			if err != nil {
				return err
			}
		}

		return n.Reply(msg, map[string]any{
			"type":  "read_ok",
			"value": value,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
