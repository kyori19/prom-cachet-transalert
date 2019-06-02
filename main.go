package main

import (
  "bytes"
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
  Component   string        `json:"id"`
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

func reqcachet(c config, method string, endpoint string, result *string, body []byte) {
  req, err := http.NewRequest(method, c.Domain + endpoint, bytes.NewBuffer(body))
  if err != nil {
    log.Fatal(err)
    return
  }
  req.Header.Set("X-Cachet-Token", c.Token)
  req.Header.Set("Content-Type", "application/json")

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
  reqcachet(c, "GET", "/api/v1/version", &res, nil)
  log.Println(res)
  log.Println("Success.")

  r := gin.Default()

  r.GET("/api/v1/ping", func(context *gin.Context) {
    context.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  })

  r.POST("/api/v1/alert", func(context *gin.Context) {
    var req webhook
    if err := context.ShouldBindJSON(&req); err != nil {
      context.JSON(http.StatusBadRequest, gin.H{
        "error": err.Error(),
      })
      return
    }

    if req.Version != "4" {
      log.Println("[MAIN][WARN][alert] AlertManager json version mismatch.")
      log.Println("[MAIN][WARN][alert] Supported version is 4. (recieved " + req.Version + ")")
    }

    for _, alert := range req.Alerts {
      for _, component := range c.Components {
        if alert.Labels.Instance == component.Alert.Labels.Instance &&
            alert.Labels.Job == component.Alert.Labels.Job &&
            alert.Status == component.Alert.Status {
          data := map[string]int{
            "status": component.Status,
          }
          b, _ := json.Marshal(&data)
          var res string
          reqcachet(c, "PUT", "/api/v1/components/" + component.Component, &res, b)
          log.Println(res)
        }
      }
    }

    context.JSON(http.StatusAccepted, nil)
  })

  r.Run(":9136")
}
