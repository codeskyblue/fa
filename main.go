package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/manifoldco/promptui"
	"gopkg.in/urfave/cli.v1"
)

var (
	version = "develop"
)

type Device struct {
	Serial      string
	Description string
}

func (d *Device) String() string {
	return d.Serial
}

func shortDeviceInfo(s string) string {
	fields := strings.Fields(s)
	props := make(map[string]string, 4)
	for _, v := range fields {
		kv := strings.SplitN(v, ":", 2)
		if len(kv) != 2 {
			continue
		}
		props[kv[0]] = kv[1]
	}
	if props["model"] != "" {
		return props["model"]
	}
	return s
}

func listDevices() (ds []Device, err error) {
	output, err := exec.Command("adb", "devices", "-l").CombinedOutput()
	if err != nil {
		return
	}
	re := regexp.MustCompile(`(?m)^([^\s]+)\s+device\s+(.+)$`)
	matches := re.FindAllStringSubmatch(string(output), -1)
	for _, m := range matches {
		desc := shortDeviceInfo(m[2])
		ds = append(ds, Device{
			Serial:      m[1],
			Description: desc,
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
		Active:   "➤ {{ .Serial | cyan }}  {{ .Description | faint }}",
		Inactive: "  {{ .Serial | faint }}  {{ .Description | faint }}",
	}
	prompt := promptui.Select{
		Label:     "Select device",
		Items:     devices,
		Templates: templates,
	}

	i, _, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}
	return devices[i]
}

func adbWrap(args ...string) {
	devices, err := listDevices()
	if err != nil {
		log.Fatal(err)
	}
	if len(devices) == 0 {
		log.Fatal("no devices detected")
	}

	d := choose(devices)
	cmd := exec.Command(adbPath(), args...)
	cmd.Env = append(os.Environ(), "ANDROID_SERIAL="+d.Serial)
	// cmd.Args = os.Args
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

func adbPath() string {
	return "adb"
}

func main() {
	app := cli.NewApp()
	app.Name = "ya"
	app.Version = version
	app.Usage = "ya: your adb helps you win at adb"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "codeskyblue",
			Email: "codeskyblue@gmail.com",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "version",
			Usage: "show version",
			Action: func(ctx *cli.Context) error {
				fmt.Printf("[ya]\n  version %s\n", version)
				fmt.Println("[adb]")
				c := exec.Command(adbPath(), "version")
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				c.Run()
				return nil
			},
		},
		{
			Name:            "adb",
			Usage:           "exec adb with device select",
			SkipFlagParsing: true,
			Action: func(ctx *cli.Context) error {
				adbWrap(ctx.Args()...)
				return nil
			},
		},
		{
			Name:  "screenshot",
			Usage: "take screenshot",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output, o",
					Value: "screenshot.png",
					Usage: "output screenshot name",
				},
			},
			Action: func(ctx *cli.Context) error {
				log.Println(ctx.String("output"))
				c := exec.Command(adbPath(), "exec-out", "screencap", "-p")
				imgfile, err := os.Create(ctx.String("output"))
				if err != nil {
					return err
				}
				defer imgfile.Close()
				c.Stdout = imgfile
				c.Stderr = os.Stderr
				return c.Run()
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
