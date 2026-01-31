![gophish logo](https://raw.github.com/gophish/gophish/master/static/images/gophish_purple.png)

Gophish
=======

![Build Status](https://github.com/gophish/gophish/workflows/CI/badge.svg) [![GoDoc](https://godoc.org/github.com/gophish/gophish?status.svg)](https://godoc.org/github.com/gophish/gophish)

Gophish: Open-Source Phishing Toolkit

[Gophish](https://getgophish.com) is an open-source phishing toolkit designed for businesses and penetration testers. It provides the ability to quickly and easily setup and execute phishing engagements and security awareness training.

### Install

Installation of Gophish is dead-simple - just download and extract the zip containing the [release for your system](https://github.com/gophish/gophish/releases/), and run the binary. Gophish has binary releases for Windows, Mac, and Linux platforms.

### Building From Source (Local)

#### Prerequisites
- Go (use the version specified in `go.mod`)
- SQLite (default local DB uses SQLite)
- Node.js + npm (only required if you modify frontend assets under `static/js`)

#### Build
```bash
git clone https://github.com/gophish/gophish.git
cd gophish
go mod download
go build ./...
```

This produces the `gophish` binary in the repo root.

#### Run (Local)
```bash
./gophish
```

By default, Gophish reads `config.json`. Update it as needed for local ports, TLS certificates, and database settings.

After starting, open https://localhost:3333 and log in with the credentials printed in the server log output.

#### Frontend Asset Build (Optional)
If you change files in `static/js/src`, rebuild the bundled assets:

```bash
npm install
npm run build
```

### Testing

Run the full Go test suite:

```bash
go test ./...
```

Optional lint/static checks:

```bash
go vet ./...
```

### Docker
You can also use Gophish via the official Docker container [here](https://hub.docker.com/r/gophish/gophish/).

### Setup
After running the Gophish binary, open an Internet browser to https://localhost:3333 and login with the default username and password listed in the log output.
e.g.
```
time="2020-07-29T01:24:08Z" level=info msg="Please login with the username admin and the password 4304d5255378177d"
```

Releases of Gophish prior to v0.10.1 have a default username of `admin` and password of `gophish`.

### Documentation

Documentation can be found on our [site](http://getgophish.com/documentation). Find something missing? Let us know by filing an issue!

### Campaign CC

When creating a new campaign, you can optionally add a comma-separated list of CC recipients. This list is applied to every email sent for that campaign.

Steps:
1. Open **New Campaign**.
2. Fill out campaign details as usual.
3. In **CC (Optional)**, enter addresses like `cc1@example.com, cc2@example.com` (semicolons are also accepted).
4. Launch the campaign.

### Issues

Find a bug? Want more features? Find something missing in the documentation? Let us know! Please don't hesitate to [file an issue](https://github.com/gophish/gophish/issues/new) and we'll get right on it.

### License
```
Gophish - Open-Source Phishing Framework

The MIT License (MIT)

Copyright (c) 2013 - 2020 Jordan Wright

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software ("Gophish Community Edition") and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```
