## How to Integrate this BS
For Irina :3

### .env

First things first, in your .env file for each node add the following vars: `PORT` (the gRPC streaming port the server should run on), `NODES` (list of the other nodes' addresses), and `ID` (something that denotes this specific node).

Ex.:
```
PORT=:50001
NODES=node3:50003,node2:50002
ID=1
```

### Application Module

In your app, you will need to have a module that implements this interface:

```go
type Application interface {
	DetermineConflict(c1, c2 []byte) bool
	Execute(c []byte)
}
```

`DetermineConflict` is used by Caesar module to determine if 2 commands are conflicting (ex., in the case of a datastore application, 2 commands are conflicting if they refer to the same key).

`Execute` is at your discretion -- Caesar calls this function when it and all its predecessors are stable and it can be executed by your application.

Simple example "application":

```go
type SampleApp struct {
	conalg caesar.Conalg
}

func (s *SampleApp) DetermineConflict(c1, c2 []byte) bool {
	return string(c1) == string(c2)
}

func (s *SampleApp) Execute(c []byte) {
  // Do what you want with the stable command payload!
}

func (s *SampleApp) SetConalgModule(m caesar.Conalg) {
	s.conalg = m
}
```

### Initiating the Caesar Module

Your application will only need to call one function:

```go
type Conalg interface {
	Propose(payload []byte)
}
```

Here is how to  integrate the module:

1. Create your Application module:
```go
app := SampleApp{}
```
2. Create the consensus module 

Parameters:
* app - the application module
* envpath - path to the .env file (leave "" if the file is an ".env" in the same directory)
* lvl - the level of the logger (do not set to Debug unless you want to receive thousands of logs)
* analyzerOn - set `true` if you want to log delivery times whenever a request is delivered and receive statistics about how many requests have been delivered etc. every 10 seconds. Otherwise pass `false`

```go
conalg := caesar.InitConalgModule(&app, "", slog.InfoLevel, true)
```
3. Set the consensus module after you've created it
  
```go
app.SetConalgModule(conalg)
```

4. Call `conalg.Propose()` whenever you receive a command. For example my sample app receives a simple HTTP request and 

```go
func main() {
  // Set up modules
	app := SampleApp{}
	conalg := caesar.InitConalgModule(&app, "", slog.InfoLevel, true)
	app.SetConalgModule(conalg)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.POST("/propose", func(c *gin.Context) {
		var json struct {
			Command string `json:"command"`
		}
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

      // Propose your command ->
		app.conalg.Propose([]byte(json.Command))
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("SAMPLEAPP_PORT")
	router.Run(port)
}
```

