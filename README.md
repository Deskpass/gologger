# Deskpass Logger

This is a simple logger module that ensures log standardization across all
Deskpass services along with some basic log features not provided by Go out of
the box.

Logging features:

- Distinct log level functions: `Debug`, `Info`, `Warn`, `Error`
- Human readable or structured JSON output depending on specified `ENVIRONMENT`
  env var
- Local and remote logging to Papertrail for things running outside of
  Kubernetes such as cloud functions

## Usage

The module provides pretty standard log functions that generally accept a
message and an optional map of metadata. The metadata map is useful for
attaching additional context to a log message. The `Error` function also accepts
an error object:

```go
dlog.Debug(message, meta)
dlog.Info(message, meta)
dlog.Warn(message, meta)
dlog.Error(message, err, meta)
```

Full example with metadata:

```go
dlog.Info("User logged in", map[string]interface{}{
  "user_id": 123,
  "email": "abc123@test.com",
})
```

## Setup

### Standalone

```go
import (
	"github.com/deskpass/gologger"
	"os"
)

dlog := logger.ConfigureLogger(
  "module-name",
  os.Getenv("PAPERTRAIL_LOGGING_URL"),
  os.Getenv("ENVIRONMENT"),
)

dlog.Info("Go forth and log things!")
```

### Lib file

The above will work in a single file, but you'll probably want to use the logger
in a lib file. All the lib file needs to do is configure a logging instance and
export the basic log functions.

```go
package dlog

import (
	"github.com/deskpass/gologger"
	"github.com/joho/godotenv"
	"os"
)

var logInstance *logger.Logger

func init() {
	err := godotenv.Load()

	logInstance = logger.ConfigureLogger(
		"<service name>",
		os.Getenv("PAPERTRAIL_LOGGING_URL"),
		os.Getenv("ENVIRONMENT"),
	)

	if err != nil {
		Info("Error loading .env file", nil)
	}
}

func Debug(message string, meta map[string]interface{}) {
	logInstance.Debug(message, meta)
}

func Info(message string, meta map[string]interface{}) {
	logInstance.Info(message, meta)
}

func Warn(message string, meta map[string]interface{}) {
	logInstance.Warn(message, meta)
}

func Error(message string, err error, meta map[string]interface{}) {
	logInstance.Error(message, err, meta)
}
```

Then in your main file, you can use the logger like so:

```go
package main

import (
  "<module name>/lib/log"
)

dlog.Info("Test log", nil)
```

## Testing

Run `go test` to run the tests. This will test dev/prod log output along with
local/remote log output if Papertrail URL is set.
