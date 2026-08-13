package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	WeatherApiServer                 = "https://api.open-meteo.com/v1/forecast?latitude=%v&longitude=%v&hourly=temperature_2m,relative_humidity_2m,apparent_temperature,precipitation_probability,precipitation,rain,showers,snowfall,weather_code,cloud_cover,wind_speed_10m,wind_direction_10m,wind_gusts_10m&current=temperature_2m,relative_humidity_2m,apparent_temperature,precipitation,weather_code,cloud_cover,wind_speed_10m,wind_direction_10m,wind_gusts_10m&forecast_hours=%v&timezone=auto"
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
	Time                string  `json:"time"`
	Interval            int     `json:"interval"`
	Temperature         float64 `json:"temperature_2m"`
	RelativeHumidity    float64 `json:"relative_humidity_2m"`
	ApparentTemperature float64 `json:"apparent_temperature"`
	Precipitation       float64 `json:"precipitation"`
	WeatherCode         int     `json:"weather_code"`
	CloudCover          float64 `json:"cloud_cover"`
	WindSpeed10m        float64 `json:"wind_speed_10m"`
	WindDirection10m    float64 `json:"wind_direction_10m"`
	WindGusts10m        float64 `json:"wind_gusts_10m"`
}

func formatTime(value string) string {
	return strings.Replace(value, "T", " ", 1)
}

func (c Current) String() string {
	return fmt.Sprintf(
		"Aktuell (%s)\n"+
			"  Aktualisierung: %d s\n"+
			"  Temperatur: %5.1f °C (gefühlt %5.1f °C)\n"+
			"  Luftfeuchtigkeit: %3.0f %%\n"+
			"  Niederschlag: %4.2f mm\n"+
			"  Wettercode: WMO %d\n"+
			"  Bewölkung: %3.0f %%\n"+
			"  Wind: %4.1f km/h aus %3.0f° (Böen %4.1f km/h)",
		formatTime(c.Time), c.Interval, c.Temperature, c.ApparentTemperature,
		c.RelativeHumidity, c.Precipitation, c.WeatherCode, c.CloudCover,
		c.WindSpeed10m, c.WindDirection10m, c.WindGusts10m,
	)
}

func (h Hourly) String() string {
	var result strings.Builder
	result.WriteString("Vorhersage\n")
	result.WriteString("Zeit              Temperatur       Feuchte  Regenwahrsch.  Niederschlag                 Wetter                                                Wind\n")
	result.WriteString("──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────\n")

	for i, time := range h.Time {
		result.WriteString(fmt.Sprintf("%-17s %4.1f °C (gefühlt %4.1f °C)  %3.0f %%      %3.0f %%      %4.2f mm (Regen %4.2f, Schauer %4.2f, Schnee %4.2f cm)  WMO %3d, Wolken %3.0f %%  %4.1f km/h aus %3.0f° (Böen %4.1f km/h)\n",
			formatTime(time), valueAt(h.Temperature, i), valueAt(h.ApparentTemperature, i),
			valueAt(h.RelativeHumidity, i), valueAt(h.PrecipitationProbability, i),
			valueAt(h.Precipitation, i), valueAt(h.Rain, i), valueAt(h.Showers, i),
			valueAt(h.Snowfall, i), intAt(h.WeatherCode, i),
			valueAt(h.CloudCover, i), valueAt(h.WindSpeed10m, i), valueAt(h.WindDirection10m, i),
			valueAt(h.WindGusts10m, i)))
	}

	return strings.TrimSuffix(result.String(), "\n")
}

func valueAt(values []float64, index int) float64 {
	if index >= 0 && index < len(values) {
		return values[index]
	}
	return 0
}

func intAt(values []int, index int) int {
	if index >= 0 && index < len(values) {
		return values[index]
	}
	return 0
}

