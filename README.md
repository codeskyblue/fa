# fa = fast adb
[![Build Status](https://travis-ci.org/codeskyblue/fa.svg?branch=master)](https://travis-ci.org/codeskyblue/fa)

`fa` is a command line tool that wraps `adb` in order to extend it with extra features and commands that make working with Android easier.

## Features
- [x] show device selection when multi device connected
- [x] screenshot
- [x] install support http url
- [x] support launch after install apk
- [ ] show wlan (ip,mac,signal), enable and disable it

## Install
**For mac**

```bash
brew install codeskyblue/tap/fa
```

**For windows and linux**

download binary from [**releases**](https://github.com/codeskyblue/fa/releases)

## Usage
Show version

```bash
$ fa version
fa version v0.0.5 # just example
```

Screenshot (only png support for now)

```bash
fa screenshot -o screenshot.png
```

Install APK

```
$ fa install ApiDemos-debug.apk
```

Install APK then start app

```
$ fa install --launch ApiDemos-debug.apk
```

Install APK from URL with _uninstall first and launch after installed_


```
$ fa install --force --launch https://github.com/appium/java-client/raw/master/src/test/java/io/appium/java_client/ApiDemos-debug.apk
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

Run adb command, if multi device connected, `fa` will give you choice to select one.

```
$ fa adb pwd
@ select device
  > 3aff8912  Smartion
    vv12afvv  Google Nexus 5
{selected 3aff8912}
/
```

## Reference
- <https://github.com/mzlogin/awesome-adb>

## LICENSE
[MIT](LICENSE)
