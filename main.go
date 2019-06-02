package main

import (
  "log"
  "io/ioutil"
  "encoding/json"
  "net/http"

  "github.com/gin-gonic/gin"
)

type labels struct {
  Instance    string        `json:"instance"`
  Job         string        `json:"job"`
}

type alert struct {
  Status      string        `json:"status"`
  Labels      labels        `json:"labels"`
}

type component struct {
  Component   int           `json:"id"`
  Alert       alert         `json:"alert"`
  Status      int           `json:"status"`
}

type config struct {
  Domain      string        `json:"cachet_instance"`
  Token       string        `json:"cachet_token"`
  Components  []component   `json:"components"`
}

type webhook struct {
  Version     string        `json:"version"`
  Alerts      []alert       `json:"alerts"`
}

func reqcachet(c config, endpoint string, result *string) {
  req, _ := http.NewRequest("GET", c.Domain + endpoint, nil)
  req.Header.Set("X-Cachet-Token", c.Token)

  client := new(http.Client)
  res, err := client.Do(req)

  if err != nil {
    log.Fatal(err)
    return
  }

  defer res.Body.Close()
  byteArray, err := ioutil.ReadAll(res.Body)
  if err != nil {
    log.Fatal(err)
    return
  }
  *result = string(byteArray)
}

func main() {
  bytes, err := ioutil.ReadFile("transalert.json")
  if err != nil {
    log.Fatal(err)
    return
  }

  var c config
  if err := json.Unmarshal(bytes, &c); err != nil {
    log.Fatal(err)
    return
  }

  if c.Domain == "" || c.Token == "" {
    log.Fatal("cachet_instance or cachet_token is missing.")
    return
  }

  log.Println("Checking cachet configuration...")
  var res string
  reqcachet(c, "/api/v1/version", &res)
  log.Println(res)
  log.Println("Success.")

  r := gin.Default()

  r.GET("/api/v1/ping", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  })

  r.Run(":9136")
}
