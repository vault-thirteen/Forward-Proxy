package main

import "net/http"

func setTargetHeaders(clientHeaders, targetHeaders http.Header) {
	for i, clientHeader := range clientHeaders {
		for _, clientHeaderValue := range clientHeader {
			targetHeaders.Add(i, clientHeaderValue)
		}
	}

	targetHeaders.Del("Proxy-Connection")
	targetHeaders.Del("Proxy-Authenticate")
	targetHeaders.Del("Proxy-Authorization")
	targetHeaders.Del("Connection")
}

func setClientHeaders(clientHeaders http.Header) {
	clientHeaders.Set("User-Agent", "Mozilla/5.0")
}
