package wxbot

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// Bot
// DecryptBody 解密请求体
type Bot interface {
	DecryptRequest(req *CallBackRequest) (*RequestMessage, error)
	Decrypt(msg string) ([]byte, error)
	Encrypt(msg string) (data []byte)
	GetSign(timestamp, nonce, data string) string
	GetFullName() string
	SendMsg(addr, chatid, content string, msgtype MsgType) error
}

// impl Bot
type bot struct {
	FullName           string
	Token, AccessToken string
	AESKey             []byte
	DeBlockMode        cipher.BlockMode
	EnBlockMode        cipher.BlockMode
}

func (b *bot) GetFullName() string {
	return b.FullName
}

// New return Bot
// token, at(access_token) 通过机器人设置获取
func New(fullname, token, at string) (b Bot, err error) {
	bot := &bot{Token: token, AccessToken: at, FullName: fullname}

	bot.AESKey, err = base64.StdEncoding.DecodeString(bot.AccessToken + "=")
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(bot.AESKey)
	if err != nil {
		return nil, err
	}
	bot.DeBlockMode = cipher.NewCBCDecrypter(block, bot.AESKey[:aes.BlockSize])
	bot.EnBlockMode = cipher.NewCBCEncrypter(block, bot.AESKey[:aes.BlockSize])
	return bot, nil
}

func (b *bot) DecryptRequest(req *CallBackRequest) (*RequestMessage, error) {
	plain, err := b.Decrypt(req.Body())
	if err != nil {
		return nil, err
	}
	msg := &RequestMessage{}
	if !strings.HasPrefix(string(plain), "<xml>") {
		msg.IsFirst = true
		return msg, nil
	}
	if err = xml.Unmarshal(plain, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (b *bot) Decrypt(msg string) ([]byte, error) {
	pt, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return nil, err
	}
	if len(pt) < aes.BlockSize || len(pt)%aes.BlockSize != 0 {
		return nil, errors.New("plaintext size is wrong")
	}
	b.DeBlockMode.CryptBlocks(pt, pt)

	// 16随机字节 + 4字节 length
	return pt[20 : binary.BigEndian.Uint32(pt[16:20])+20], nil
}

func (b *bot) Encrypt(msg string) (data []byte) {
	var buf bytes.Buffer
	// 随机字符串前缀，这里置空
	prefix := make([]byte, 16)
	buf.Write(prefix)

	// 写入消息长度信息
	msgsize := make([]byte, 4)
	binary.BigEndian.PutUint32(msgsize, uint32(len(msg)))
	buf.Write(msgsize)

	// 写入消息
	buf.WriteString(msg)

	// 补全 block 长度
	tmp := make([]byte, 32-(buf.Len()%32))
	buf.Write(tmp)

	data = make([]byte, buf.Len())
	b.EnBlockMode.CryptBlocks(data, buf.Bytes())
	crypted := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(crypted, data)
	return crypted
}

// GetSign 获取签名
func (b *bot) GetSign(timestamp, nonce, data string) string {
	arr := []string{b.Token, timestamp, nonce, data}
	sort.Strings(arr)

	sha := sha1.New()
	for _, x := range arr {
		sha.Write([]byte(x))
	}
	return fmt.Sprintf("%x", sha.Sum(nil))
}

func (b *bot) SendMsg(addr, chatid, content string, msgtype MsgType) error {
	payload := &SendMsgRequest{MsgType: msgtype, ChatID: chatid}
	switch msgtype {
	case "text":
		payload.Text.Content = content
	default:
		//
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = http.Post(addr, "application/json", bytes.NewReader(body))
	return err
}
