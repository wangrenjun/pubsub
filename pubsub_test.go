package pubsub_test

import (
	"sync"
	"testing"
	"time"

	"github.com/wangrenjun/pubsub"
)

func TestPubsub(t *testing.T) {
	ps := pubsub.New()
	if ps.NumTopic() != 0 {
		t.Errorf("ps.NumTopic() != 0")
	}
	ch1 := make(chan pubsub.TopicMessage)
	ch2 := make(chan pubsub.TopicMessage)
	ch3 := make(chan pubsub.TopicMessage)
	ps.Sub(ch1, "test")
	ps.Sub(ch2, "test")
	ps.Sub(ch3, "test")
	if ps.NumTopic() != 1 {
		t.Errorf("ps.NumTopic() != 1")
	}
	if ps.NumSub("test") != 3 {
		t.Errorf("ps.NumSub() != 3")
	}
	if ps.NumSub("none") != 0 {
		t.Errorf("ps.NumSub() != 0")
	}
	msg1, msg2, msg3 := pubsub.TopicMessage{}, pubsub.TopicMessage{}, pubsub.TopicMessage{}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		msg1 = <-ch1
		wg.Done()
	}()

	go func() {
		msg2 = <-ch2
		wg.Done()
	}()

	go func() {
		msg3 = <-ch3
		wg.Done()
	}()

	n := ps.Pub("hello", "test")
	if n != 3 {
		t.Errorf("n != 3")
	}
	wg.Wait()

	t.Logf("%+v, %+v, %+v, %+v", ps.Topics(), msg1, msg2, msg3)
	if msg1 != msg2 || msg2 != msg3 {
		t.Errorf("msg1 != msg2 || msg2 != msg3")
	}
	ps.UnSub(ch1, "test")
	if ps.NumSub("test") != 2 {
		t.Errorf("ps.NumSub() != 2")
	}
	ps.RemoveTopics("test")
	if ps.NumTopic() != 0 {
		t.Errorf("ps.NumTopic() != 0")
	}
	if ps.NumSub("test") != 0 {
		t.Errorf("ps.NumSub() != 0")
	}
}

func TestPubsub2(t *testing.T) {
	ps := pubsub.New()
	if ps.NumTopic() != 0 {
		t.Errorf("ps.NumTopic() != 0")
	}
	ch1 := make(chan pubsub.TopicMessage)
	ch2 := make(chan pubsub.TopicMessage)
	ch3 := make(chan pubsub.TopicMessage)
	ps.Sub(ch1, "test")
	ps.Sub(ch2, "test")
	ps.Sub(ch3, "test")
	msg1, msg2, msg3 := pubsub.TopicMessage{}, pubsub.TopicMessage{}, pubsub.TopicMessage{}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		msg1 = <-ch1
		wg.Done()
	}()

	go func() {
		msg2 = <-ch2
		wg.Done()
	}()

	go func() {
		msg3 = <-ch3
		wg.Done()
	}()

	ch1 <- pubsub.TopicMessage{"hold", "fake"}
	n := ps.TimedPub("hello", 3*time.Second, "test")
	if n != 2 {
		t.Errorf("n != 2")
	}
	wg.Wait()

	t.Logf("%+v, %+v, %+v, %+v", ps.Topics(), msg1, msg2, msg3)
	if msg1.Topic != "hold" || msg1.Message != "fake" {
		t.Errorf("msg1.Topic != hold")
	}

}
