# Weather CLI

Weather CLI is a small command-line weather application written in Go. It
resolves a location, retrieves the current weather and prints an hourly
forecast for the next 24 hours.

The project was created as a learning project for:

- HTTP requests
- JSON parsing and API responses
- Environment variables
- Formatting structured data for a terminal

## Features

- Search for a location by name
- Optionally restrict the search with a two-letter country code
- Provide the location and country code with command-line flags
- Choose the number of forecast hours with a command-line flag (1–168)
- Select a location when the geocoder returns multiple results
- Display current weather conditions
- Display readable weather descriptions such as sun, rain, showers and snow
- Display an hourly forecast for the next 24 hours
- Format timestamps as `YYYY-MM-DD HH:MM`
- Use metric/European units:
  - temperature: `°C`
  - precipitation: `mm`
  - snowfall: `cm`
  - humidity and precipitation probability: `%`
  - wind speed and gusts: `km/h`
  - wind direction: `°`

## Requirements

- Go 1.25 or newer
- A Geoapify API key
- Internet access

The weather data is provided by [Open-Meteo](https://open-meteo.com/).
Open-Meteo does not require an API key for this application. Location search
uses [Geoapify Geocoding](https://www.geoapify.com/geocoding-api), which
requires an API key.

## Configuration

Create a `.env` file in the project root:

```env
GEOAPIFY_API_KEY=your_geoapify_api_key
```

The `.env` file should not be committed to version control. A key can be
created in the Geoapify dashboard.

## Running the application

Clone the repository and change it into its directory:

```bash
git clone <repository-url>
cd Weather-CLI
```

Install the Go dependency and start the application:

```bash
go mod download
go run .
```

Enter a location when prompted. For example:

```text
Enter a location:
> Köln
```

To restrict the search to a country, provide the location and its two-letter
country code separated by `, `:

```text
> de, Köln
```

If multiple locations are found, the application prints a numbered list and
asks for the corresponding index.

### Command-line flags

The application supports the following flags:

| Flag        | Default | Description                                                                        |
|-------------|---------|------------------------------------------------------------------------------------|
| `-location` | empty   | Location to search for. If omitted, the application asks for it interactively.     |
| `-country`  | empty   | Two-letter country code used to restrict the location search, for example `DE`.    |
| `-hours`    | `24`    | Number of forecast hours to display. Allowed values are `1` to `168` (seven days). |

Use `-h` to display the available flags:

```bash
go run . -h
```

Examples:

```bash
# Search for Cologne and show the default 24-hour forecast
go run . -location "Köln"

# Restrict the search to Germany
go run . -location "Köln" -country DE

# Show a 48-hour forecast
go run . -location "Berlin" -country DE -hours 48
```

The flags can also be used with the compiled binary:

```bash
./weather-cli -location "München" -country DE -hours 12
```

If `-location` is omitted, `-country` and `-hours` are still used while the
location is entered interactively.

## Building a binary

Create an executable with:

```bash
go build -o weather-cli .
```

Run it from the project directory so that the `.env` file can be loaded:

```bash
./weather-cli
```

## How it works

1. The application loads `GEOAPIFY_API_KEY` from `.env`.
2. It sends the entered location to the Geoapify geocoding API.
3. The selected result provides latitude and longitude coordinates.
4. These coordinates are sent to the Open-Meteo forecast API.
5. The JSON response is unmarshaled into `Current` and `Hourly` structs.
6. The `String()` methods format the weather data for terminal output.

Both `Current` and `Hourly` implement Go's `fmt.Stringer` interface. This
keeps the presentation of the weather data in one place and allows the CLI to
print the structs directly with `fmt.Println`.

## Example output

```text
Aktuell (2026-08-12 13:30)
  Aktualisierung: 900 s
  Temperatur:  28.0 °C (gefühlt  26.0 °C)
  Luftfeuchtigkeit:  18 %
  Niederschlag: 0.00 mm
  Wetter: Bewölkt (WMO 2)
  Bewölkung:  56 %
  Wind: 10.8 km/h aus  64° (Böen 26.3 km/h)

Vorhersage
Zeit              Temperatur                  Feuchte  Regenwahrsch.  Niederschlag                                      Wetter        Wolken   Wind
2026-08-12 14:00  28.7 °C (gefühlt 26.6 °C)   16 %        0 %      0.00 mm (Regen 0.00, Schauer 0.00, Schnee 0.00 cm)  Sonnenschein     0 %  Wind 10.8 km/h aus  64° (Böen 26.3 km/h)
```

## Project structure

```text
.
├── main.go       # CLI, API clients, data types and Stringer implementations
├── go.mod        # Go module definition
├── go.sum        # Dependency checksums
├── .env          # API Key
└── README.md     # Project documentation
```

## Current limitations

- The application provides a forecast for up to seven days (168 hours).
- API errors and malformed API responses are only handled partially.
- The application requires a Geoapify API key for every location search.
- There are currently no automated tests.
- The forecast output is intentionally wide and is best viewed in a terminal
  window with sufficient horizontal space.
