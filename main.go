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
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/joho/godotenv"
)

const EnvPath = "/opt/.env"

var vk_access_token string
var vk_api_version string
var vk_owner_id string
var vk_post_last int
var vk_post_requested int

var telegram_bot_token string
var telegram_temp_chat_id string
var telegram_chat_id string
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

func Request() {
	var vk_response api.WallGetResponse
	var vk_response_repost api.WallGetResponse

	vk := api.NewVK(vk_access_token)

	params := api.Params{
		"access_token": vk_access_token,
		"owner_id":     vk_owner_id,
		"count":        1,
		"filter":       "all",
		"v":            vk_api_version,
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
		log.Fatal(err)
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
		log.Fatalf("impossible to send request: %s", err)
	}
	log.Printf("status Code: %d", res.StatusCode)

	defer res.Body.Close()
	dec := json.NewDecoder(res.Body)
	if dec == nil {
		panic("Failed to start decoding JSON data")
	}
	json_map := make(map[string]interface{})
	err = dec.Decode(&json_map)
	if err != nil {
		panic(err)
	}
	url240 := json_map["payload"].([]interface{})[1].([]interface{})[4].(map[string]interface{})["player"].(map[string]interface{})["params"].([]interface{})[0].(map[string]interface{})["url240"]
	url360 := json_map["payload"].([]interface{})[1].([]interface{})[4].(map[string]interface{})["player"].(map[string]interface{})["params"].([]interface{})[0].(map[string]interface{})["url360"]
	url480 := json_map["payload"].([]interface{})[1].([]interface{})[4].(map[string]interface{})["player"].(map[string]interface{})["params"].([]interface{})[0].(map[string]interface{})["url480"]
	url720 := json_map["payload"].([]interface{})[1].([]interface{})[4].(map[string]interface{})["player"].(map[string]interface{})["params"].([]interface{})[0].(map[string]interface{})["url720"]
	if url720 == nil {
		if url480 == nil {
			if url360 == nil {
				if url240 == nil {
					log.Fatal("Ты еблан")
				} else {
					url_best = fmt.Sprintf("%v", url240)
				}
			} else {
				url_best = fmt.Sprintf("%v", url360)
			}
		} else {
			url_best = fmt.Sprintf("%v", url480)
		}
	} else {
		url_best = fmt.Sprintf("%v", url720)
	}

	req, err = http.NewRequest("GET", url_best, nil)
	if err != nil {
		log.Fatal(err)
	}
	out, err := os.Create(fmt.Sprintf("%v_%v.mp4", oid, id))
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
	// send the request
	res, err = client.Do(req)
	if err != nil {
		log.Fatalf("impossible to send request: %s", err)
	}
	log.Printf("status Code: %d", res.StatusCode)
	defer res.Body.Close()
	n, err := io.Copy(out, res.Body)
	fmt.Printf("", n)
	respo, err := os.Open(fmt.Sprintf("%d_%d.mp4", oid, id))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("", respo)
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	fw, err := writer.CreateFormFile("video", fmt.Sprintf("%d_%d.mp4", oid, id))
	if err != nil {
		log.Fatal(err)
	}
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fd, err := os.Open(fmt.Sprintf("%v/%d_%d.mp4", exPath, oid, id))
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()
	_, err = io.Copy(fw, fd)
	if err != nil {
		log.Fatal(err)
	}
	formField, err := writer.CreateFormField("chat_id")
	if err != nil {
		log.Fatal(err)
	}
	_, err = formField.Write([]byte(telegram_temp_chat_id))
	writer.Close()
	client = http.Client{}
	req, err = http.NewRequest("POST", telegram_api_send_video, form)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	data_vid := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&data_vid)
	vid_idd := data_vid["result"].(map[string]interface{})["video"].(map[string]interface{})["file_id"]
	vid_id := fmt.Sprintf("%v", vid_idd)
	return vid_id
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

func PostMessage(post_response api.WallGetResponse) {
	if len(post_response.Items[0].Attachments) > 0 {
		telegram_api_photos := make([]telegram_photo_params, len(post_response.Items[0].Attachments))
		telegram_api_audio_params := make([]telegram_audio_params, len(post_response.Items[0].Attachments))
		for i := range post_response.Items[0].Attachments {
			if post_response.Items[0].Attachments[i].Type == "photo" {
				telegram_api_photos[i].Type_photo = "photo"
				telegram_api_photos[i].Media = post_response.Items[0].Attachments[i].Photo.MaxSize().URL
			} else if post_response.Items[0].Attachments[i].Type == "doc" && post_response.Items[0].Attachments[i].Doc.Ext == "gif" {
				telegram_api_photos[i].Type_photo = "video"
				telegram_api_photos[i].Media = post_response.Items[0].Attachments[i].Doc.URL
			} else if post_response.Items[0].Attachments[i].Type == "video" {
				telegram_api_photos[i].Type_photo = "video"
				GetVideoURL(post_response.Items[0].Attachments[i].Video.OwnerID, post_response.Items[0].Attachments[i].Video.ID)
				telegram_api_photos[i].Media = fmt.Sprintf("%v", GetVideoURL(post_response.Items[0].Attachments[i].Video.OwnerID, post_response.Items[0].Attachments[i].Video.ID))
			} else if post_response.Items[0].Attachments[i].Type == "audio" {
				telegram_api_audio_params[i].Type_audio = "audio"
				telegram_api_audio_params[i].Media = GetAudioURL(fmt.Sprintf("%v", post_response.Items[0].Attachments[i].Audio.OwnerID), fmt.Sprintf("%v", post_response.Items[0].Attachments[i].Audio.ID))
				telegram_api_audio := telegram_api_params_audio{}
				telegram_api_audio.Chat_id = telegram_chat_id
				telegram_api_audio_params[i].Parse_mode = "html"
				telegram_api_audio_params[i].Caption = fmt.Sprintf("<b>Пидорасня не дает сделать в одном посте, поэтому отдельно</b>")
				telegram_api_audio_params = DeleteEmptyAudio(telegram_api_audio_params)
				telegram_api_audio.Media = telegram_api_audio_params
				log.Print(telegram_api_audio)
			}
		}
		telegram_api_media := telegram_api_params{}
		telegram_api_media.Chat_id = telegram_chat_id
		telegram_api_photos[0].Parse_mode = "html"
		telegram_api_photos[0].Caption = fmt.Sprintf("%s\n\n<a href=\"https://vk.com/wall%d_%d\"><b>Ссылка на пост</b></a>", post_response.Items[0].Text, post_response.Items[0].OwnerID, post_response.Items[0].ID)
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
		telegram_api_text := telegram_api_text_params{}
		telegram_api_text.Text = fmt.Sprintf("%s\n\n<a href=\"https://vk.com/wall%d_%d\"><b>Ссылка на пост</b></a>", post_response.Items[0].Text, post_response.Items[0].OwnerID, post_response.Items[0].ID)
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
	telegram_api_send_media = fmt.Sprintf("http://127.0.0.1:8081/bot%s/sendMediaGroup", telegram_bot_token)
	telegram_api_send_text = fmt.Sprintf("http://127.0.0.1:8081/bot%s/sendMessage", telegram_bot_token)
	telegram_api_send_video = fmt.Sprintf("http://127.0.0.1:8081/bot%s/sendVideo", telegram_bot_token)

	Poll()
}
