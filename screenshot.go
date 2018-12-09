package main

import (
	"log"
	"os"
	"os/exec"

	cli "gopkg.in/urfave/cli.v1"
)

func adbCommand(serial string, args ...string) *exec.Cmd {
	c := exec.Command(adbPath(), args...)
	c.Env = append(os.Environ(), "ANDROID_SERIAL="+serial)
	return c
}

func screenshotExecOut(serial, output string) error {
	serial, err := chooseOne()
	if err != nil {
		return err
	}
	c := adbCommand(serial, "exec-out", "screencap", "-p")
	imgfile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func() {
		imgfile.Close()
		if err != nil {
			os.Remove(output)
		}
	}()
	c.Stdout = imgfile
	// c.Stderr = os.Stderr
	return c.Run()
}

func screenshotScreencap(serial, output string) error {
	tmpPath := "/sdcard/fa-screenshot.png"
	c := adbCommand(serial, "shell", "screencap", "-p", tmpPath)
	if err := c.Run(); err != nil {
		return err
	}
	defer adbCommand(serial, "shell", "rm", tmpPath).Run()
	return adbCommand(serial, "pull", tmpPath, output).Run()
}

func actScreenshot(ctx *cli.Context) (err error) {
	serial, err := chooseOne()
	if err != nil {
		return err
	}
	output := ctx.String("output")
	err = screenshotExecOut(serial, output)
	if err != nil {
		// log.Println("FAIL:", "exec-out", "screencap")
		err = screenshotScreencap(serial, output)
	}
	if err == nil {
		// log.Println("OKAY:", "shell", "screencap")
		log.Println("save to", output)
		if ctx.Bool("open") {
			// TODO(ssx): only works on mac
			exec.Command("open", output).Run()
		}
	}
	return
}
