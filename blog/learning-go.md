# Learning Go: A Weather App Adventure

A few months ago, I decided to dive into the world of Go and incorporate it into my current stack. Now, I'm no Go guru, but I can read, write, and understand it well enough to utilize it as effectively as possible.

Today, we'll examine a basic Go program I created that fetches weather data and prints it out to the terminal. The source code for this application can be found [here](https://github.com/mmatongo/sun).

## The Journey Begins

```go
import (
    "encoding/json"
    fl "flag"
    f "fmt"
    "io"
    "net/http"
    "os"
    t "time"
)
```

The application leans heavily on the standard library, a focus that has been pivotal in my learning journey. If I can accomplish a task with the standard library, I should at least try before reaching for external libraries because, well, learning is the goal (go?). Most of those imports may seem familiar, except for the `flag` package, which, if you've played around with CLI applications, you're no stranger to flags.

```go
type Weather struct {
    ResolvedAddress string `json:"resolvedAddress"`
    TimeZone        string `json:"timezone"`
    Days            []struct {
        DatetimeEpoch int64  `json:"datetimeEpoch"`
        Conditions    string `json:"conditions"`
        Descriptions  string `json:"descriptions"`
        Hours         []struct {
            DatetimeEpoch int64  `json:"datetimeEpoch"`
            Conditions    string `json:"conditions"`
            Temp          float64 `json:"temp"`
        } `json:"hours"`
    } `json:"days"`
    CurrentConditions struct {
        DatetimeEpoch int64  `json:"datetimeEpoch"`
        Conditions    string `json:"conditions"`
        Temp          float64 `json:"temp"`
    } `json:"currentConditions"`
}
```

These days, I tend to avoid mixing logic and structs, so I typically create a `./types/types.go` file and keep all my types there while exposing them to the `package main` or wherever I intend to use them.

The struct is a JSON representation of how I expect the data from the API to arrive. Since some of the nested objects arrive as arrays, we need to specify that we're expecting data nested within arrays in our object. The beauty of this declaration style is that you can easily specify keys, and as the data is fetched, it will be matched to those exact keys.

```go
func convertToCelcius(fahrenheit float64) float64 {
    // (°F − 32) × 5/9 = °C
    return (fahrenheit - 32.0) * 5 / 9
}
```

This function is pretty self-explanatory.

## Putting It All Together

The main function looks something like this:

```go
func main() {
    q := "bath"
    var key string

    if len(os.Args) >= 3 {
        q = os.Args[1]
    }

    fl.StringVar(&key, "key", "NULL", "API KEY")
    fl.Parse()

    if key == "" {
        f.Println("No API key provided. Exiting.")
        os.Exit(1)
    }
```

I start by creating a variable `q` which holds the city we're querying. In hindsight, most of this should have made heavier use of the flag package, but for the moment, this suffices. It's a small app, after all.

I then create a `key` variable to store the API key. At this point, we check if the number of received command-line arguments is greater than or equal to 3. The expectation is that the first argument will always be the city name.

The reason for checking `len(os.Args) >= 3` is that when using the `flag` package, it modifies how command-line arguments are interpreted. Without the `flag` package, we would typically expect two arguments (the program name `os.Args[0]` and the city name). However, with the `flag` package, it adds additional arguments, so we now expect at least four arguments: the program name, the city name, the flag name (e.g., "-key"), and the flag value (e.g., "YOUR_API_KEY"). The two additional arguments being the flag name and value introduced by the `flag` package.

```go
    url := f.Sprintf("https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/%s?unitGroup=us&key=%s&contentType=json", q, key)
    res, err := http.Get(url)
    if err != nil {
        panic(err)
    }

    defer res.Body.Close()

    if res.StatusCode != 200 {
        panic("Weather API not available or API Key was not supplied")
    }
```

After that, we format the URL to include the city and API key, then make a `GET` request to the API and do some, let's say, adventurous error handling (don't forget to always defer your response closures).

This was a shoddy assumption, but we want to freak out if we don't get a status 200.

```go
    body, err := io.ReadAll(res.Body)
    if err != nil {
        panic(err)
    }

    var weather Weather
    err = json.Unmarshal(body, &weather)
    if err != nil {
        panic(err)
    }
```

Once we get the data, we use `io.ReadAll`, which will read from a writer until it errors out or gets an EOF (more adventurous error handling).

At this point, we create an instance of the weather struct and use the json package to unmarshal the body into the [address](https://www.golang-book.com/books/intro/8) of `weather` (even more adventurous error handling).

```go
    location, timezone, currentCondition, hours := weather.ResolvedAddress, weather.TimeZone, weather.CurrentConditions, weather.Days[0].Hours
    temp := convertToCelcius(currentCondition.Temp)

    f.Printf("%s, %s: %.0f°C, %s\n",
        location,
        timezone,
        temp,
        currentCondition.Conditions,
    )
```

After that, we get the values that matter to us and assign them to new vars (I prefer to do it this way; it feels cleaner).

We then do our temperature math and print out all the information.

```go
    for _, hour := range hours {
        timeNow := t.Unix(hour.DatetimeEpoch, 0)
        temp := convertToCelcius(hour.Temp)

        if timeNow.Before(t.Now()) {
            continue
        }

        f.Printf("%s - %.0f°C, %s\n",
            timeNow.Format("15:04"),
            temp,
            hour.Conditions,
        )
    }
}
```

Now, because the app will typically receive data for a few hours ahead at any given time, I extend the logic to range over the hours array, convert the time to a human-readable format and the temperature to Celsius, making sure to skip anything that's in the past or right now (`time.Now()`), then print the details. In the end, we get something like this:

```bash
go run main.go lusaka -key=YOUR_API_KEY

Lusaka, Zambia, Africa/Lusaka: 18°C, Rain, Overcast
06:00 - 18°C, Rain, Overcast
07:00 - 21°C, Rain, Partially cloudy
08:00 - 21°C, Rain, Partially cloudy
09:00 - 22°C, Rain, Partially cloudy
10:00 - 22°C, Rain, Partially cloudy
11:00 - 19°C, Rain, Overcast
12:00 - 19°C, Rain, Overcast
```

### Some Known Caveats

- The order of the arguments is important; this can be overcome by using flags more effectively.
- The usage of `Time.Now()` means that if you try to look up the weather for another location, it will base the results off your current time. A potential solution could be to fetch the time zone from the API and adjust accordingly.
- The output is not the neatest. A little formatting could go a long way.
- The code is a little messy. As they say, Rome wasn't built in a day, and neither is clean code.

Despite these caveats, this has been a fantastic learning experience, and I'm excited to continue exploring the world of Go and refining my skills along the way.
