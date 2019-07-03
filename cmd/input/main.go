package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
)

// Input accepts a POST body and turns the first 6 chars of the body into a
// color event.
type Input struct {
	Addr   string `envconfig:"ADDRESS" default:":8080"`
	Target string `envconfig:"TARGET" required:"true"`
	ce     cloudevents.Client
}

func (s *Input) handler(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		return
	}
	ctx := cloudevents.ContextWithTarget(context.Background(), s.Target)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	color := string(body)
	if len(color) > 6 {
		color = color[:6]
	}
	if color[0] != "#"[0] {
		color = "#" + color
	}

	event := cloudevents.NewEvent(cloudevents.VersionV03)
	event.SetSource("github.com/n3wscott/knative-munich/messaging/input/")
	event.SetType("color")
	event.SetSubject(color)
	_ = event.SetData("{}")

	_, err = s.ce.Send(ctx, event)
	if err != nil {
		panic(err)
	}

	fmt.Println(color, "event sent.")

	resp.WriteHeader(http.StatusOK)
}

func main() {
	ce, err := cloudevents.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	s := &Input{ce: ce}
	if err := envconfig.Process("", s); err != nil {
		panic(err)
	}

	http.HandleFunc("/", s.handler)
	log.Fatal(http.ListenAndServe(s.Addr, nil))
}
