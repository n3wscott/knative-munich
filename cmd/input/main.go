package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
)

type Color struct {
	Color string `json:"color"`
}

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

	// accept strings, quoted strings, json color=value

	if color[0] == "{"[0] {
		c := Color{}
		if json.Unmarshal(body, &c); err != nil {
			panic(err)
		}
		color = c.Color
	} else {
		if color[0] == "\""[0] {
			color, err = strconv.Unquote(color)
			if err != nil {
				panic(err)
			}
		}
	}
	if color[0] != "#"[0] {
		color = "#" + color
	}
	if len(color) > 7 {
		color = color[:7]
	}

	event := cloudevents.NewEvent(cloudevents.VersionV03)
	event.SetSource("github.com/n3wscott/knative-munich/messaging/input/")
	event.SetType("color")
	event.SetSubject(color)
	_ = event.SetData(color)

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
