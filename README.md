# ya = your adb
[![Build Status](https://travis-ci.org/codeskyblue/ya.svg?branch=master)](https://travis-ci.org/codeskyblue/ya)

`ya` is a command line tool that wraps `adb` in order to extend it with extra features and commands that make working with Android easier.

## Features
- [x] show device selection when multi device connected
- [x] screenshot
- [ ] install support http url
- [ ] support launch after install apk
- [ ] show wlan (ip,mac,signal), enable and disable it

## Install
**For mac**

```bash
brew install codeskyblue/tap/ya
```

**For windows and linux**

download binary from [**releases**](https://github.com/codeskyblue/ya/releases)

## Usage
Screenshot

```bash
ya screenshot -o screenshot.png
```

~~Install APK~~

```
ya install https://example.org/demo.apk
```

## Reference
- <https://github.com/mzlogin/awesome-adb>

## LICENSE
[MIT](LICENSE)
