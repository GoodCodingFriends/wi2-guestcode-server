# wi2-guestcode-server

[![CircleCI](https://img.shields.io/circleci/project/github/GoodCodingFriends/wi2-guestcode-server.svg?style=for-the-badge)](https://circleci.com/gh/GoodCodingFriends/wi2-guestcode-server)

Getting the Guest Code for public Wi-Fi is messy and stressfull? Ok, the agent will do instead of you!

## Environment

- Go 1.9
- Google App Engine(Standard Environment)

## Setup

1. Clone it.
   ```
   $ go get github.com/GoodCodingFriends/wi2-guestcode-server
   ```
2. Create `secret.yaml` including some environment variable specifications:
   ```
   env_variables:
     WI2_GUESTCODE_SERVER_SENDER: "hoge@projectname.appspotmail.com"
     WI2_GUESTCODE_SERVER_TO: "dcs@forguest.wi2.ne.jp"
   ```
3. Deploy it.
   ```
   $ gcloud deploy app.yaml cron.yaml index.yaml
   ```

## Usage

The agent takes the Guest Code periodically by 10 mins.

The code can be got by hitting `/code` entrypoint.

```
$ curl 'https://projectname.appspot.com/code'
DCS9404XWP
```

Save your time and enjoy creative one!
