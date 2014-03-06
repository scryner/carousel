package main

import (
	"fmt"
	"os/exec"
)

// osascript -e 'tell app "Safari" to make new document at end of documents with properties {URL:"http://naver.com"}'

func launchWebBrowser() {
	osascriptCommand := fmt.Sprintf("tell app \"safari\" to make new document at end of documents with properties {URL:\"http://localhost:%d\"}", port)
	cmd := exec.Command("osascript", "-e", osascriptCommand)

	go func() {
		err := cmd.Start()
		if err != nil {
			logger.Errorf("can't launch Safari: %v", err)
			return
		}

		err = cmd.Wait()
		if err != nil {
			logger.Errorf("can't wait Safari: %v", err)
			return
		}

		logger.Infof("launch Safari to http://localhost:%d", port)
	}()
}
