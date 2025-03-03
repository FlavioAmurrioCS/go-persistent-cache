# go-persistent-cache

`go-persistent-cache` is a Go library that provides a persistent caching mechanism using SQLite and Gob encoding. It allows you to memoize function results with a specified time-to-live (TTL), storing the results in a SQLite database for efficient retrieval. This can significantly improve the performance of your applications by avoiding redundant computations.

## Features

- Persistent caching using SQLite
- Gob encoding for efficient serialization and deserialization
- Memoization of functions with configurable TTL
- Thread-safe singleton cache instance

## Example Usage

```go
package main

import (
    "fmt"
    "log/slog"
    "os"
    "time"
)

// multiply is a sample function to demonstrate memoization
func multiply(a, b int) int {
    time.Sleep(2 * time.Second)
    return a * b
}

func main() {
    // Configure slog to show debug messages
    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }))
    slog.SetDefault(logger)

    // Memoize the multiply function with a TTL of 5 seconds
    memoizeMultiply := Memoize2(5*time.Second, multiply)

    // Call the memoized function multiple times
    fmt.Println(memoizeMultiply(2, 3))
    fmt.Println(memoizeMultiply(2, 3))
    fmt.Println(memoizeMultiply(2, 3))
    fmt.Println(memoizeMultiply(2, 3))
}
```

## Installation
To install the go-persistent-cache library, use the following command:
```bash
go get github.com/FlavioAmurrioCS/go-persistent-cache
```

## License
This project is licensed under the MIT License.
