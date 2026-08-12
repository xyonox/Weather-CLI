package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	WeatherApiServer                 = "https://api.open-meteo.com/v1/forecast?latitude=48.1374&longitude=11.5755&current=temperature_2m"
	LocationApiServer                = "https://api.geoapify.com/v1/geocode/search?text=%v&format=json&apiKey=%v"
	LocationApiServerWithCountryCode = "https://api.geoapify.com/v1/geocode/search?text=%v&filter=countrycode:%v&format=json&apiKey=%v"
)

func main() {
	fmt.Println("---------------------- Weather CLI ----------------------")

	// TODO Wetter API und sich einfach wetter von einer vorbestimmten Stadt zu holen
	// TODO Automatisch überm Standort die Stadt bestimmen?
	// TODO über Args eine Stadt bestimmen lassen?

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
	apiKey := os.Getenv("GEOAPIFY_API_KEY")
	if apiKey == "" {
		fmt.Println("GEOAPIFY_API_KEY is not set")
		return
	}

	fmt.Println("Enter a location: ")
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)

	if !scanner.Scan() {
		fmt.Println("No input")
		return
	}

	input := strings.TrimSpace(scanner.Text())

	fmt.Println("Searching for: ", input)

	splits := strings.SplitN(strings.ToLower(input), ", ", 2)

	if len(splits) == 2 {
		fmt.Println(fmt.Sprintf(LocationApiServerWithCountryCode, splits[1], splits[0], os.Getenv("GEOAPIFY_API_KEY")))
	} else {
		fmt.Println(fmt.Sprintf(LocationApiServer, input, os.Getenv("GEOAPIFY_API_KEY")))
	}

	resp, err := http.Get(WeatherApiServer)
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
