package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zboyco/pachong/model"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	fmt.Print("请输入花瓣网画板地址: ")
	reader := bufio.NewScanner(os.Stdin)
	url := ""
	if reader.Scan() {
		url = reader.Text()
	}

	board, err := getIndexPage(url)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir(fmt.Sprintf("./%v", board.Title), os.ModePerm)
	if err != nil && os.IsNotExist(err) {
		log.Fatal(err)
	}

	interval := rand.Intn(600) + 800
	time.Sleep(time.Millisecond * time.Duration(interval))

	count := 0
	lastPinID := 0
	for {
		for _, pin := range board.Pins {
			count++
			lastPinID = pin.PinID
			fmt.Print(count)
			fmt.Print(" >>> ")
			if pin.Trusted {
				err := downloadImage(board.Title, pin)
				if os.IsExist(err) {
					fmt.Println(pin.PinID, "已存在,跳过...")
					continue
				}
				if err != nil {
					fmt.Println(pin.PinID, err)
				} else {
					fmt.Println(pin.PinID, "保存成功...", )
				}
				interval := rand.Intn(600) + 800
				time.Sleep(time.Millisecond * time.Duration(interval))
			} else {
				fmt.Println(pin.PinID, "该采集待公开...")
			}
		}
		if count >= board.PinCount {
			break
		}

		board, err = getNextPage(board.BoardID, lastPinID)
		if err != nil {
			log.Fatal(err)
		}

		if len(board.Pins) == 0 {
			break
		}

		interval := rand.Intn(600) + 800
		time.Sleep(time.Millisecond * time.Duration(interval))
	}
}

func getIndexPage(url string) (*model.Board, error) {
	bodyByte, err := httpGetPage(url)

	if err != nil {
		return nil, err
	}
	expr := "(app.page\\[\"board\"\\] = )(.*)(;)"
	reg, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	text := string(bodyByte)
	if reg.MatchString(text) {
		text = reg.FindString(text)
		text = strings.ReplaceAll(text, "app.page[\"board\"] = ", "")
		text = text[:len(text)-1]

		board := &model.Board{}
		if err := json.Unmarshal([]byte(text), board); err != nil {
			return nil, err
		}
		return board, nil
	}
	return nil, errors.New("No Match")
}

func getNextPage(boardID, lastPinID int) (*model.Board, error) {
	url := fmt.Sprintf("https://huaban.com/boards/%v?%v&max=%v&limit=20&wfl=1", boardID, RandString(), lastPinID)
	referer := fmt.Sprintf("https://huaban.com/boards/%v", boardID)
	jsonByte, err := httpGetJson(url, referer)
	if err != nil {
		return nil, err
	}

	page := &model.PageJson{}
	if err := json.Unmarshal(jsonByte, page); err != nil {
		return nil, err
	}
	return page.Board, nil
}

func httpGetPage(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	request.Header.Add("Accept-Encoding", "gzip, deflate, br")
	request.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	request.Header.Add("Connection", "keep-alive")
	request.Header.Add("Host", "huaban.com")
	request.Header.Add("Cache-Control", "no-cache")
	request.Header.Add("Sec-Fetch-Mode", "navigate")
	request.Header.Add("Sec-Fetch-Site", "none")
	request.Header.Add("Sec-Fetch-User", "?1")
	request.Header.Add("Upgrade-Insecure-Requests", "1")
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36 Edg/79.0.309.68")
	client := http.Client{}

	resp, err := client.Do(request)
	if resp.StatusCode != 200 {
		return nil, errors.New("Status Code Not 200")
	}
	defer resp.Body.Close()

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func httpGetJson(url, referer string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Accept-Encoding", "gzip, deflate, br")
	request.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	request.Header.Add("Connection", "keep-alive")
	request.Header.Add("Host", "huaban.com")
	request.Header.Add("Referer", referer)
	request.Header.Add("Sec-Fetch-Mode", "cors")
	request.Header.Add("Sec-Fetch-Site", "same-origin")
	request.Header.Add("X-Request", "JSON")
	request.Header.Add("X-Requested-With", "XMLHttpRequest")
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36 Edg/79.0.309.68")
	client := http.Client{}

	resp, err := client.Do(request)
	if resp.StatusCode != 200 {
		return nil, errors.New("Status Code Not 200")
	}
	defer resp.Body.Close()

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func downloadImage(dirName string, pin *model.Pin) error {
	fileType := "png"
	if strings.Contains(pin.File.Type, "image/jpeg") {
		fileType = "jpg"
	} else if strings.Contains(pin.File.Type, "image/gif") {
		fileType = "gif"
	}

	filePath := fmt.Sprintf("./%v/%v.%v", dirName, pin.PinID, fileType)

	_, err := os.Stat(filePath)
	if err == nil {
		return os.ErrExist
	}

	url := fmt.Sprintf("https://%v.huabanimg.com/%v", pin.File.Bucket, pin.File.Key)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	request.Header.Add("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")
	request.Header.Add("Accept-Encoding", "gzip, deflate, br")
	request.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	request.Header.Add("Connection", "keep-alive")
	request.Header.Add("Host", "huaban.com")
	request.Header.Add("Referer", fmt.Sprintf("https://huaban.com/pins/%v/", pin.PinID))
	request.Header.Add("Sec-Fetch-Mode", "no-cors")
	request.Header.Add("Sec-Fetch-Site", "cross-site")
	request.Header.Add("X-Request", "JSON")
	request.Header.Add("X-Requested-With", "XMLHttpRequest")
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36 Edg/79.0.309.68")
	client := http.Client{}

	resp, err := client.Do(request)
	if resp.StatusCode != 200 {
		return errors.New("Status Code Not 200")
	}
	defer resp.Body.Close()

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 获得文件的writer对象
	writer := bufio.NewWriter(file)

	_, err = io.Copy(writer, reader)
	if err != nil {
		return err
	}
	return nil
}

// 将rune数组用字符串常量替换
const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandString 生成随机字符串
func RandString() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
