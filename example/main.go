package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/snxq/wxbot"
)

func main() {
	if err := http.ListenAndServe(":80", NewHandler()); err != nil {
		panic(err)
	}
}

type Handler struct {
	b wxbot.Bot
}

func NewHandler() http.Handler {
	b, err := wxbot.New("bot_en(中文名)", "token", "access_token")
	if err != nil {
		panic(err)
	}
	return &Handler{b: b}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.handle(r); err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
}

func (h *Handler) handle(r *http.Request) (err error) {
	br, err := wxbot.Parse(r)
	if err != nil {
		return err
	}
	sign := h.b.GetSign(br.Timestamp, br.Nonce, br.MessageBody.Encrypt)
	if sign != br.MsgSignature {
		return fmt.Errorf("签名不一致, sign: %s, got: %s\n", br.MsgSignature, sign)
	}
	msg, err := h.b.DecryptRequest(br)
	if err != nil {
		return err
	}
	content := strings.TrimSpace(strings.TrimPrefix(
		msg.Text.Content, "@"+h.b.GetFullName()))

	return h.b.SendMsg(msg.WebHookURL, msg.ChatID, content, wxbot.Text)
}
