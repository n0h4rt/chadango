// chadango package is a modern framework designed to facilitate the creation and management of bots for the Chatango messaging platform. Written in Go, this package leverages Go's concurrency capabilities to provide efficient and robust bot interactions.
//
// Key Features:
//   - Chatango Client: A client for connecting to and interacting with the Chatango messaging platform.
//   - Connection Management: Functions to establish and terminate connections securely.
//   - Concurrency Support: Utilizes Go's synchronization mechanisms to ensure thread-safe operations.
//   - Customizable API: Provides a flexible API structure for handling various Chatango functionalities.
//   - Error Handling: Implements mechanisms for managing errors and maintaining stable connections.
//
// Usage Example:
//
//	package main
//
//	import (
//	    "context"
//
//	    dango "github.com/n0h4rt/chadango"
//	)
//
//	func main() {
//	    config := &dango.Config{
//	        Username: "username",
//	        Password: "password",
//	        Prefix:   ".",
//	        EnableBG: true,
//	        EnablePM: true,
//	        Groups:   []string{"groupchat1", "groupchat2"},
//	    }
//
//	    app := dango.New(config)
//
//	    // add handlers
//
//	    app.Initialize()
//
//	    ctx := context.Background()
//	    app.Start(ctx)
//
//	    // The `app.Park()` call is blocking, use CTRL + C to stop the application.
//	    // Use this if it is the top layer application.
//	    app.Park()
//	}
//
// The chadango package aims to streamline the development of Chatango bots, providing a solid foundation for building advanced and scalable messaging solutions.
package chadango

func init() {
	// Initializes the [http.Client] with custom headers and a cookie jar.
	// Every Chatango API calls should made with this client for efficiency.
	initHttpClient()
}
