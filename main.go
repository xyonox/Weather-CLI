package main

import (
	"fmt"
	"io"
	"net/http"
)

const (
	ApiServer = "https://api.open-meteo.com/v1/forecast?latitude=48.1374&longitude=11.5755&current=temperature_2m"
)

func main() {
	fmt.Println("---------------------- Weather CLI ----------------------")

	// TODO Wetter API und sich einfach wetter von einer vorbestimmten Stadt zu holen
	// TODO Automatisch überm Standort die Stadt bestimmen
	// TODO über Args eine Stadt bestimmen lassen

	resp, err := http.Get(ApiServer)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
		return
	}(resp.Body)

	fmt.Println(resp.Status)
	fmt.Println(resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))

	fmt.Println("---------------------------------------------------------")
}
