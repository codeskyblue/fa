package main

import (
	"log"
	"os"
	"os/exec"
	"regexp"
	"syscall"

	"github.com/manifoldco/promptui"
)

type Device struct {
	Serial      string
	Description string
}

func listDevices() (ds []Device, err error) {
	output, err := exec.Command("adb", "devices", "-l").CombinedOutput()
	if err != nil {
		return
	}
	re := regexp.MustCompile(`(?m)^([^\s]+)\s+device\s+(.+)$`)
	matches := re.FindAllStringSubmatch(string(output), -1)
	for _, m := range matches {
		ds = append(ds, Device{
			Serial:      m[1],
			Description: m[2],
		})
	}
	return
}

func choose(devices []Device) Device {
	if len(devices) == 1 {
		return devices[0]
	}
	templates := &promptui.SelectTemplates{
		Label:    "✨ {{ . | green}}", //"{{ . }}?",
		Active:   "➤ {{ .Serial | cyan }} {{ .Description | faint }}",
		Inactive: "  {{ .Serial | faint }} {{ .Description | faint }}",
	}
	prompt := promptui.Select{
		Label:     "Select Device",
		Items:     devices,
		Templates: templates,
	}

	i, _, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}
	return devices[i]
}

func main() {
	devices, err := listDevices()
	if err != nil {
		log.Fatal(err)
	}
	if len(devices) == 0 {
		log.Fatal("no devices detected")
	}

	d := choose(devices)
	cmd := exec.Command("adb")
	cmd.Env = append(os.Environ(), "ANDROID_SERIAL="+d.Serial)
	cmd.Args = os.Args
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		log.Fatal(err)
	}
}
