# fa = fast adb
[![Build Status](https://travis-ci.org/codeskyblue/fa.svg?branch=master)](https://travis-ci.org/codeskyblue/fa)

`fa` is a command line tool that wraps `adb` in order to extend it with extra features and commands that make working with Android easier.

## Features
- [x] show device selection when multi device connected
- [x] screenshot
- [ ] install support http url
- [ ] support launch after install apk
- [ ] show wlan (ip,mac,signal), enable and disable it

## Install
**For mac**

```bash
brew install codeskyblue/tap/fa
```

**For windows and linux**

download binary from [**releases**](https://github.com/codeskyblue/fa/releases)

## Usage
Screenshot

```bash
fa screenshot -o screenshot.png
```

~~Install APK~~

```
fa install https://example.org/demo.apk
```

## Reference
- <https://github.com/mzlogin/awesome-adb>

## LICENSE
[MIT](LICENSE)
