package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"time"

	"go.mercari.io/datastore"
	"go.mercari.io/datastore/aedatastore"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	gaemail "google.golang.org/appengine/mail"
)

func main() {
	to := os.Getenv("WI2_GUESTCODE_SERVER_TO")
	if to == "" {
		panic("WI2_GUESTCODE_SERVER_TO is not set")
	}

	a := &app{
		to: to,
	}

	a.handle(http.DefaultServeMux)
	appengine.Main()
}

type codeEntity struct {
	Code    string    `datastore:"code"`
	Used    bool      `datastore:"used"`
	Created time.Time `datastore:"created"`
}

type app struct {
	to string
}

func (a *app) handle(mux *http.ServeMux) {
	// Request code
	mux.HandleFunc("/code", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		client, err := aedatastore.FromContext(ctx)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}
		defer client.Close()

		var code *codeEntity
		if _, err := client.RunInTransaction(ctx, func(tx datastore.Transaction) error {
			var codes []*codeEntity
			keys, err := client.GetAll(ctx, client.NewQuery("code").Transaction(tx).Ancestor(dcsKey(client)).Filter("used =", false).Order("-created").Limit(1), &codes)
			if err != nil {
				return err
			}

			if len(codes) == 0 || codes[0].Used {
				log.Infof(ctx, "could not get code")

				w.Write([]byte("wait please"))
				return nil
			}

			log.Infof(ctx, "%#+v", codes)
			code = codes[0]
			key := keys[0]

			code.Used = true

			if _, err := tx.Put(key, code); err != nil {
				return err
			}

			log.Infof(ctx, "updated")

			return nil
		}); err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}

		if code != nil {
			w.Write([]byte(code.Code))
		}
	})

	// Check code availability cron
	mux.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		client, err := aedatastore.FromContext(ctx)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}
		defer client.Close()

		var codes []*codeEntity
		if _, err := client.GetAll(ctx, client.NewQuery("code").Ancestor(dcsKey(client)).Filter("used =", false).Order("-created").Limit(1), &codes); err != nil {
			log.Errorf(ctx, "%v", err)
		}

		log.Infof(ctx, "%#+v", codes)

		if len(codes) > 0 && !codes[0].Used {
			log.Infof(ctx, "code available")
			return
		}

		if err := a.sendMail(ctx); err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}

		w.Write([]byte("sent"))
	})

	// Receive mail
	mux.HandleFunc("/_ah/mail/", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		client, err := aedatastore.FromContext(ctx)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}
		defer client.Close()

		msg, err := mail.ReadMessage(r.Body)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}

		content, err := ioutil.ReadAll(msg.Body)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}

		log.Infof(ctx, "Received mail: %v", string(content))

		code := regexp.MustCompile(`DCS[\w\d]+`).FindString(string(content))
		if code == "" {
			log.Infof(ctx, "No code in the reply. Retrying to send mail.")
			if err := a.sendMail(ctx); err != nil {
				log.Errorf(ctx, "%v", err)
				return
			}
		} else {
			count, err := client.Count(ctx, client.NewQuery("code").Filter("code =", code).Ancestor(dcsKey(client)).Limit(1))
			if err != nil {
				log.Errorf(ctx, "error on counting the code: %v", err)
				return
			}
			if count == 1 {
				log.Errorf(ctx, "Received old code. Retrying to send mail.")
				if err := a.sendMail(ctx); err != nil {
					log.Errorf(ctx, "%v", err)
				}
				return
			}

			k := client.IncompleteKey("code", dcsKey(client))
			if _, err := client.Put(ctx, k, &codeEntity{
				Code:    code,
				Used:    false,
				Created: time.Now(),
			}); err != nil {
				log.Errorf(ctx, "%v", err)
				return
			}
			log.Infof(ctx, "put new record")
		}
	})
}

func (a *app) sendMail(ctx context.Context) error {
	if err := gaemail.Send(ctx, a.composeMessage(ctx)); err != nil {
		return err
	}

	return nil
}

func (a *app) composeMessage(ctx context.Context) *gaemail.Message {
	return &gaemail.Message{
		Sender:  fmt.Sprintf("<%s@%s.appspotmail.com>", randstr(), appengine.AppID(ctx)),
		To:      []string{a.to},
		Subject: "",
		Body:    "",
	}
}

func randstr() string {
	letters := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	str := make([]byte, 0, 8)
	rand.Seed(time.Now().Unix())
	for i := 0; i < 8; i++ {
		li := rand.Intn(len(letters))
		str = append(str, letters[li])
	}

	return string(str)
}

func dcsKey(client datastore.Client) datastore.Key {
	return client.NameKey("code", "dcs", nil)
}
