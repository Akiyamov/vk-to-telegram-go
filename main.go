package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
)

var vk_access_token string
var vk_api_version string
var vk_owner_id string
var vk_post_link string
var vk_link_images []string
var vk_post_text string
var vk_response api.WallGetResponse
var vk_post_last int
var vk_post_requested int

var telegram_bot_token string
var telegram_chat_id string
var telegram_api_url string
var telegram_api telegram_api_params
var tmp []telegram_photo_params

type telegram_api_params struct {
	Chat_id string                  `json:"chat_id"`
	Media   []telegram_photo_params `json:"media"`
}
type telegram_photo_params struct {
	Type_photo string `json:"type"`
	Media      string `json:"media"`
	Caption    string `json:"caption"`
	Parse_mode string `json:"parse_mode"`
}

func Request() {
	vk := api.NewVK(vk_access_token)

	params := api.Params{
		"owner_id": vk_owner_id,
		"count":    1,
		"filter":   "all",
		"v":        vk_api_version,
	}

	err := vk.RequestUnmarshal("wall.get", &vk_response, params)
	if err != nil {
		log.Fatal(err)
	}

	vk_post_requested = vk_response.Items[0].ID

	if vk_post_last == 1 {
		vk_post_last = vk_post_requested
		log.Print("First poll, getting last post ID\n")
	} else if vk_post_last != vk_post_requested {
		vk_post_last = vk_post_requested
		log.Print("New post, initiate post sequence\n")
		PostPhoto()
	} else {
		log.Print("No posts, continue polling\n")
	}
}

func PostPhoto() {
	vk_link_images = make([]string, len(vk_response.Items[0].Attachments))
	for i := range vk_response.Items[0].Attachments {
		vk_link_images[i] = vk_response.Items[0].Attachments[i].Photo.MaxSize().BaseImage.URL
	}
	telegram_api.Chat_id = telegram_chat_id
	tmp = make([]telegram_photo_params, len(vk_link_images))
	for i := range tmp {
		tmp[i].Type_photo = "photo"
		tmp[i].Media = vk_link_images[i]
	}
	tmp[0].Parse_mode = "markdownv2"
	tmp[0].Caption = fmt.Sprintf("%s\n\n*[Ссылка на пост](https://vk.com/wall%s_%d)*", vk_response.Items[0].Text, vk_owner_id, vk_response.Items[0].ID)
	telegram_api.Media = tmp
	tmp_json, err := json.Marshal(telegram_api)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", telegram_api_url, bytes.NewBuffer(tmp_json))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}

func Poll() {
	r := rand.New(rand.NewSource(99))
	c := time.Tick(10 * time.Second)
	for _ = range c {
		Request()
		jitter := time.Duration(r.Int31n(5000)) * time.Millisecond
		time.Sleep(jitter)
	}
}

func main() {
	vk_access_token = os.Getenv("VK_TOKEN")
	vk_api_version = os.Getenv("VK_API_VERSION")
	vk_owner_id = os.Getenv("VK_GROUP_ID")
	vk_post_last = 1
	telegram_bot_token = os.Getenv("TG_TOKEN")
	telegram_chat_id = os.Getenv("TG_CHAT_ID")
	telegram_api_url = fmt.Sprintf("https://api.telegram.org/bot%s/sendMediaGroup", telegram_bot_token)

	Poll()
}

// kessoku_group
// -218803038
