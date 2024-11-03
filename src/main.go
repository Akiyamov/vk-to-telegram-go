package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/object"
	"github.com/joho/godotenv"
)

const EnvPath = "/opt/.env"

var vk_access_token string
var vk_api_version string
var vk_owner_id string
var vk_post_last int
var vk_what_to_clean string

var telegram_bot_token string
var telegram_temp_chat_id string
var telegram_chat_id string
var telegram_source_required string
var telegram_api_send_video string
var telegram_api_send_media string
var telegram_api_send_text string

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

func VK_to_TG(post object.WallWallpost) {
	vk_post_requested_ID := post.ID
	vk_post_requested := post

	if vk_post_last == 1 {
		vk_post_last = vk_post_requested_ID
		log.Print("[INFO] First request.\n")
	} else if vk_post_last != vk_post_requested_ID {
		vk_post_last = vk_post_requested_ID
		log.Printf("[INFO] New post vk.com/wall%v_%v\n", vk_post_requested.OwnerID, vk_post_requested.ID)
		PostMessage(vk_post_requested)
		if vk_post_requested.CopyHistory != nil {
			PostMessage(vk_post_requested.CopyHistory[0])
		}
	}
}

func Request_VK() {
	var vk_response api.WallGetResponse

	vk := api.NewVK(vk_access_token)

	params := api.Params{
		"access_token": vk_access_token,
		"owner_id":     vk_owner_id,
		"count":        2,
		"offset":       0,
		"filter":       "all",
		"v":            vk_api_version,
	}

	err := vk.RequestUnmarshal("wall.get", &vk_response, params)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	vk_is_pinned_check := vk_response.Items[0].IsPinned
	if vk_is_pinned_check {
		VK_to_TG(vk_response.Items[1])
	} else {
		VK_to_TG(vk_response.Items[0])
	}
}

func GetGifURL(link string) string {
	resp, err := http.Get(link)

	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	defer resp.Body.Close()

	gifURL := resp.Request.URL

	return fmt.Sprintf("%v", gifURL)
}

func GetAudioURL(owner_id string, audio_id string) string {
	data := url.Values{
		"access_token": {vk_access_token},
		"audios":       {fmt.Sprintf("%s_%s", owner_id, audio_id)},
		"v":            {vk_api_version},
	}

	resp, err := http.PostForm("https://api.vk.com/method/audio.getById?client_id=5776857", data)

	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if dec == nil {
		log.Fatal("Failed to start decoding JSON data")
	}

	json_map := make(map[string]interface{})
	err = dec.Decode(&json_map)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	url_of := fmt.Sprintf("%v", json_map["response"].([]interface{})[0].(map[string]interface{})["url"])
	return url_of
}

