package main

import (
    "fmt"
    "os"
    "io"
    "io/ioutil"
    "net/http"
    "log"
    "log/syslog"
    "encoding/xml"
    "github.com/marpaia/graphite-golang"
)

type SensorData struct {
    Timestamp  string
    Tool       string
    source     string
    Properties Properties
}

type Properties struct {
    Property Property
}

type Property struct {
    Key string
    Value string
}

func handleRequest(rw http.ResponseWriter, req *http.Request) {
    // Only accept POST method
    if req.Method != "POST" {
        http.Error(rw, "Accepts POST Only", http.StatusMethodNotAllowed)
        return
    }

    // Get the XML POST
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        log.Printf("Error: %v", err)
        http.Error(rw, "Could not parse Input", http.StatusBadRequest)
        return
    }

    // Read the xml data into the structures defined above
    var s SensorData
    err = xml.Unmarshal(body, &s)
    if err != nil {
        log.Printf("XML Parse Error: %v", err)
        http.Error(rw, fmt.Sprintf("XML Parse Error: %v", err), http.StatusBadRequest)
        return
    }
    usage := s.Properties.Property.Value
    log.Printf("%v: Received Usage: %v", req.RemoteAddr, usage)

    test := "power.meter.meter1.watts"
    err = sendToGraphite(test, usage)
    if err != nil {
        http.Error(rw, fmt.Sprintf("Error Writing to Graphite: %v", err), http.StatusInternalServerError)
        return
    }

    // Now we submit to Graphite
    io.WriteString(rw, fmt.Sprintf("OK, Received %v\n", usage))
}

func sendToGraphite(k string, v string) error {
    defer log.Printf("Sent %v to Graphite", v)

    Graphite, err := graphite.NewGraphite("127.0.0.1", 2003)
    if err != nil {
        log.Printf("Error Connecting to Graphite: %v", err)
        return err
    }
    err = Graphite.SimpleSend(string(k),string(v))
    if err != nil {
        log.Printf("Error Writing to Graphite: %v", err)
    }
    return err
}

func main() {

    // Setup syslog output
    logwriter, err := syslog.New(syslog.LOG_NOTICE, "powersrv")
    if err == nil {
        log.SetOutput(io.MultiWriter(logwriter, os.Stdout))
    }
    http.HandleFunc("/", handleRequest)
    http.ListenAndServe(":8000", nil)
}
