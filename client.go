//+build wasm,js

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/dennwc/dom"
	"github.com/dennwc/dom/examples/grpc-over-ws/protocol"
	"github.com/dennwc/dom/net/ws"
)

func dialer(s string, dt time.Duration) (net.Conn, error) {
	return ws.Dial(s)
}

// type work struct {
// 	input string,

// }

func main() {
	p1 := dom.Doc.CreateElement("p")
	dom.Body.AppendChild(p1)

	inp := dom.Doc.NewInput("text")
	p1.AppendChild(inp)

	btn := dom.Doc.NewButton("Go!")
	p1.AppendChild(btn)

	//workch is a channel that stores work to be done.
	// workch := make(chan string, 1000)
	ch := make(chan string, 1)
	btn.OnClick(func(_ dom.Event) {
		ch <- inp.Value()
	})

	conn, err := grpc.Dial("ws://localhost:8080/ws", grpc.WithDialer(dialer), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	cli := protocol.AsService(conn)

	printMsg := func(s string) {
		p := dom.Doc.CreateElement("p")
		p.SetTextContent(s)
		dom.Body.AppendChild(p)
	}

	doWork := func(work string) string {
		hash := sha256.New()
		hash.Write([]byte(work))
		output := hex.EncodeToString(hash.Sum(nil))
		log.Println("Did background work:", output)

		p := dom.Doc.CreateElement("p")
		p.SetTextContent("Completed background work.")
		dom.Body.AppendChild(p)

		return output
	}

	ctx := context.Background()
	for {
		// Do work here, add a new dom element that's a counter then create a go routine that does work
		work := <-ch
		printMsg("Doing work")
		work = doWork(work)

		// This is defined on the server.
		txt, err := cli.Hello(ctx, work)
		if err != nil {
			panic(err)
		}

		_, err = cli.GetWork(ctx, work)
		if err != nil {
			panic(err)
		}

		_, err = cli.ReceiveWork(ctx, work)
		if err != nil {
			panic(err)
		}

		printMsg(txt)
	}

	dom.Loop()
}
