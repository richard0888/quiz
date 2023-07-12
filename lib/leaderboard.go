package lib

import (
	"context"
	"log"
	"time"

	firebase "firebase.google.com/go"
	"golang.org/x/net/websocket"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Connect struct {
	Type string
}

type User struct {
	Type   string
	Name   string
	Points int64
}

func Leaderboard(ws *websocket.Conn) {
	var done = false
	go func() {
		heartbeat(ws, 5000*time.Millisecond)
		done = true
	}()
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
	it := client.Collection("quiz").Snapshots(ctx)
	for {
		if done {
			return
		}
		snap, err := it.Next()
		if status.Code(err) == codes.DeadlineExceeded {
			return
		}
		if err != nil {
			return
		}
		if snap != nil {
			for {
				doc, err := snap.Documents.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					return
				}
				if doc.Data()["Points"] == nil {
					_, _ = client.Collection("quiz").Doc(doc.Ref.ID).Set(ctx, map[string]interface{}{
						"Points": 0,
					})
				}
				data2 := User{
					"user",
					doc.Ref.ID,
					doc.Data()["Points"].(int64),
				}
				websocket.JSON.Send(ws, data2)
			}
		}
	}
}
