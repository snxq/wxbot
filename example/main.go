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

	rm *wxbot.RequestMessage
}

func NewHandler() http.Handler {
	b, err := wxbot.New("bot_en(中文名)", "token", "access_token")
	if err != nil {
		panic(err)
	}
	return &Handler{b: b}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
		}
	}()

	h.rm, err = h.requestMsg(r)
	if err != nil {
		return
	}

	err = h.b.SendMsg(h.rm.WebHookURL, h.rm.ChatID, h.handleMsg(), wxbot.Text)
	if err != nil {
		return
	}
}

func (h *Handler) requestMsg(r *http.Request) (
	rm *wxbot.RequestMessage, err error) {
	br, err := wxbot.Parse(r)
	if err != nil {
		return nil, err
	}
	sign := h.b.GetSign(br.Timestamp, br.Nonce, br.MessageBody.Encrypt)
	if sign != br.MsgSignature {
		return nil, fmt.Errorf("签名不一致, sign: %s, got: %s", br.MsgSignature, sign)
	}
	return h.b.DecryptRequest(br)
}

func (h *Handler) handleMsg() string {
	msg := strings.TrimSpace(strings.TrimPrefix(
		h.rm.Text.Content, "@"+h.b.GetFullName()))

	// 添加消息处理

	return msg
}
