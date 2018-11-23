package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.mercari.io/datastore/aedatastore"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

var mailMsgStub = `Date: Mon, 23 Jun 2019 11:40:36 -0400
From: 株式会社ワイヤ・アンド・ワイヤレス <cs-info@wi2.co.jp>
To: Agent <agent@hoge.appspotmail.com>
Subject: Wi2 300 ゲストコードのお知らせ
Content-Type: text/html; charset="UTF-8"

Wi2　300　ゲストサービスのお申し込みありがとうございます。

ゲストコードをご確認ください。

■お客様のゲストコード━━━━━━━━━━━━━━━━━━━
ゲストコード　：　DCS7270BRQ
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

================================================================
【ご利用いただく前に】
初回ログインから3時間のご利用が可能です。また、当ゲストコードは
当施設でご利用ください。

なお、ご利用前には、ログイン画面に掲示してあります「Wi2　フリーWi-Fiサービス
利用規約」（以下規約）をお読みください。ご利用になられた場合は規約に同意した
ものとみなします。

ゲストコードによりご利用いただくWi-Fiネットワークの通信は、暗号化
されておらず通信を傍受される恐れがあります。
あらかじめご了承頂き、ご利用くださいます様、よろしくお願いいたします。
================================================================

それでは、インターネットをお楽しみください。


■サービスに関するお問い合わせ━━━━━━━━━━━━━━━

　Wi2(ワイツー)カスタマーセンター

　メールフォームによるお問い合わせ
　⇒https://service.wi2.ne.jp/wi2net/contact/
　　※本メールアドレスは送信専用です

　お電話によるお問い合わせ
　⇒0120-858-306（受付時間　10:00～19:00/年中無休）
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
■公衆無線LANサービス　Wi2　300　http://wi2.co.jp/jp/300/
`

func TestCode(t *testing.T) {
	mux := http.NewServeMux()

	a := &app{
		to: "to@example.com",
	}
	a.handle(mux)

	w := httptest.NewRecorder()

	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/code", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := appengine.NewContext(req)

	// Setup datastore
	client, err := aedatastore.FromContext(ctx)
	if err != nil {
		t.Fatal(err)
	}

	parentKey, err := client.Put(ctx, client.NameKey("code", "dcs", nil), &struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	codeKey, err := client.Put(ctx, client.IncompleteKey("code", parentKey), &codeEntity{
		Code:    "DCS1234ABC",
		Used:    false,
		Created: time.Date(1999, 11, 7, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Put(ctx, client.IncompleteKey("code", parentKey), &codeEntity{
		Code:    "DCS5678DEF",
		Used:    false,
		Created: time.Date(1996, 7, 30, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)
	t.Log(w.Body)

	var updated codeEntity
	if err := client.Get(ctx, codeKey, &updated); err != nil {
		t.Fatal(err)
	}

	if updated.Code != "DCS1234ABC" {
		t.Errorf("the Code of updated entity is expected DCS1234ABC, but %v", updated.Code)
	}

	if !updated.Used {
		t.Errorf("the Used of updated entity is expected true, but %v", updated.Used)
	}
}

func TestComposeMessage(t *testing.T) {
	a := &app{
		to: "to@example.com",
	}

	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/code", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := appengine.NewContext(req)

	msg := a.composeMessage(ctx)

	appID := appengine.AppID(ctx)
	t.Logf("AppID: %s", appID)
	if substr := fmt.Sprintf("@%s.appspotmail.com", appID); !strings.Contains(msg.Sender, substr) {
		t.Errorf("the Sender is expected contains `%s` but does not: %s", substr, msg.Sender)
	}

	if len(msg.To) != 1 || msg.To[0] != a.to {
		t.Errorf("the To is expected %s but %#+v", a.to, msg.To)
	}
}

func TestReceivingMail(t *testing.T) {
	mux := http.NewServeMux()

	a := &app{
		to: "to@example.com",
	}
	a.handle(mux)

	w := httptest.NewRecorder()

	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("POST", "/_ah/mail/agent@hoge.appspotmail.com", strings.NewReader(mailMsgStub))
	if err != nil {
		t.Fatal(err)
	}

	ctx := appengine.NewContext(req)

	// Setup datastore
	client, err := aedatastore.FromContext(ctx)
	if err != nil {
		t.Fatal(err)
	}

	parentKey, err := client.Put(ctx, client.NameKey("code", "dcs", nil), &struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Put(ctx, client.IncompleteKey("code", parentKey), &codeEntity{
		Code:    "DCS1234ABC",
		Used:    true,
		Created: time.Date(1996, 7, 30, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)
	t.Log(w.Body)

	var codes []*codeEntity
	if _, err := client.GetAll(ctx, client.NewQuery("code").Ancestor(dcsKey(client)).Filter("used =", false).Order("-created").Limit(1), &codes); err != nil {
		t.Fatal(err)
	}

	if len(codes) != 1 {
		t.Fatalf("the len of codes got by the query is expected 1, but %v(the value is %v)", len(codes), codes)
	}

	code := codes[0]

	if code.Code != "DCS7270BRQ" {
		t.Errorf("the Code of got entity is expected DCS7270BRQ, but %v", code.Code)
	}

	if code.Used {
		t.Errorf("the Used of got entity is expected false, but %v", code.Used)
	}
}

func TestReceivingMailWithExistingCode(t *testing.T) {
	mux := http.NewServeMux()

	a := &app{
		to: "to@example.com",
	}
	a.handle(mux)

	w := httptest.NewRecorder()

	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("POST", "/_ah/mail/agent@hoge.appspotmail.com", strings.NewReader(mailMsgStub))
	if err != nil {
		t.Fatal(err)
	}

	ctx := appengine.NewContext(req)

	// Setup datastore
	client, err := aedatastore.FromContext(ctx)
	if err != nil {
		t.Fatal(err)
	}

	parentKey, err := client.Put(ctx, client.NameKey("code", "dcs", nil), &struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Put(ctx, client.IncompleteKey("code", parentKey), &codeEntity{
		Code:    "DCS7270BRQ",
		Used:    true,
		Created: time.Date(1996, 7, 30, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)
	t.Log(w.Body)

	var codes []*codeEntity
	if _, err := client.GetAll(ctx, client.NewQuery("code").Ancestor(dcsKey(client)).Filter("used =", false).Order("-created").Limit(1), &codes); err != nil {
		t.Fatal(err)
	}

	if len(codes) != 0 {
		t.Fatalf("the len of codes got by the query is expected 0, but %v(the value is %v)", len(codes), codes)
	}
}