func GetVideoURL(owner_id int, vid int) string {
	var url_best string
	oid := owner_id
	id := vid

	payloadi := url.Values{
		"act":         {"show"},
		"al":          {"1"},
		"module":      {"direct"},
		"playlist_id": {fmt.Sprintf("%d_-2", oid)},
		"video":       {fmt.Sprintf("%d_%d", oid, id)},
	}

	req, err := http.NewRequest("POST", "https://vk.com/al_video.php?act=show", strings.NewReader(payloadi.Encode()))
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	req.Header = http.Header{
		"authority":          {"authority"},
		"accept":             {"*/*"},
		"accept-language":    {"ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7"},
		"content-type":       {"application/x-www-form-urlencoded"},
		"dnt":                {"1"},
		"origin":             {"https://vk.com"},
		"referer":            {fmt.Sprintf("https://vk.com/video%d_%d", oid, id)},
		"sec-ch-ua":          {"\"Google Chrome\";v=\"117\", \"Not;A=Brand\";v=\"8\", \"Chromium\";v=\"117\""},
		"sec-ch-ua-mobile":   {"?0"},
		"sec-ch-ua-platform": {"\"Windows\""},
		"sec-fetch-dest":     {"empty"},
		"sec-fetch-mode":     {"cors"},
		"sec-fetch-site":     {"same-origin"},
		"user-agent":         {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"},
		"x-requested-with":   {"XMLHttpRequest"},
	}
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("[ERROR] impossible to send request: %s", err)
	}
	log.Printf("[INFO] GetVideoURL 1: status Code: %d", res.StatusCode)

	defer res.Body.Close()
	dec := json.NewDecoder(res.Body)
	if dec == nil {
		log.Fatal("[ERROR] Failed to start decoding JSON data")
	}
	json_map := make(map[string]interface{})
	err = dec.Decode(&json_map)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	qual_slice := [4]string{"url720", "url480", "url320", "url240"}
	for _, quality := range qual_slice {
		probe_url := json_map["payload"].([]interface{})[1].([]interface{})[4].(map[string]interface{})["player"].(map[string]interface{})["params"].([]interface{})[0].(map[string]interface{})[fmt.Sprintf("%v", quality)]
		if probe_url != nil {
			url_best = fmt.Sprintf("%v", probe_url)
			break
		} else {
			continue
		}
	}

	req, err = http.NewRequest("GET", url_best, nil)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	out, err := os.Create(fmt.Sprintf("/video/%v_%v.mp4", oid, id))
	defer out.Close()
	req.Header = http.Header{
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Accept-Language":           {"ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7"},
		"Connection":                {"keep-alive"},
		"Cookie":                    {"tstc=p"},
		"DNT":                       {"1"},
		"Sec-Fetch-Dest":            {"document"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-Site":            {"none"},
		"Sec-Fetch-User":            {"?1"},
		"Upgrade-Insecure-Requests": {"1"},
		"User-Agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"},
		"sec-ch-ua":                 {"\"Google Chrome\";v=\"117\", \"Not;A=Brand\";v=\"8\", \"Chromium\";v=\"117\""},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {"\"Windows\""},
	}

	client = http.Client{}
	res, err = client.Do(req)
	if err != nil {
		log.Fatalf("[ERROR] impossible to send request: %s", err)
	}
	log.Printf("[INFO] status Code: %d", res.StatusCode)
	defer res.Body.Close()
	_, err = io.Copy(out, res.Body)
	_, err = os.Open(fmt.Sprintf("/video/%d_%d.mp4", oid, id))
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	fw, err := writer.CreateFormFile("video", fmt.Sprintf("%d_%d.mp4", oid, id))
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	fd, err := os.Open(fmt.Sprintf("/video/%d_%d.mp4", oid, id))
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	defer fd.Close()
	_, err = io.Copy(fw, fd)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	formField, err := writer.CreateFormField("chat_id")
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	_, err = formField.Write([]byte(telegram_temp_chat_id))
	writer.Close()
	client = http.Client{}
	req, err = http.NewRequest("POST", telegram_api_send_video, form)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	defer resp.Body.Close()
	data_vid := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&data_vid)
	vid_idd := data_vid["result"].(map[string]interface{})["video"].(map[string]interface{})["file_id"]
	vid_id := fmt.Sprintf("%v", vid_idd)
	dir, err := os.ReadDir("/video")
	for _, d := range dir {
		os.RemoveAll(path.Join([]string{"video", d.Name()}...))
	}
	return vid_id
}

func NoAttachPrepare(text string, tg_src string) {
	telegram_api_text := telegram_api_text_params{}
	telegram_api_text.Text = fmt.Sprintf("%s%s", text, tg_src)
	telegram_api_text.Parse_mode = "html"
	telegram_api_text.Chat_id = telegram_chat_id
	tmp_json, err := json.Marshal(telegram_api_text)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	SendToTelegramNoAttach(tmp_json)
}

func SendToTelegramNoAttach(post_data []byte) {
	req, err := http.NewRequest("POST", telegram_api_send_text, bytes.NewBuffer(post_data))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Print("Response is not 200, ", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		log.Fatalln("response Body:", string(body))
	}
}

