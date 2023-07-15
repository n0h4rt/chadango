package main

import (
	"context"
	"fmt"

	dango "github.com/n0h4rt/chadango"
)

func main() {
	// Example config
	config := &dango.Config{
		Username: "username",                           // The username of your bot account.
		Password: "password",                           // The password of your bot account.
		Prefix:   ".",                                  // The prefix for command handlers.
		EnableBG: true,                                 // Turns ON the background (if avaliable).
		EnablePM: true,                                 // Turns ON the private chat.
		Groups:   []string{"groupchat1", "groupchat2"}, // Initial group chats slice to allow the bot to run.
	}

	// Loads config from config.json
	/* config, err := dango.LoadConfig("config.json")
	if err != nil {
		panic(err)
	} */

	// Create a new bot.
	app := dango.New(config)

	// Create a new command handler.
	echoHandler := dango.NewCommandHandler(OnEcho, filter, "echo", "say")

	// You can chain the `AddHandler` function,
	// like `app.AddHandler(Handler).AddHandler(Handler).AddHandler(Handler)`.
	app.AddHandler(echoHandler)

	app.Initialize()

	// You can use your own parent context.
	ctx := context.Background()
	// Start the bot.
	app.Start(ctx)

	// The `app.Park()` call is blocking, use CTRL + C to stop the bot.
	// Use this if it is the top layer application.
	app.Park()

	// Now that the bot has started, try sending a message ".echo Hello World!".
	// It should respond with "Hello World!".

	// This can be called from another goroutine.
	// This needs to be called when `app.Park()` is not called.
	// app.Stop()
}

var (
	// userFilter is a `Filter` based on the user name
	userFilter = dango.NewUserFilter("perorist")

	// chatFilter is a `Filter` based on the group name
	chatFilter = dango.NewChatFilter("khususme", "square-enix")

	// lockFilter is a `Filter` based on the group name
	lockFilter = dango.NewChatFilter("square-enix")

	// What this filter does:
	// - The user name should be in the `userFilter` (a whitelist) OR
	// - The group name should be in the `chatFilter` (a whitelist) AND
	// - The group name should NOT be in the `lockFilter` (a blacklist)
	filter = userFilter.Or(chatFilter).And(lockFilter.Not())
)

// OnEcho prints the command to the console and sends the argument as a reply.
func OnEcho(event *dango.Event, context *dango.Context) {
	fmt.Printf("A command received: %s %s\n", event.Command, event.Argument)

	// You can check whether the command is from a group chat or a private chat.
	if event.IsPrivate {
		fmt.Println("and it's private")
	}

	// Create a non-local variable to make it accessible outside of the if-else statement.
	var msg *dango.Message
	var err error

	// The `event.Message.Reply` is a shortcut for `event.Group.SendMessage` or `event.Private.SendMessage`.
	// It returns two values:
	// - The sent message and an error if it's from a group chat, or
	// - Nil and an error if it's from a private chat.
	if event.WithArgument {
		msg, err = event.Message.Reply(event.Argument)
	} else {
		msg, err = event.Message.Reply(`How to use ".echo Hello World!"`)
	}

	// Checks if there is any error (optional).
	if err != nil {
		fmt.Printf("OnEcho: an error occurred when sending a reply (%s)\n", err)
	}

	// You can use the sent message, such as printing it to the console.
	// Note: The message returned from `event.Private.SendMessage` is always nil by design.
	if msg != nil {
		fmt.Printf("OnEcho: replied with: %s\n", msg.Text)
	}
}
