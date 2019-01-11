# fa = fast adb
[![Build Status](https://travis-ci.org/codeskyblue/fa.svg?branch=master)](https://travis-ci.org/codeskyblue/fa)
[![GoDoc](https://godoc.org/github.com/codeskyblue/fa/adb?status.svg)](https://godoc.org/github.com/codeskyblue/fa/adb)

`fa` is a command line tool that wraps `adb` in order to extend it with extra features and commands that make working with Android easier.

## Features
- [x] show device selection when multi device connected
- [x] screenshot
- [x] install support http url
- [x] support launch after install apk
- [x] support `fa devices --json`
- [x] support `fa shell`
- [x] colorful logcat and filter with package name
- [ ] install apk and auto click confirm
- [.] check device health status
- [ ] show current app
- [ ] unlock device
- [ ] reset device state, clean up installed packages
- [ ] show wlan (ip,mac,signal), enable and disable it
- [ ] share device to public web
- [ ] install ipa support

## Install
**For mac**

```bash
brew install codeskyblue/tap/fa
```

**For windows and linux**

download binary from [**releases**](https://github.com/codeskyblue/fa/releases)

## Usage
### Show version

```bash
$ fa version
fa version v0.0.5 # just example
adb server version 28
```

### Show devices
- [x] Remove header `List of devices attached` to make it easy parse

```bash
$ fa devices
3578298f        device

$ fa devices --json
[
   {"serial": "3578298f", "status": "device"}
]
```

### Run adb command with device select
if multi device connected, `fa` will give you list of devices to choose.

```bash
$ fa adb shell
@ select device
  > 3aff8912  Smartion
    vv12afvv  Google Nexus 5
{selected 3aff8912}
shell $
```

`-s` option and `$ANDROID_SERIAL` is also supported, but if you known serial, maybe use `adb` directly is better.

```bash
$ fa -s 3578298f adb shell pwd
/
$ ANDROID_SERIAL=3578298 fa adb shell pwd
/
```

### Screenshot
only `png` format now.

```bash
fa screenshot -o screenshot.png
```

### Install APK

```bash
fa install ApiDemos-debug.apk # from local file
fa install http://example.org/demo.apk # from URL
fa install -l ApiDemos-debug.apk # launch after install
fa install -f ApiDemos-debug.apk # uninstall before install
```

Show debug info when install

```bash
$ fa -d install --force --launch https://github.com/appium/java-client/raw/master/src/test/java/io/appium/java_client/ApiDemos-debug.apk
Downloading ApiDemos-debug.apk...
 2.94 MiB / 2.94 MiB [================================] 100.00% 282.47 KiB/s 10s
Download saved to ApiDemos-debug.apk
+ adb -s 0123456789ABCDEF uninstall io.appium.android.apis
+ adb -s 0123456789ABCDEF install ApiDemos-debug.apk
ApiDemos-debug.apk: 1 file pushed. 4.8 MB/s (3084877 bytes in 0.609s)
        pkg: /data/local/tmp/ApiDemos-debug.apk
Success
Launch io.appium.android.apis ...
+ adb -s 0123456789ABCDEF shell am start -n io.appium.android.apis/.ApiDemos
```

### App
```
$ fa app list # show all app package names
$ fa app list -3 # only show third party packages
```

### Shell
Like `adb shell`, run `fa shell` will open a terminal

The difference is `fa shell` will add `/data/local/tmp` into `$PATH`
So if you have binary `busybox` in `/data/local/tmp`,
You can just run

```
$ fa shell busybox ls
# using adb shell you have to
$ adb shell /data/local/tmp/busybox ls
```

### Pidcat (logcat)
Current implementation is wrapper of [pidcat.py](https://github.com/JakeWharton/pidcat)
So use this feature, you need python installed.

```bash
$ fa help pidcat
USAGE:
   fa pidcat [package-name ...]

OPTIONS:
   --current                     filter logcat by current running app
   --clear                       clear the entire log before running
   --min-level value, -l value   Minimum level to be displayed {V,D,I,W,E,F}
   --tag value, -t value         filter output by specified tag(s)
   --ignore-tag value, -i value  filter output by ignoring specified tag(s)
```

The pidcat is very beautiful.

![pidcat](https://github.com/JakeWharton/pidcat/raw/master/screen.png)

## Thanks for these Articles and Codes
- <https://labstack.com/docs/tunnel>
- <https://github.com/mzlogin/awesome-adb>
- [Facebook One World Project](https://code.fb.com/android/managing-resources-for-large-scale-testing/)
- [Facebook Device Lab](https://code.fb.com/android/the-mobile-device-lab-at-the-prineville-data-center/)
- Article reverse ssh tunnling <https://www.howtoforge.com/reverse-ssh-tunneling>
- [openstf/adbkit](https://github.com/openstf/adbkit)
- [ADB Source Code](https://github.com/aosp-mirror/platform_system_core/blob/master/adb)
- ADB Protocols [OVERVIEW.TXT](https://github.com/aosp-mirror/platform_system_core/blob/master/adb/OVERVIEW.TXT) [SERVICES.TXT](https://github.com/aosp-mirror/platform_system_core/blob/master/adb/SERVICES.TXT) [SYNC.TXT](https://github.com/aosp-mirror/platform_system_core/blob/master/adb/SYNC.TXT)
- [JakeWharton/pidcat](https://github.com/JakeWharton/pidcat)
- <https://github.com/wmbest2/android>
- <https://github.com/zach-klippenstein/goadb>


Libs might be useful

- <https://github.com/fatedier/frp>
- <https://golanglibs.com/search?q=tunnel>
- <https://github.com/koding/tunnel>
- <https://github.com/mmatczuk/go-http-tunnel>
- <https://github.com/inconshreveable/go-tunnel>
- <https://github.com/labstack/tunnel-client> SSH Tunnel
- <https://github.com/gliderlabs/ssh> Easy SSH servers in Golang

## LICENSE
[MIT](LICENSE)
