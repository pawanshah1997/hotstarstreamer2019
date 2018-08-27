//https://www.hotstar.com/tv/neeya-naana/s-80/hard-work-vs-smart-work/1100007278
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type cdn_token_struct struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

type get_metaData_struct struct {
	ErrorDescription string `json:"errorDescription"`
	Message          string `json:"message"`
	ResultCode       string `json:"resultCode"`
	ResultObj        struct {
		CheckCacheResult string `json:"checkCacheResult"`
		Height           string `json:"height"`
		Src              string `json:"src"`
		Width            string `json:"width"`
	} `json:"resultObj"`
	SystemTime int `json:"systemTime"`
}
type chunks struct {
	id  int
	url string
}

func worker(i int, jobs <-chan chunks, results chan<- int) {
	for j := range jobs {
		// fmt.Println("worker", i, "started  job", j.id)
		client := &http.Client{}
		req, _ := http.NewRequest("GET", j.url, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
		resp, err := client.Do(req)
		// fmt.Println(resp.StatusCode)
		isError(err)
		defer resp.Body.Close()
		reader, _ := ioutil.ReadAll(resp.Body)
		ioutil.WriteFile(strconv.Itoa(j.id), []byte(string(reader)), 0x777) // Write to the file i as a byte array
		resp.Body.Close()
		//fmt.Println("worker", i, "finished job", j.id)
		results <- j.id
	}
}

func OneGetCDNToken() string {
	var cdnurl = "https://www.hotstar.com/get_cdn_token.php"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", cdnurl, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	req.Header.Set("authority", "www.hotstar.com")
	req.Header.Set("referer", "https://www.hotstar.com/tv/neeya-naana/s-80/poets-vs-women/1100006996")
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	_cdntoken := new(cdn_token_struct)
	json.NewDecoder(resp.Body).Decode(_cdntoken)
	return _cdntoken.Token
}

func TwoGetMetaDataURL(cdnToken string, id string) string {
	var url = "https://secure-getcdn.hotstar.com/AVS/besc?hotstarauth=" + cdnToken + "&action=GetCDN&appVersion=5.0.40&asJson=Y&channel=TABLET&id=" + id + "&type=VOD"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	req.Header.Set("authority", "www.hotstar.com")
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	_metaDataStruct := new(get_metaData_struct)
	json.NewDecoder(resp.Body).Decode(_metaDataStruct)
	return _metaDataStruct.ResultObj.Src
}
func ThreeGetQualityMetaData(url string, quality string) string {
	client := &http.Client{}
	// println(url)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	req.Header.Set("authority", "www.hotstar.com")
	resp, _ := client.Do(req)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	metaData := strings.Split(string(bodyBytes), "\n")
	for i := 0; i < len(metaData); i++ {
		if strings.Index(metaData[i], quality) != -1 {
			return metaData[i+1]
		}
	}
	return ""
}

func FourGetVideoChunksMetaData(url string) []string {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	req.Header.Set("authority", "www.hotstar.com")
	resp, _ := client.Do(req)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	sl := strings.Split(string(bodyBytes), "\n")
	chunks := make([]string, 0)
	for i := 0; i < len(sl); i++ {
		if strings.Index(sl[i], "#") == -1 && len(sl[i]) != 0 {
			chunks = append(chunks, sl[i])
		}
	}
	return chunks
}

func Initiate() (string, string, int) {
	sl := strings.Split(os.Args[1], "/")
	workers := 2
	return sl[len(sl)-1], sl[len(sl)-4] + "_" + sl[len(sl)-3] + "_" + sl[len(sl)-2] + ".ts", workers
}

func mergeParts(name string, limit int) {
	os.Remove(name)
	f, err := os.OpenFile(strings.TrimSpace(name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	isError(err)
	defer f.Close()
	for i := 0; i < limit; i++ {
		var part = strconv.Itoa(i)
		content, err := ioutil.ReadFile(part) // just pass the file name
		isError(err)
		f.WriteString(string(content))
		os.Remove(part)
	}
}
func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}
	return (err != nil)
}
func getQuality() string {
	fmt.Println("Choose the Quality:")
	fmt.Println("1. 1080p")
	fmt.Println("2. 900p")
	fmt.Println("3. 720p")
	fmt.Println("4. 360p")
	fmt.Println("5. 240p")
	fmt.Println("6. 180p")
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter the options: ")
	text, _ := reader.ReadString('\n')
	switch strings.TrimSpace(text) {
	case "1":
		return "1920x1080"
	case "2":
		return "1600x900"
	case "3":
		return "1280x720"
	case "4":
		return "640x360"
	case "5":
		return "416x234"
	case "6":
		return "320x180"
	default:
		{
			fmt.Println("Choose the right option")
			os.Exit(1)
		}
	}
	return ""
}

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("pass the hotstar url")
		return
	}
	id, fileName, workers := Initiate()
	// println(id, fileName)
	cdnToken := OneGetCDNToken()
	QualityMetaDataUrl := TwoGetMetaDataURL(cdnToken, id)
	baseUrl := strings.Split(QualityMetaDataUrl, "master")[0]
	// println(baseUrl)
	Quality := getQuality()
	VideoChunksMetaDataUrl := ThreeGetQualityMetaData(QualityMetaDataUrl, Quality)
	videoChunksMetaDatas := FourGetVideoChunksMetaData(baseUrl + VideoChunksMetaDataUrl)
	fmt.Println("Count of Video Chunks: ", len(videoChunksMetaDatas))
	jobs := make(chan chunks, len(videoChunksMetaDatas)+1)
	results := make(chan int, len(videoChunksMetaDatas)+1)
	start := time.Now()
	fmt.Println("Workers: ", workers)
	for i := 0; i < workers; i++ {
		go worker(i, jobs, results)
	}
	for i := 0; i < len(videoChunksMetaDatas); i++ {
		jobs <- chunks{id: i, url: baseUrl + videoChunksMetaDatas[i]}
	}
	close(jobs)
	// Finally we collect all the results of the work.
	for a := 0; a < len(videoChunksMetaDatas); a++ {
		<-results
	}
	mergeParts(fileName, len(videoChunksMetaDatas))
	fmt.Println("Downloaded ", fileName)
	fmt.Println("Time Taken: ", time.Since(start).String())
}