func SendToTelegram(post_data []byte) {
	req, err := http.NewRequest("POST", telegram_api_send_media, bytes.NewBuffer(post_data))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Print("Response is not 200, ", resp.Status)
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

func PostMessage(post_vk object.WallWallpost) {
	if len(post_vk.Attachments) != 0 {
		telegram_api_photos := make([]telegram_photo_params, len(post_vk.Attachments))
		telegram_api_audio_params := make([]telegram_audio_params, len(post_vk.Attachments))
		for i := range post_vk.Attachments {
			if post_vk.Attachments[i].Type == "photo" {
				telegram_api_photos[i].Type_photo = "photo"
				telegram_api_photos[i].Media = post_vk.Attachments[i].Photo.MaxSize().URL
			} else if post_vk.Attachments[i].Type == "doc" && post_vk.Attachments[i].Doc.Ext == "gif" {
				telegram_api_photos[i].Type_photo = "video"
				telegram_api_photos[i].Media = fmt.Sprintf("%v", GetGifURL(post_vk.Attachments[i].Doc.Preview.Video.Src))
			} else if post_vk.Attachments[i].Type == "video" {
				telegram_api_photos[i].Type_photo = "video"
				telegram_api_photos[i].Media = fmt.Sprintf("%v", GetVideoURL(post_vk.Attachments[i].Video.OwnerID, post_vk.Attachments[i].Video.ID))
			} else if post_vk.Attachments[i].Type == "audio" {
				telegram_api_audio_params[i].Type_audio = "audio"
				telegram_api_audio_params[i].Media = GetAudioURL(fmt.Sprintf("%v", post_vk.Attachments[i].Audio.OwnerID), fmt.Sprintf("%v", post_vk.Attachments[i].Audio.ID))
			}
		}
		var telegram_source_link string
		if telegram_source_required == "true" {
			if post_vk.Copyright.Link != "" {
				telegram_source_link = fmt.Sprintf("\n\n<a href=\"https://vk.com/wall%d_%d\"><b>Ссылка на пост</b></a><b> | </b><a href=\"%s\"><b>Ссылка на источник</b></a>", post_vk.OwnerID, post_vk.ID, post_vk.Copyright.Link)
			} else {
				telegram_source_link = fmt.Sprintf("\n\n<a href=\"https://vk.com/wall%d_%d\"><b>Ссылка на пост</b></a>", post_vk.OwnerID, post_vk.ID)
			}
		}
		telegram_api_media := telegram_api_params{}
		telegram_api_media.Chat_id = telegram_chat_id
		telegram_api_photos = DeleteEmptyMedia(telegram_api_photos)
		telegram_api_audio := telegram_api_params_audio{}
		telegram_api_audio.Chat_id = telegram_chat_id
		telegram_api_audio_params = DeleteEmptyAudio(telegram_api_audio_params)
		if len(telegram_api_photos) == 0 && len(telegram_api_audio_params) == 0 {
			log.Print("[DEBUG] Poll or something else in post, maybe add them?")
			telegram_api_descr_cleansed := strings.ReplaceAll(post_vk.Text, vk_what_to_clean, "")
			NoAttachPrepare(telegram_api_descr_cleansed, telegram_source_link)
		} else if len(telegram_api_photos) != 0 && len(telegram_api_audio_params) == 0 {
			telegram_api_photos[0].Parse_mode = "html"
			if post_vk.Attachments[0].Type == "video" && post_vk.Attachments[0].Video.Type == "short_video" {
				telegram_api_descr_cleansed := strings.ReplaceAll(post_vk.Attachments[0].Video.Description, vk_what_to_clean, "")
				telegram_api_photos[0].Caption = fmt.Sprintf("%s%s", telegram_api_descr_cleansed, telegram_source_link)
			} else {
				telegram_api_descr_cleansed := strings.ReplaceAll(post_vk.Text, vk_what_to_clean, "")
				telegram_api_photos[0].Caption = fmt.Sprintf("%s%s", telegram_api_descr_cleansed, telegram_source_link)
			}
			telegram_api_media.Media = telegram_api_photos
			tmp_json, err := json.Marshal(telegram_api_media)
			if err != nil {
				log.Fatalf("[ERROR] %v", err)
			}
			SendToTelegram(tmp_json)
		} else if len(telegram_api_photos) == 0 && len(telegram_api_audio_params) != 0 {
			telegram_api_audio_params[0].Parse_mode = "html"
			telegram_api_descr_cleansed := strings.ReplaceAll(post_vk.Text, vk_what_to_clean, "")
			telegram_api_photos[0].Caption = fmt.Sprintf("%s%s", telegram_api_descr_cleansed, telegram_source_link)
			telegram_api_audio.Media = telegram_api_audio_params
			tmp_json, err := json.Marshal(telegram_api_audio)
			if err != nil {
				log.Fatalf("[ERROR] %v", err)
			}
			SendToTelegram(tmp_json)
		} else if len(telegram_api_photos) != 0 && len(telegram_api_audio_params) != 0 {
			telegram_api_photos[0].Parse_mode = "html"
			telegram_api_descr_cleansed := strings.ReplaceAll(post_vk.Text, vk_what_to_clean, "")
			telegram_api_photos[0].Caption = fmt.Sprintf("%s%s", telegram_api_descr_cleansed, telegram_source_link)
			telegram_api_media.Media = telegram_api_photos
			tmp_json, err := json.Marshal(telegram_api_media)
			if err != nil {
				log.Fatalf("[ERROR] %v", err)
			}
			SendToTelegram(tmp_json)
			telegram_api_audio_params[0].Parse_mode = "html"
			telegram_api_audio_params[0].Caption = fmt.Sprintf("")
			telegram_api_audio.Media = telegram_api_audio_params
			tmp_json_aud, err := json.Marshal(telegram_api_audio)
			if err != nil {
				log.Fatalf("[ERROR] %v", err)
			}
			SendToTelegram(tmp_json_aud)
		}
	} else {
		var telegram_source_link string
		if telegram_source_required == "true" {
			if post_vk.Copyright.Link != "" {
				telegram_source_link = fmt.Sprintf("\n\n<a href=\"https://vk.com/wall%d_%d\"><b>Ссылка на пост</b></a><b> | </b><a href=\"%s\"><b>Ссылка на источник</b></a>", post_vk.OwnerID, post_vk.ID, post_vk.Copyright.Link)
			} else {
				telegram_source_link = fmt.Sprintf("\n\n<a href=\"https://vk.com/wall%d_%d\"><b>Ссылка на пост</b></a>", post_vk.OwnerID, post_vk.ID)
			}
		}
		log.Print("Post has no media, post caption only\n")
		telegram_api_descr_cleansed := strings.ReplaceAll(post_vk.Text, vk_what_to_clean, "")
		NoAttachPrepare(telegram_api_descr_cleansed, telegram_source_link)
	}
}

func Poll() {
	r := rand.New(rand.NewSource(99))
	c := time.Tick(10 * time.Second)
	for range c {
		Request_VK()
		jitter := time.Duration(r.Int31n(5000)) * time.Millisecond
		time.Sleep(jitter)
	}
}

func LoadEnv() {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Error generating env dir")
	}
	dir := filepath.Join(filepath.Dir(f), "../..", EnvPath)

	err := godotenv.Load(dir)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	LoadEnv()
	vk_access_token = os.Getenv("VK_TOKEN")
	vk_api_version = os.Getenv("VK_API_VERSION")
	vk_owner_id = os.Getenv("VK_GROUP_ID")
	vk_post_last = 1
	telegram_bot_token := os.Getenv("TG_TOKEN")
	telegram_chat_id = os.Getenv("TG_CHAT_ID")
	telegram_temp_chat_id = os.Getenv("TG_TEMP_CHAT_ID")
	telegram_source_required = os.Getenv("TG_SRC")
	telegram_api_send_media = fmt.Sprintf("http://telegram-bot-api:8081/bot%s/sendMediaGroup", telegram_bot_token)
	telegram_api_send_text = fmt.Sprintf("http://telegram-bot-api:8081/bot%s/sendMessage", telegram_bot_token)
	telegram_api_send_video = fmt.Sprintf("http://telegram-bot-api:8081/bot%s/sendVideo", telegram_bot_token)
	vk_what_to_clean = os.Getenv("CLEAN_TXT")

	Poll()
}
