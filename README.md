# Chadango
Chadango is a Go library for interacting with Chatango, a platform for creating and managing chat rooms. It provides functionalities to work with Chatango APIs, handle WebSocket connections, and perform various operations related to chat rooms and user profiles.

## Table of Contents
- [Introduction](#introduction)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Introduction
Welcome to Chadango!

Chadango is a powerful and flexible chatbot framework inspired by the combination of Chatango and Dango. Just like the delightful combination of flavors in Dango, Chadango aims to bring together the best of both worlds, offering a robust chatbot solution with a touch of whimsy.

## Features
These are some of the features, but not limited to:
1. Each event is handled by its own goroutine.
2. Lightweight.
3. Some features are wrapped into solicited style (e.g., getting user list, online statuses, etc.).
4. Extensible.

## Installation
To run the example file, please follow the steps below:

### Prerequisites
Go 1.20 or higher should be installed on your system. You can download it from the official Go website: [https://golang.org/dl/](https://golang.org/dl/)

### Download the Example File
Download the example file minimal.go from the examples folder of this repository: [minimal.go](https://raw.githubusercontent.com/n0h4rt/chadango/master/examples/minimal/minimal.go)

### Install Required Packages
Navigate to the directory where you downloaded the example file and run the following command to install the required packages:

```shell
go mod init minimal
go get -u github.com/n0h4rt/chadango
go mod tidy
```

### Run the Example
Once the required packages are installed, execute the example file using the go run command:

```shell
go run minimal.go
```

This will execute the example code and display the output in the console.
Make sure you are in the correct directory where the `minimal.go` file is located.
Feel free to modify the example file according to your needs and explore the functionality provided by the packages you installed.
That's it! You have successfully installed the necessary packages and executed the example file.

## Usage
This is a simple usage.

```go
package main

import (
	"context"
	"fmt"

	dango "github.com/n0h4rt/chadango"
)

func main() {
	config := &dango.Config{
		Username: "username", // Change this
		Password: "password", // Change this
		Prefix:   ".",
		EnableBG: true,
		EnablePM: true,
		Groups:   []string{"groupchat1", "groupchat2"}, // Change this
	}

	app := dango.New(config)

	echoHandler := dango.NewCommandHandler(OnEcho, nil, "echo", "say")

	app.AddHandler(echoHandler)

	app.Initialize()

	ctx := context.Background()
	app.Start(ctx)

	// The `app.Park()` call is blocking, use CTRL + C to stop the application.
	// Use this if it is the top layer application.
	app.Park()
}

// OnEcho prints the command to the console and sends the argument as a reply.
func OnEcho(event *dango.Event, context *dango.Context) {
	var msg *dango.Message
	var err error

	if event.WithArgument {
		msg, err = event.Message.Reply(event.Argument)
	} else {
		msg, err = event.Message.Reply(`How to use ".echo Hello World!"`)
	}

	if err != nil {
		fmt.Printf("OnEcho: an error occurred when sending a reply (%s)\n", err)
	}
	if msg != nil {
		fmt.Printf("OnEcho: replied with: %s\n", msg.Text)
	}
}
```

## Dependencies
The Chadango library has the following dependencies:
- [github.com/rs/zerolog](https://github.com/rs/zerolog): Zero Allocation JSON Logger.
- [github.com/stretchr/testify](https://github.com/stretchr/testify): A toolkit with common assertions and mocks that plays nicely with the standard library.

## Contributing
Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the Chadango GitHub repository.

## License
Chadango is licensed under the [MIT License](https://opensource.org/license/mit/).

## Contact
You can find me in [khususme](https://khususme.chatango.com), available from 21:00 to 23:00 UTC-7. Feel free to reach out if you have any questions, suggestions, or just want to chat about Chadango!
