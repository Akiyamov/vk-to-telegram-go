package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
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
var vk_response_repost api.WallGetResponse
var vk_post_last int
var vk_post_requested int

var telegram_bot_token string
var telegram_chat_id string
var telegram_is_video bool
var telegram_is_audio bool
var telegram_is_unsupported bool
var telegram_api_send_media string
var telegram_api_send_text string
var telegram_api_text telegram_api_text_params
var telegram_api_media telegram_api_params
var telegram_api_audio telegram_api_params_audio
var telegram_api_photos []telegram_photo_params
var telegram_api_audio_params []telegram_audio_params

type telegram_api_text_params struct {
	Chat_id    string `json:"chat_id"`
	Text       string `json:"text"`
	Parse_mode string `json:"parse_mode"`
}

type telegram_api_params struct {
	Chat_id string                  `json:"chat_id"`
	Media   []telegram_photo_params `json:"media"`
}

type telegram_api_params_audio struct {
	Chat_id string                  `json:"chat_id"`
	Media   []telegram_audio_params `json:"media"`
}

type telegram_photo_params struct {
	Type_photo string `json:"type"`
	Media      string `json:"media"`
	Caption    string `json:"caption"`
	Parse_mode string `json:"parse_mode"`
}

type telegram_audio_params struct {
	Type_audio string `json:"type"`
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
		PostMessage(vk_response)
		if vk_response.Items[0].CopyHistory != nil {
			vk_response_repost.Items = vk_response.Items[0].CopyHistory
			PostMessage(vk_response_repost)
		}
	} else {
		log.Print("No posts, continue polling\n")
	}
}

func GetAudioURL(owner_id string, audio_id string) string {
	data := url.Values{
		"access_token": {vk_access_token},
		"audios":       {fmt.Sprintf("%s_%s", owner_id, audio_id)},
		"v":            {vk_api_version},
	}

	resp, err := http.PostForm("https://api.vk.com/method/audio.getById?client_id=5776857", data)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if dec == nil {
		panic("Failed to start decoding JSON data")
	}

	json_map := make(map[string]interface{})
	err = dec.Decode(&json_map)
	if err != nil {
		panic(err)
	}

	url_of := fmt.Sprintf("%v", json_map["response"].([]interface{})[0].(map[string]interface{})["url"])
	return url_of
}

func SendToTelegram(post_data []byte) {
	req, err := http.NewRequest("POST", telegram_api_send_media, bytes.NewBuffer(post_data))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Print("Response is not 200, %s", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		log.Fatalln("response Body:", string(body))
	}
}

func DeleteEmptyAudio(s []telegram_audio_params) []telegram_audio_params {
	var r []telegram_audio_params
	for _, str := range s {
		if (str != telegram_audio_params{}) {
			r = append(r, str)
		}
	}
	return r
}

func DeleteEmptyMedia(s []telegram_photo_params) []telegram_photo_params {
	var r []telegram_photo_params
	for _, str := range s {
		if (str != telegram_photo_params{}) {
			r = append(r, str)
		}
	}
	return r
}

func PostMessage(post_response api.WallGetResponse) {
	if len(post_response.Items[0].Attachments) > 0 {
		telegram_api_photos = make([]telegram_photo_params, len(post_response.Items[0].Attachments))
		telegram_api_audio_params = make([]telegram_audio_params, len(post_response.Items[0].Attachments))
		for i := range post_response.Items[0].Attachments {
			if post_response.Items[0].Attachments[i].Type == "photo" {
				telegram_api_photos[i].Type_photo = "photo"
				telegram_api_photos[i].Media = post_response.Items[0].Attachments[i].Photo.MaxSize().URL
			} else if post_response.Items[0].Attachments[i].Type == "doc" && post_response.Items[0].Attachments[i].Doc.Ext == "gif" {
				telegram_api_photos[i].Type_photo = "video"
				telegram_api_photos[i].Media = post_response.Items[0].Attachments[i].Doc.URL
			} else if post_response.Items[0].Attachments[i].Type == "audio" {
				telegram_api_audio_params[i].Type_audio = "audio"
				telegram_api_audio_params[i].Media = GetAudioURL(fmt.Sprintf("%v", post_response.Items[0].Attachments[i].Audio.OwnerID), fmt.Sprintf("%v", post_response.Items[0].Attachments[i].Audio.ID))
				telegram_api_audio.Chat_id = telegram_chat_id
				telegram_api_audio_params[i].Parse_mode = "html"
				telegram_api_audio_params[i].Caption = fmt.Sprintf("<b>Пидорасня не дает сделать в одном посте, поэтому отдельно</b>")
				telegram_api_audio_params = DeleteEmptyAudio(telegram_api_audio_params)
				telegram_api_audio.Media = telegram_api_audio_params
				log.Print(telegram_api_audio)
				tmp_json, err := json.Marshal(telegram_api_audio)
				if err != nil {
					log.Fatal(err)
				}
				SendToTelegram(tmp_json)
			}
		}
		telegram_api_media.Chat_id = telegram_chat_id
		telegram_api_photos[0].Parse_mode = "html"
		telegram_api_photos[0].Caption = fmt.Sprintf("%s\n\n<a href=\"https://vk.com/wall%s_%d\"><b>Ссылка на пост</b></a>", post_response.Items[0].Text, vk_owner_id, post_response.Items[0].ID)
		telegram_api_photos = DeleteEmptyMedia(telegram_api_photos)
		telegram_api_media.Media = telegram_api_photos
		log.Print(telegram_api_media)
		tmp_json, err := json.Marshal(telegram_api_media)
		if err != nil {
			log.Fatal(err)
		}
		SendToTelegram(tmp_json)
	} else {
		log.Print("Post has no media, post caption only\n")
		telegram_api_text.Text = fmt.Sprintf("%s\n\n<a href=\"https://vk.com/wall%s_%d\"><b>Ссылка на пост</b></a>", post_response.Items[0].Text, vk_owner_id, post_response.Items[0].ID)
		telegram_api_text.Parse_mode = "html"
		telegram_api_text.Chat_id = telegram_chat_id
		tmp_json, err := json.Marshal(telegram_api_text)
		if err != nil {
			log.Fatal(err)
		}
		req, err := http.NewRequest("POST", telegram_api_send_text, bytes.NewBuffer(tmp_json))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Print("Response is not 200, %s", resp.Status)
			body, _ := io.ReadAll(resp.Body)
			log.Fatalln("response Body:", string(body))
		}
	}
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
	telegram_api_send_media = fmt.Sprintf("https://api.telegram.org/bot%s/sendMediaGroup", telegram_bot_token)
	telegram_api_send_text = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegram_bot_token)

	Poll()
}
