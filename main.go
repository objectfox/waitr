package main

import (
    "fmt"
    "net/http"
    "time"
    "strconv"
    "regexp"
    "log"
    "os"
)

func handler(w http.ResponseWriter, r *http.Request) {

    // We don't have a favicon
    if r.URL.Path == "/favicon.ico" {
        log.Println("Returned 404 for favicon.ico")
        w.WriteHeader(http.StatusNotFound)
        return
    }

    w.Header().Add("Content-Type", "text/plain")

    // Show our help
    if r.URL.Path == "/" {
        var help = `waitr
-----

A simple waiting service which reponds in the time you specify.

Request option format:

/3m - Wait 3 minutes before returning OK
/10s - Wait 10 seconds before returning OK
/50ms - Wait 50 milliseconds before returning OK

Code by @objectfox at https://github.com/objectfox/waitr
Inspired by @racheldne
`
    log.Println("Got", r.URL.Path, "- Returning help")
    fmt.Fprintf(w, help)
    return
    }


    // Does the request match our format?
    if m, _ := regexp.MatchString("(\\d+)([a-zA-Z]+)", r.URL.Path); m == false {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println("Returned invalid format error for req:", r.URL.Path)
        fmt.Fprintf(w, "Invalid format (/countunits i.e. /15s")
        return
    }

    p, _ := regexp.Compile("([\\d\\.]+)(\\w+)")
    match := p.FindAllStringSubmatch(r.URL.Path, -1)
    
    var num float64
    var ms float64

    // Can we parse the number?
    if s, err := strconv.ParseFloat(match[0][1], 64); err == nil {
        num = s
    } else {
        w.WriteHeader(http.StatusInternalServerError)
        log.Println("Returned number parse error for req:", r.URL.Path)
        fmt.Fprintf(w, "Unable to parse number: %s\n", match[0][1])
        return
    }

    // Can we parse the units?
    var u string
    switch units := match[0][2]; units {
    case "ms":
        u = "millisecond(s)"
        ms = num * 1.0
    case "s":
        u = "second(s)"
        ms = num * 1000.0
    case "m":
        u = "minute(s)"
        ms = num * 1000.0 * 60.0
    default:
        w.WriteHeader(http.StatusInternalServerError)
        log.Println("Returned unit parse error for req:", r.URL.Path)
        fmt.Fprintf(w, "Unable to parse unit (supported are ms, s, m): %s\n", match[0][2])
        return
    }

    // Looks good, let's sleep on it.
    log.Println("Got", r.URL.Path, "- Waiting for", num, u, "(", ms, "ms )")
    time.Sleep(time.Duration(ms) * time.Millisecond)
    fmt.Fprintf(w,"OK\nWaited %v %s (%v ms)\n", num, u, ms)
    log.Println("Finished", r.URL.Path)
}

func main() {
    http.HandleFunc("/", handler)
    log.Println("Starting server")
    
    // Determine port for HTTP service.
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
        log.Printf("defaulting to port %s", port)
    }

    // Start HTTP server.
    log.Printf("listening on port %s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal(err)
    }
}
