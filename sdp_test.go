package util

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestSdp_1(t *testing.T) {
	agent := kFirefoxAgent
	offerFile := "../testing/firefox_sdp_offer.txt"
	offer, err := ioutil.ReadFile(offerFile)
	if err != nil {
		log.Warnln("fail to read offer:", err)
		return
	}

	var desc MediaDesc
	if !desc.Parse(offer) {
		log.Warnln("invalid offer")
		return
	}

	certFile := "../testing/certs/cert.pem"
	if !desc.CreateAnswer(agent, certFile) {
		log.Warnln("invalid offer for answer")
		return
	}
	log.Println("firefox answer: ", desc.AnswerSdp())
}

func TestSdp_2(t *testing.T) {
	agent := kChromeAgent
	offerFile := "../testing/chrome_sdp_offer.txt"
	offer, err := ioutil.ReadFile(offerFile)
	if err != nil {
		log.Warnln("fail to read offer:", err)
		return
	}

	var desc MediaDesc
	if !desc.Parse(offer) {
		log.Warnln("invalid offer")
		return
	}

	certFile := "../testing/certs/cert.pem"
	if !desc.CreateAnswer(agent, certFile) {
		log.Warnln("invalid offer for answer")
		return
	}
	log.Println("chrome answer: ", desc.AnswerSdp())
}