func main() {

	location := flag.String("location", "", "Location to search for")
	hoursToShow := flag.Int("hours", 24, "Number of hours to show")
	countryCode := flag.String("country", "", "Country code to search for")

	flag.Parse()

	fmt.Println("Location: ", *location)
	fmt.Println("Hours: ", *hoursToShow)
	fmt.Println("Country: ", *countryCode)

	if *hoursToShow < 1 || *hoursToShow > 24*7 {
		fmt.Println("Invalid number of hours")
		return
	}

	fmt.Println("---------------------- Weather CLI ----------------------")

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
		return
	}
	apiKey := os.Getenv("GEOAPIFY_API_KEY")
	if apiKey == "" {
		fmt.Println("GEOAPIFY_API_KEY is not set")
		return
	}

	scanner := bufio.NewScanner(os.Stdin)

	var input string
	var splits []string

	if *location == "" {
		fmt.Println("Enter a location: ")
		fmt.Print("> ")

		if !scanner.Scan() {
			fmt.Println("No input")
			return
		}

		input = strings.TrimSpace(scanner.Text())

		fmt.Printf("\nSearching for: %v\n\n", input)

		splits = strings.SplitN(strings.ToLower(input), ", ", 2)

		if len(splits) == 2 {
			*countryCode = strings.ToUpper(splits[0])
			*location = strings.TrimSpace(splits[1])
		} else {
			*location = strings.TrimSpace(input)
		}
	}

	var geoResponse *http.Response

	// Escape the location and country code because they cannot fetch chars like "ö"
	*countryCode = url.QueryEscape(strings.ToLower(*countryCode))
	*location = url.QueryEscape(*location)

	if len(splits) == 2 || *countryCode != "" {
		geoUrl := fmt.Sprintf(LocationApiServerWithCountryCode, *location, *countryCode, os.Getenv("GEOAPIFY_API_KEY"))
		fmt.Println(geoUrl)
		geoResponse, err = http.Get(geoUrl)
		if err != nil {
			fmt.Println(err)
			return
		}

	} else {
		geoUrl := fmt.Sprintf(LocationApiServer, *location, os.Getenv("GEOAPIFY_API_KEY"))
		geoResponse, err = http.Get(geoUrl)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	if geoResponse.StatusCode != http.StatusOK {
		fmt.Println("Error while fetching Geoapify Data. Status code: ", geoResponse.Status)
		fmt.Println("Error: ", geoResponse.Body)
		return
	}

	err = geoResponse.Body.Close()
	if err != nil {
		return
	}

	geoBody, err := io.ReadAll(geoResponse.Body)

	if err != nil {
		fmt.Println(err)
		return
	}

	var geoapifyResponse GeoapifyResponse
	var finallyResult GeoapifyResult

	err = json.Unmarshal(geoBody, &geoapifyResponse)
	if err != nil {
		fmt.Println(err)
		return
	}

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
			fmt.Println("Invalid input: Non Integer")
			return
		}
		if index < 0 || index >= len(geoapifyResponse.Results) {
			fmt.Println("Invalid input: Out of range")
			return
		}
		t := geoapifyResponse.Results[index]
		finallyResult = t
	}

	fmt.Println("\nSelected location:")
	fmt.Printf("%v\nlat: %v\nlong: %v\n\n", finallyResult.FormattedAddress, finallyResult.Latitude, finallyResult.Longitude)

	fmt.Println("Searching for weather:")

	weatherUrl := fmt.Sprintf(WeatherApiServer,
		finallyResult.Latitude,
		finallyResult.Longitude,
		*hoursToShow,
	)

	fmt.Println(weatherUrl)

	resp, err := http.Get(weatherUrl)
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

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error while fetching Weather Data. Status code: ", resp.Status)
		fmt.Println("Error: ", resp.Body)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Weather data fetched\n")

	var weatherResponse WeatherResponse
	err = json.Unmarshal(body, &weatherResponse)
	if err != nil {
		fmt.Println("Error while encoding Weather response to JSON: ", err)
	}
	fmt.Println(weatherResponse.Current)
	fmt.Println()
	fmt.Println(weatherResponse.Hourly)

	fmt.Println("---------------------------------------------------------")
}
