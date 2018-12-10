package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/pkg/errors"
	"github.com/shogo82148/androidbinary/apk"
	pb "gopkg.in/cheggaaa/pb.v1"
	cli "gopkg.in/urfave/cli.v1"
)

func httpDownload(dst string, url string) (resp *grab.Response, err error) {
	client := grab.NewClient()
	req, err := grab.NewRequest(dst, url)
	if err != nil {
		return nil, err
	}
	// start download
	resp = client.Do(req)
	fmt.Printf("Downloading %v...\n", resp.Filename)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

	bar := pb.New(int(resp.Size))
	bar.SetMaxWidth(80)
	bar.ShowSpeed = true
	bar.ShowTimeLeft = false
	bar.SetUnits(pb.U_BYTES)
	bar.Start()

Loop:
	for {
		select {
		case <-t.C:
			bar.Set(int(resp.BytesComplete()))
		case <-resp.Done:
			bar.Set(int(resp.Size))
			bar.Finish()
			break Loop
		}
	}
	// check for errors
	if err := resp.Err(); err != nil {
		return nil, errors.Wrap(err, "download failed")
	}
	fmt.Println("Download saved to", resp.Filename)
	return resp, err
}

func actInstall(ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return errors.New("apkfile or apkurl should provided")
	}
	serial, err := chooseOne()
	if err != nil {
		return err
	}
	arg := ctx.Args().First()

	// download apk
	apkpath := arg
	if regexp.MustCompile(`^https?://`).MatchString(arg) {
		resp, err := httpDownload(".", arg)
		if err != nil {
			return err
		}
		apkpath = resp.Filename
	}

	// parse apk
	pkg, err := apk.OpenFile(apkpath)
	if err != nil {
		return err
	}

	// handle --force
	if ctx.Bool("force") {
		pkgName := pkg.PackageName()
		adbCommand(serial, "uninstall", pkgName).Run()
	}

	// install
	outBuffer := bytes.NewBuffer(nil)
	c := adbCommand(serial, "install", "-r", apkpath)
	c.Stdout = io.MultiWriter(os.Stdout, outBuffer)
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return err
	}

	if strings.Contains(outBuffer.String(), "Failure") {
		return errors.New("install failed")
	}
	if ctx.Bool("launch") {
		packageName := pkg.PackageName()
		mainActivity, er := pkg.MainActivity()
		if er != nil {
			fmt.Println("apk have no main-activity")
			return nil
		}
		if !strings.Contains(mainActivity, ".") {
			mainActivity = "." + mainActivity
		}
		fmt.Println("Launch app", packageName, "...")
		adbCommand(serial, "shell", "am", "start", "-n", packageName+"/"+mainActivity).Run()
	}
	return nil
}
