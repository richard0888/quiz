package lib

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"golang.org/x/net/websocket"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Change struct {
	Type   string
	Name   string
	Answer string
}

func newQuestion(ws *websocket.Conn, client *firestore.Client, ctx context.Context, data Change) {
	iter := client.Collection("questions").Documents(ctx)
	var num = 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return
		}
		num++
	}
	if num <= 0 {
		return
	}
	var random_num = rand.Intn(num)
	_, err := client.Collection("quiz").Doc(data.Name).Set(ctx, map[string]interface{}{
		"Question": random_num,
	})
	if err != nil {
		return
	}
	doc, err := client.Collection("questions").Doc(strconv.Itoa(random_num)).Get(ctx)
	if err != nil {
		return
	}
	var data2 = map[string]interface{}{
		"Type":     "question",
		"Question": doc.Data()["Question"].(string),
	}
	websocket.JSON.Send(ws, data2)
}
func Quiz(ws *websocket.Conn) {
	ctx := context.Background()
	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalln(err)
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()
	var name = ""
	var done = false
	go func() {
		heartbeat(ws, 5000*time.Millisecond)
		done = true
	}()
	go func() {
		var it = client.Collection("quiz").Doc(name).Snapshots(ctx)
		for {
			if done {
				return
			}
			if name != "" {
				snap, err := it.Next()
				if status.Code(err) == codes.DeadlineExceeded {
					return
				}
				if err != nil {
					return
				}
				if !snap.Exists() {
					_, _ = client.Collection("quiz").Doc(name).Set(ctx, map[string]interface{}{
						"Points": 0,
						"Name":   name,
					})
				}
				var data3 = map[string]interface{}{
					"Type":   "points",
					"Points": snap.Data()["Points"].(int64),
				}
				websocket.JSON.Send(ws, data3)
			}
		}
	}()
	for {
		if done {
			return
		}
		var data Change
		websocket.JSON.Receive(ws, &data)
		switch data.Type {
		case "change":
			if data.Name == "" {
				break
			}
			doc2, err := client.Collection("quiz").Doc(data.Name).Get(ctx)
			if err != nil {
				return
			}
			var question = 0
			if doc2.Data()["Question"] != nil {
				question = doc2.Data()["Question"].(int)
			} else {
				newQuestion(ws, client, ctx, data)
				break
			}
			doc, err := client.Collection("questions").Doc(strconv.Itoa(question)).Get(ctx)
			if err != nil {
				return
			}
			var addPoints = 0
			if doc.Data()["Answer"].(string) == data.Answer {
				addPoints = 1
			} else {
				addPoints = -1
			}
			_, err = client.Collection("quiz").Doc(data.Name).Set(ctx, map[string]interface{}{
				"Points": doc2.Data()["Points"].(int64) + int64(addPoints),
				"Name":   data.Name,
			})
			if err != nil {
				log.Fatalln(err)
			}
			var data3 = map[string]interface{}{
				"Type":    "correct",
				"Correct": true,
			}
			websocket.JSON.Send(ws, data3)
			newQuestion(ws, client, ctx, data)
		}
	}
}
