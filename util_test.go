package util

import (
	"log"
	"testing"
)

func TestUtil_1(t *testing.T) {
	agent1 := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:64.0) Gecko/20100101 Firefox/64.0"
	log.Println(ParseAgent(agent1))

	agent2 := "Firefox/64.0"
	log.Println(ParseAgent(agent2))
}
