package main

import "github.com/pivaldi/mmw/todo"

func main() {
	// if err := run(config, logger); err != nil {
	// 	logger.Error("application failed", "error", err)
	// 	os.Exit(1)
	// }

	app, err := todo.New(nil)

	if err != nil {
		panic(err)
	}

	err = app.Run()

	// TODO: gracefull restart here ?
	if err != nil {
		panic(err)
	}
}
