package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"
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

func chooseOne() (serial string, err error) {
	devices, err := listDevices()
	if err != nil {
		return
	}
	if len(devices) == 0 {
		err = errors.New("no devices/emulators found")
		return
	}
	d := choose(devices)
	return d.Serial, nil
}

func adbWrap(args ...string) {
	serial, err := chooseOne()
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command(adbPath(), args...)
	cmd.Env = append(os.Environ(), "ANDROID_SERIAL="+serial)
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
	exeName := "adb"
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}
	path, err := exec.LookPath(exeName)
	if err != nil {
		panic(err)
	}
	return path
}

func main() {
	app := cli.NewApp()
	app.Version = version
	app.Usage = "fa (fast adb) helps you win at adb"
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
				fmt.Printf("fa  version %s\n", version)
				adbVersion, err := DefaultAdbClient.Version()
				if err != nil {
					fmt.Printf("adb version err: %v\n", err)
					return err
				}
				fmt.Println("adb version", adbVersion)
				fmt.Println("adb path", adbPath())
				return nil
				// output, err := exec.Command(adbPath(), "version").Output()
				// for _, line := range strings.Split(string(output), "\n") {
				// 	fmt.Println("  " + line)
				// }
				// return err
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
				cli.BoolFlag{
					Name:  "open",
					Usage: "open file after screenshot",
				},
			},
			Action: actScreenshot,
		},
		{
			Name:      "install",
			Usage:     "install apk",
			UsageText: "fa install [ul] <apk-file | url>",
			// UseShortOptionHandling: true,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "uninstall if already installed",
				},
				cli.BoolFlag{
					Name:  "launch, l",
					Usage: "launch after success installed",
				},
			},
			Action: actInstall,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
