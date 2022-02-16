package goutil

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestSdp_1(t *testing.T) {
	agent := FirefoxAgent
	offerFile := "../testing/firefox_sdp_offer.txt"
	offer, err := ioutil.ReadFile(offerFile)
	if err != nil {
		fmt.Println("fail to read offer:", err)
		return
	}

	var desc MediaDesc
	if !desc.Parse(offer) {
		fmt.Println("invalid offer")
		return
	}

	certFile := "../testing/certs/cert.pem"
	if !desc.CreateAnswer(agent, certFile) {
		fmt.Println("invalid offer for answer")
		return
	}
	fmt.Println("firefox answer: ", desc.AnswerSdp())
}

func TestSdp_2(t *testing.T) {
	agent := ChromeAgent
	offerFile := "../testing/chrome_sdp_offer.txt"
	offer, err := ioutil.ReadFile(offerFile)
	if err != nil {
		fmt.Println("fail to read offer:", err)
		return
	}

	var desc MediaDesc
	if !desc.Parse(offer) {
		fmt.Println("invalid offer")
		return
	}

	certFile := "../testing/certs/cert.pem"
	if !desc.CreateAnswer(agent, certFile) {
		fmt.Println("invalid offer for answer")
		return
	}
	fmt.Println("chrome answer: ", desc.AnswerSdp())
}
