

<h1 align="center">Go-APOD</h1>

<p align="center">
  <i>A CORS-enabled, no-auth wrapper to NASA's Astronomy Picture of the Day </i><br>
  <b>Public API: <a href="https://apod.as93.net/">apod.as93.net</a></b><br><br>
  <img width="100" src="https://raw.githubusercontent.com/Lissy93/go-apod/master/static/assets/pwa/apple-touch-icon.png" />
</p>


<details>
<summary><b>Contents</b></summary>

- [API Usage](#api-usage)
  - [`/apod`](#apod)
  - [`/image`](#image)
- [Deployment](#deployment)
  - [Heroku](#heroku)
  - [Docker](#docker)
  - [Executable](#from-executable)
  - [From Source](#from-source)
- [Development](#building-locally)
  - [Project Commands](#commands)
  - [Configuration Options](#environmental-variables)
- [Frontend App](#app)
- [Contributing](#contributing)
- [License](#license)

</details>

---

## API Usage


### `/apod`

> Returns full JSON info about todays picture

**Example**

```
GET https://apod.as93.net/apod
```

**Response**

```json
{
  "copyright": "Robert Eder",
  "date": "2026-09-30",
  "explanation": "Peculiar spiral galaxy Arp 78 is found within the boundaries of the head strong constellation Aries. Some 100 million light-years beyond the stars and nebulae of our Milky Way galaxy, the island universe is an enormous 200,000 light-years across. Also known as NGC 772, it sports a prominent, outer spiral arm in this detailed cosmic portrait. Tracking along sweeping dust lanes and lined with young blue star clusters, Arp 78's overdeveloped spiral arm is pumped-up by galactic-scale gravitational tides. Interactions with its brightest companion galaxy, the more compact NGC 770 seen directly below the larger spiral, are likely responsible. Embedded in faint star streams revealed in the deep telescopic exposure, NGC 770's fuzzy, elliptical appearance contrasts nicely with spiky foreground Milky Way stars.",
  "hdurl": "https://assets.science.nasa.gov/dynamicimage/assets/science/cds/apod/apod/2026/october/NGC772_Robert_Eder.jpg?w=1772&h=1182&fit=clip&crop=faces%2Cfocalpoint",
  "media_type": "image",
  "title": "Arp 78: Peculiar Galaxy in Aries",
  "url": "https://assets.science.nasa.gov/dynamicimage/assets/science/cds/apod/apod/2026/october/NGC772_Robert_Eder.jpg"
}
```

---

### `/image`

> Returns todays image

**Example**

```html
<img
  src="https://apod.as93.net/image"
  alt="Astronomy Picture of the Day"
  width="350"
/>
```

**Response**

<img src="https://apod.as93.net/image" alt="Astronomy Picture of the Day" width="350" />

---

## Deployment

> _Go-APOD can be self-hosted, either with Docker, via the 1-click Vercel or Heroku deployment, or by running the executable directly._<br>
> No API key is needed, data comes from NASA's public APOD feed on [science.nasa.gov](https://science.nasa.gov/apod/).

### Vercel

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2FLissy93%2Fgo-apod&project-name=apod&repository-name=go-apod&demo-title=Go-APOD&demo-description=A%20demo%20is%20published%20to%20apod.as93.net&demo-url=https%3A%2F%2Fapod.as93.net%2F&demo-image=https%3A%2F%2Fraw.githubusercontent.com%2FLissy93%2Fgo-apod%2Fmaster%2Fstatic%2Fassets%2Fpwa%2Fapple-touch-icon.png)

### Heroku

[![Deploy to Heroku](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/Lissy93/go-apod)

### Docker
A multi-arch container is available on DockerHub, under [`lissy93/apod`](https://hub.docker.com/r/lissy93/apod), or GHCR  as [`ghcr.io/lissy93/go-apod`](https://github.com/Lissy93/go-apod/pkgs/container/go-apod).<br> Or, use this [`docker-compose.yml`](https://github.com/Lissy93/go-apod/blob/master/docker-compose.yml) template, and run `docker compose up`.

```bash
docker run -p 8080:8080 -d lissy93/apod
```

### From Executable

Each release has pre-compiled binaries attached for Windows, Mac and Linux, which can be run directly.
From the [Releases Page](https://github.com/Lissy93/go-apod/releases), download and extract the version for your system, then execute it with: `./go-apod`

### From Source

See the [Building Locally](#building-locally) section below

---


## Building Locally

> If you haven't already done so, you'll need to [install Go Lang](https://go.dev/doc/install).<br>
> Then clone the repo `git clone https://github.com/Lissy93/go-apod.git && cd go-apod`


### Commands
- Run Directly > `go run .`
- Compile App > `go build`
- Run Tests > `go test`

### Environmental Variables

- `PORT` (Optional) - The port to start the web server on, defaults to `8080`
- `CORS_ALLOWED_ORIGINS` (Optional) - Comma-separated list of origins which can use the API, defaults to `*` / all
- `NASA_BASE_URL` (Optional) - The upstream feed URL, defaults to NASA's `apod-basic` feed on science.nasa.gov
- `CACHE_TTL` (Optional) - How long to cache NASA's response for, defaults to `15m`

---

## App

> The service also includes an optional simple web app, which can be used to show todays image and associated information from the API.

<p align="center">
  <a href="https://apod.as93.net">
  <img src="https://i.ibb.co/rvCfrbn/go-apod-screenshot.png" width="600" />
  </a>
</p>

---

## Contributing

Contributions of any kind are very welcome, and would be much appreciated :)
For Code of Conduct, see [Contributor Convent](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).

To get started, fork the repo, make your changes, add, commit and push the code, then come back here to open a pull request. If you're new to GitHub or open source, [this guide](https://www.freecodecamp.org/news/how-to-make-your-first-pull-request-on-github-3#let-s-make-our-first-pull-request-) or the [git docs](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request) may help you get started, but feel free to reach out if you need any support.

[![contributors](https://contrib.rocks/image?repo=lissy93/go-apod)](https://github.com/Lissy93/go-apod/graphs/contributors)

---

## License

> _**[Lissy93/Go-APOD](https://github.com/Lissy93/go-apod)** is licensed under [MIT](https://github.com/Lissy93/go-apod/blob/master/LICENSE) © [Alicia Sykes](https://aliciasykes.com) 2022._<br>
> <sup align="right">For information, see <a href="https://tldrlegal.com/license/mit-license">TLDR Legal > MIT</a></sup>

<details>
<summary>Expand License</summary>

```
The MIT License (MIT)
Copyright (c) Alicia Sykes <alicia@omg.com> 

Permission is hereby granted, free of charge, to any person obtaining a copy 
of this software and associated documentation files (the "Software"), to deal 
in the Software without restriction, including without limitation the rights 
to use, copy, modify, merge, publish, distribute, sub-license, and/or sell 
copies of the Software, and to permit persons to whom the Software is furnished 
to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included install 
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANT ABILITY, FITNESS FOR A
PARTICULAR PURPOSE AND NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```

</details>

---

<!-- License + Copyright -->
<p  align="center">
  <i>© <a href="https://aliciasykes.com">Alicia Sykes</a> 2022 - Present</i><br>
  <i>Licensed under <a href="https://gist.github.com/Lissy93/143d2ee01ccc5c052a17">MIT</a></i><br>
  <a href="https://github.com/lissy93"><img src="https://cdn.as93.net/84m3gc?w=56" /></a><br>
  <sup>Thanks for visiting :)</sup>
</p>

<!-- Dinosaurs are Awesome -->
<!-- 
                        . - ~ ~ ~ - .
      ..     _      .-~               ~-.
     //|     \ `..~                      `.
    || |      }  }              /       \  \
(\   \\ \~^..'                 |         }  \
 \`.-~  o      /       }       |        /    \
 (__          |       /        |       /      `.
  `- - ~ ~ -._|      /_ - ~ ~ ^|      /- _      `.
              |     /          |     /     ~-.     ~- _
              |_____|          |_____|         ~ - . _ _~_-_
-->

