package main

import (
	"fmt"
	"os/exec"
)

func launchWebBrowser() {
	cmd := exec.Command("x-www-browser", fmt.Sprintf("http://localhost:%d", port))

	go func() {
		err := cmd.Start()
		if err != nil {
			logger.Errorf("can't launch default web browser: %v", err)
			return
		}

		err = cmd.Wait()
		if err != nil {
			logger.Errorf("can't wait default web browser: %v", err)
			return
		}

		logger.Infof("launch to http://localhost:%d", port)
	}()
}
