package main

import (
	"bufio"
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"net/url"

	"github.com/tidwall/gjson"
)

var templateUri string
var inputReader *bufio.Reader

type Search struct {
	text   string
	offset int
	limit  int64
}

type Song struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

/**
 * httpGet method is for http get requesting
 */
func httpGet(uri string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	}
	//fmt.Println(string(body))
	return string(body)
}

func main() {
	fmt.Println("欢迎使用不负责任音乐下载器！")
	templateUri = "https://wy.azurewebsites.net/search?keywords={search}&offset={count}&limit=10"
	fmt.Print("请输入下载路径: ")
	path := ""
	fmt.Scanln(&path)
	fmt.Print("请输入要查找的歌曲名: ")
	text := ""
	fmt.Scanln(&text)
	search := &Search{text, 0, 10}
	t := strings.Replace(templateUri, "{search}", url.QueryEscape(search.text), -1)
	httpUrl := strings.Replace(t, "{count}", strconv.Itoa(search.offset), -1)
	result := httpGet(httpUrl)

	code := gjson.Get(result, "code").Int()
	fmt.Println(gjson.Get(result, "data"))
	if code != 200 {
		fmt.Println("网络错误，请重试！")
	} else {
		songCount := gjson.Get(result, "result.songCount").Int()
		pageCount := songCount/search.limit + 1
		fmt.Println("歌曲数量:" + strconv.Itoa(int(songCount)))
		for j := 1; j <= int(pageCount); j++ {
			fmt.Println("开始下载第" + strconv.Itoa(j) + "页")
			if j > 1 {
				t = strings.Replace(templateUri, "{search}", url.QueryEscape(search.text), -1)
				httpUrl = strings.Replace(t, "{count}", strconv.Itoa((j-1)*int(search.limit)), -1)
				fmt.Println(httpUrl)
				result = httpGet(httpUrl)
			}

			songsResult := gjson.Get(result, "result.songs").Array()
			songsList := list.New()
			for _, item := range songsResult {
				bytes := []byte(item.Raw)
				var s Song
				json.Unmarshal(bytes, &s)
				fmt.Println(s)
				songsList.PushBack(s)
			}
			i := 0

			for e := songsList.Front(); e != nil; e = e.Next() {
				i++
				s := e.Value.(Song)

				surl := "https://wy.azurewebsites.net/song/url?id="

				result := httpGet(surl + strconv.Itoa(int(s.Id)))
				realUrl := gjson.Get(result, "data.0.url").String()
				fmt.Println("开始下载第" + strconv.Itoa(i) + "个-" + s.Name)
				DownloadFile(path+"/"+s.Name+"-"+strconv.Itoa(int(s.Id))+".mp3", realUrl)

			}

		}

	}

}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
