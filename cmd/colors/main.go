package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
)

type Colors struct {
	Target string `envconfig:"SINK" required:"true"`
	ce     cloudevents.Client
}

func randomColor() string {
	return fmt.Sprintf("#%02X%02X%02X", rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func (c *Colors) do() {
	ctx := cloudevents.ContextWithTarget(context.Background(), c.Target)

	for {
		color := randomColor()

		event := cloudevents.NewEvent(cloudevents.VersionV03)
		event.SetSource("github.com/n3wscott/knative-munich/messaging/input/")
		event.SetType("color.random")
		event.SetSubject(color)
		_ = event.SetData(color)

		_, err := c.ce.Send(ctx, event)
		if err != nil {
			panic(err)
		}

		fmt.Println(color, "event sent.")

		time.Sleep(10 * time.Second)
	}
}

func main() {
	ce, err := cloudevents.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	s := &Colors{ce: ce}
	if err := envconfig.Process("", s); err != nil {
		panic(err)
	}

	s.do()
}
