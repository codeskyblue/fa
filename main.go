package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	shellquote "github.com/kballard/go-shellquote"
	tty "github.com/mattn/go-tty"

	"github.com/manifoldco/promptui"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	version       = "develop"
	debug         = false
	defaultSerial string
	defaultHost   string
	defaultPort   int
)

type Device struct {
	Serial      string `json:"serial"`
	Status      string `json:"status"`
	Description string `json:"-"`
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
	output, err := exec.Command("adb", "devices").CombinedOutput()
	if err != nil {
		return
	}
	re := regexp.MustCompile(`(?m)^([^\s]+)\s+(device|offline|unauthorized)\s*$`)
	matches := re.FindAllStringSubmatch(string(output), -1)
	ds = make([]Device, 0, len(matches))
	for _, m := range matches {
		status := m[2]
		ds = append(ds, Device{
			Serial: m[1],
			Status: status,
		})
	}
	return
}

func listDetailedDevices() (ds []Device, err error) {
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
	if defaultSerial != "" {
		return Device{Serial: defaultSerial}
	}
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
	devices, err := listDetailedDevices()
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
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "show debug info",
			Destination: &debug,
		},
		cli.StringFlag{
			Name:        "serial, s",
			Usage:       "use device with given serial",
			EnvVar:      "ANDROID_SERIAL",
			Destination: &defaultSerial,
		},
		cli.StringFlag{
			Name:        "host, H",
			Usage:       "name of adb server host",
			Value:       "localhost",
			Destination: &defaultHost,
		},
		cli.IntFlag{
			Name:        "port, P",
			Usage:       "port of adb server",
			Value:       5037,
			Destination: &defaultPort,
		},
	}
	app.Commands = []cli.Command{
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
			Name:  "version",
			Usage: "show version",
			Action: func(ctx *cli.Context) error {
				fmt.Printf("fa version %s\n", version)
				adbVersion, err := NewAdbClient().Version()
				if err != nil {
					fmt.Printf("adb version err: %v\n", err)
					return err
				}
				fmt.Println("adb path", adbPath())
				fmt.Println("adb server version", adbVersion)
				return nil
				// output, err := exec.Command(adbPath(), "version").Output()
				// for _, line := range strings.Split(string(output), "\n") {
				// 	fmt.Println("  " + line)
				// }
				// return err
			},
		},
		{
			Name:  "devices",
			Usage: "show connected devices",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "json",
					Usage: "output json format",
				},
			},
			Action: func(ctx *cli.Context) error {
				ds, err := listDevices()
				if err != nil {
					return err
				}
				if ctx.Bool("json") {
					data, _ := json.MarshalIndent(ds, "", "  ")
					fmt.Println(string(data))
				} else {
					for _, d := range ds {
						fmt.Printf("%s\t%s\n", d.Serial, d.Status)
					}
				}
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
			Name:            "shell",
			Usage:           "run shell command",
			SkipFlagParsing: true,
			Action: func(ctx *cli.Context) error {
				serial, err := chooseOne()
				if err != nil {
					return err
				}
				device := DefaultAdbClient.DeviceWithSerial(serial)

				var cmd string
				if len(ctx.Args()) != 0 {
					cmd = `PATH="$PATH:/data/local/tmp" ` + shellquote.Join(ctx.Args()...)
				}
				rwc, err := device.OpenShell(cmd)
				if err != nil {
					return err
				}
				defer rwc.Close()
				tty, err := tty.Open()
				if err != nil {
					log.Fatal(err)
				}
				defer tty.Close()
				go io.Copy(rwc, tty.Input())
				_, err = io.Copy(tty.Output(), rwc)
				return err
			},
		},
		{
			Name:      "install",
			Usage:     "install apk",
			UsageText: "fa install [ul] <apk-file | url>",
			// UseShortOptionHandling: true, // not supported in current urfav/cli
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
		{
			Name:      "pidcat",
			Usage:     "logcat filter with package name",
			UsageText: "fa pidcat [package-name ...]",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "current",
					Usage: "filter logcat by current running app",
				},
				cli.BoolFlag{
					Name:  "clear",
					Usage: "clear the entire log before running",
				},
				cli.StringFlag{
					Name:  "min-level, l",
					Usage: "Minimum level to be displayed {V,D,I,W,E,F}",
				},
				cli.StringSliceFlag{
					Name:  "tag, t",
					Usage: "filter output by specified tag(s)",
				},
				cli.StringSliceFlag{
					Name:  "ignore-tag, i",
					Usage: "filter output by ignoring specified tag(s)",
				},
			},
			Action: func(ctx *cli.Context) error {
				serial, err := chooseOne()
				if err != nil {
					return err
				}
				faDir := filepath.Join(os.Getenv("HOME"), ".fa")
				if err := os.MkdirAll(faDir, 0755); err != nil {
					return err
				}
				pidcatPath := filepath.Join(faDir, "pidcat.py")
				err = ioutil.WriteFile(pidcatPath, []byte(pidcatCode), 0644)
				if err != nil {
					return err
				}
				args := []string{pidcatPath, "-s", serial}
				if ctx.Bool("current") {
					args = append(args, "--current")
				}
				if ctx.Bool("clear") {
					args = append(args, "--clear")
				}
				if ctx.String("min-level") != "" {
					args = append(args, "-l", ctx.String("min-level"))
				}
				for _, tag := range ctx.StringSlice("tag") {
					args = append(args, "-t", tag)
				}
				for _, ignore := range ctx.StringSlice("ignore-tag") {
					args = append(args, "-i", ignore)
				}
				args = append(args, ctx.Args()...)
				return runCommand("python", args...)
			},
		},
		{
			Name:  "get-serialno",
			Usage: "print serial-number",
			Action: func(ctx *cli.Context) error {
				serial, err := chooseOne()
				if err != nil {
					return err
				}
				client := NewAdbClient()
				device := client.DeviceWithSerial(serial)
				realSerial, err := device.SerialNo()
				if err != nil {
					return err
				}
				println(realSerial)
				return nil
			},
		},
		{
			Name:  "healthcheck",
			Usage: "check device health status",
			Action: func(ctx *cli.Context) error {
				log.Println("check install")
				err := runCommand(os.Args[0], "install", "-f", "https://github.com/appium/java-client/raw/master/src/test/java/io/appium/java_client/ApiDemos-debug.apk")
				if err != nil {
					return err
				}
				log.Println("OKAY")
				return nil
			},
		},
		{
			Name:  "watch",
			Usage: "show newest state when device state change",
			Action: func(ctx *cli.Context) error {
				client := NewAdbClient()
				eventC, err := client.Watch()
				if err != nil {
					return err
				}
				for ev := range eventC {
					println(ev)
				}
				return nil
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
