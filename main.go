package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	WeatherApiServer                 = "https://api.open-meteo.com/v1/forecast?latitude=%v&longitude=%v&hourly=temperature_2m,relative_humidity_2m,apparent_temperature,precipitation_probability,precipitation,rain,showers,snowfall,weather_code,cloud_cover,wind_speed_10m,wind_direction_10m,wind_gusts_10m&current=temperature_2m,relative_humidity_2m,apparent_temperature,precipitation,weather_code,cloud_cover,wind_speed_10m,wind_direction_10m,wind_gusts_10m&forecast_hours=24&timezone=auto"
	LocationApiServer                = "https://api.geoapify.com/v1/geocode/search?text=%v&format=json&apiKey=%v"
	LocationApiServerWithCountryCode = "https://api.geoapify.com/v1/geocode/search?text=%v&filter=countrycode:%v&format=json&apiKey=%v"
)

type GeoapifyResponse struct {
	Results []GeoapifyResult `json:"results"`
}

type GeoapifyResult struct {
	Country          string  `json:"country"`
	State            string  `json:"state"`
	City             string  `json:"city"`
	Latitude         float64 `json:"lat"`
	Longitude        float64 `json:"lon"`
	FormattedAddress string  `json:"formatted"`
}

type WeatherResponse struct {
	Current Current `json:"current"`
	Hourly  Hourly  `json:"hourly"`
}

type Hourly struct {
	Time                     []string  `json:"time"`
	Temperature              []float64 `json:"temperature_2m"`
	RelativeHumidity         []float64 `json:"relative_humidity_2m"`
	ApparentTemperature      []float64 `json:"apparent_temperature"`
	PrecipitationProbability []float64 `json:"precipitation_probability"`
	Precipitation            []float64 `json:"precipitation"`
	Rain                     []float64 `json:"rain"`
	Showers                  []float64 `json:"showers"`
	Snowfall                 []float64 `json:"snowfall"`
	WeatherCode              []int     `json:"weather_code"`
	CloudCover               []float64 `json:"cloud_cover"`
	WindSpeed10m             []float64 `json:"wind_speed_10m"`
	WindDirection10m         []float64 `json:"wind_direction_10m"`
	WindGusts10m             []float64 `json:"wind_gusts_10m"`
}

type Current struct {
	Temperature              float64 `json:"temperature_2m"`
	RelativeHumidity         float64 `json:"relative_humidity_2m"`
	ApparentTemperature      float64 `json:"apparent_temperature"`
	PrecipitationProbability float64 `json:"precipitation_probability"`
}

func main() {
	fmt.Println("---------------------- Weather CLI ----------------------")

	// TODO Bessere Angaben wie das Wetter ist: Regen, Wolken etc

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

	fmt.Printf("\nSearching for: %v\n\n", input)

	splits := strings.SplitN(strings.ToLower(input), ", ", 2)

	var geoRespone *http.Response

	if len(splits) == 2 {
		url := fmt.Sprintf(LocationApiServerWithCountryCode, splits[1], splits[0], os.Getenv("GEOAPIFY_API_KEY"))
		geoRespone, err = http.Get(url)
		if err != nil {
			fmt.Println(err)
			return
		}

	} else {
		url := fmt.Sprintf(LocationApiServer, input, os.Getenv("GEOAPIFY_API_KEY"))
		geoRespone, err = http.Get(url)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	geoBody, err := io.ReadAll(geoRespone.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var geoapifyResponse GeoapifyResponse
	var finallyResult GeoapifyResult

	err = json.Unmarshal(geoBody, &geoapifyResponse)

	fmt.Println("Found the following locations:")
	for key, result := range geoapifyResponse.Results {
		fmt.Printf("%v: %v\n", key, result.FormattedAddress)
	}
	if len(geoapifyResponse.Results) == 1 {
		finallyResult = geoapifyResponse.Results[0]
	} else if len(geoapifyResponse.Results) == 0 {
		fmt.Println("\nNo results found")
		return
	} else if len(geoapifyResponse.Results) > 1 {
		fmt.Println("\nMultiple results found, please select a location")
		fmt.Print("> ")
		if !scanner.Scan() {
			fmt.Println("No input")
		}
		indexInput := strings.TrimSpace(scanner.Text())
		index, err := strconv.Atoi(indexInput)
		if err != nil {
			fmt.Println("Invalid input")
			return
		}
		if index < 0 || index >= len(geoapifyResponse.Results) {
			fmt.Println("Invalid input")
			return
		}
		t := geoapifyResponse.Results[index]
		finallyResult = t
	}

	fmt.Println("\nSelected location:")
	fmt.Printf("%v\nlat: %v\nlong: %v\n\n", finallyResult.FormattedAddress, finallyResult.Latitude, finallyResult.Longitude)

	fmt.Println("Searching for weather:")

	url := fmt.Sprintf(WeatherApiServer, finallyResult.Latitude, finallyResult.Longitude)

	fmt.Println(url)

	resp, err := http.Get(url)
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

	fmt.Println("Weather data fetched\n")

	fmt.Println("Weather for the next 24 hours:")

	var weatherResponse WeatherResponse
	err = json.Unmarshal(body, &weatherResponse)
	fmt.Printf("Current temperature: %v°C\n", weatherResponse.Current.Temperature)
	for key, time := range weatherResponse.Hourly.Time {

		splitTime := strings.SplitN(time, "T", 2)

		fmt.Printf("%v, %v°C\n", splitTime[1], weatherResponse.Hourly.Temperature[key])

	}

	fmt.Println("---------------------------------------------------------")
}
