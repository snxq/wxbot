package wxbot

import (
	"encoding/xml"
	"io"
	"net/http"
)

// Message 请求Body中的Msg信息
type RequestMessage struct {
	From           RequestMessageFrom `xml:"From"`
	WebHookURL     string             `xml:"WebhookUrl"`
	ChatID         string             `xml:"ChatId"`
	GetChatInfoURL string             `xml:"GetChatInfoUrl"`
	MsgID          string             `xml:"MsgId"`
	ChatType       string             `xml:"ChatType"`
	MsgType        string             `xml:"MsgType"`
	Text           RequestMessageText `xml:"Text"`

	IsFirst bool // 为了区分第一次的验证
}

// RequestMessageFrom 消息来源信息
type RequestMessageFrom struct {
	UserID string `xml:"UserId"`
	Name   string `xml:"Name"`  // 中文名
	Alias  string `xml:"Alias"` // 英文名
}

// RequestMessageText 消息内容
type RequestMessageText struct {
	Content string `xml:"Content"`
}

type ResponseMessage struct {
	XMLName       xml.Name     `xml:"xml"`
	MsgType       string       `xml:"MsgType"`
	VisibleToUser string       `xml:"VisibleToUser,omitempty"`
	ResponseText  ResponseText `xml:"Text"`
}

type ResponseText struct {
	Content             CDATA              `xml:"Content"`
	MentionedList       []ResponseTextItem `xml:"MentionedList"`
	MentionedMobileList []ResponseTextItem `xml:"MentionedMobileList"`
}

type ResponseTextItem CDATA

type CDATA struct {
	Value string `xml:",cdata"`
}

type CallBackRequest struct {
	MsgSignature string
	Nonce        string
	Timestamp    string
	MsgEncrypt   string

	MessageBody *MessageBody
}

type MessageBody struct {
	Encrypt string `xml:"Encrypt"`
}

func Parse(r *http.Request) (br *CallBackRequest, err error) {
	args := r.URL.Query()
	br = &CallBackRequest{
		MsgSignature: args.Get("msg_signature"),
		Nonce:        args.Get("nonce"),
		Timestamp:    args.Get("timestamp"),
		MsgEncrypt:   args.Get("msg_encrypt"),
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if len(body) != 0 {
		mb := &MessageBody{}
		if err := xml.Unmarshal(body, mb); err != nil {
			return nil, err
		}
		br.MessageBody = mb
	}

	return br, nil
}

func (rr *CallBackRequest) Body() string {
	switch {
	case len(rr.MsgEncrypt) != 0:
		return rr.MsgEncrypt
	case len(rr.MessageBody.Encrypt) != 0:
		return rr.MessageBody.Encrypt
	default:
		return ""
	}
}

type MsgType string

const (
	Text     MsgType = "text"
	Markdown MsgType = "markdown"
)

type SendMsgRequest struct {
	ChatID  string  `json:"chatid"`
	MsgType MsgType `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}
