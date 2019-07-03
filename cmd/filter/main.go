package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"

	"github.com/n3wscott/knative-munich/pkg/color"
)

type Filter struct {
	Color string `envconfig:"COLOR" required:"true"` // red, green, or blue
}

type ColorElm struct {
	Red   int `json:"red,omitempty"`
	Blue  int `json:"blue,omitempty"`
	Green int `json:"green,omitempty"`
}

func (f *Filter) receive(event cloudevents.Event, resp *cloudevents.EventResponse) {
	elm, err := color.ParseHexColor(event.Subject())
	if err != nil {
		resp.Error(http.StatusBadRequest, err.Error())
		return
	}
	celm := ColorElm{}
	switch f.Color {
	case "red":
		celm.Red = int(elm.R)
	case "blue":
		celm.Blue = int(elm.B)
	case "green":
		celm.Green = int(elm.G)
	}

	filtered := cloudevents.NewEvent(cloudevents.VersionV03)
	filtered.SetID(event.ID())
	filtered.SetSource("github.com/n3wscott/knative-munich/messaging/filtered/" + strings.ToLower(f.Color))
	filtered.SetType("color." + strings.ToLower(f.Color))
	filtered.SetSubject(event.Subject())
	err = filtered.SetData(celm)
	if err != nil {
		fmt.Println("Error.", err.Error())
		resp.Error(http.StatusBadRequest, err.Error())
		return
	}

	resp.RespondWith(http.StatusAccepted, &filtered)
	fmt.Println(f.Color, event.Subject(), "event sent. \n", filtered.String())
}

func main() {
	ce, err := cloudevents.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	f := &Filter{}
	if err := envconfig.Process("", f); err != nil {
		panic(err)
	}

	log.Fatal(ce.StartReceiver(context.Background(), f.receive))
}
