//https://www.hotstar.com/movies/baahubali2-the-conclusion/1770016247/watch
package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
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
	Body struct {
		Results struct {
			Item struct {
				PlaybackURL string `json:"playbackUrl"`
			} `json:"item"`
			ResponseType string `json:"responseType"`
		} `json:"results"`
	} `json:"body"`
	StatusCode      string `json:"statusCode"`
	StatusCodeValue int    `json:"statusCodeValue"`
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

//https://api.hotstar.com/h/v1/play?contentId
func TwoGetMetaDataURL(cdnToken string, id string, hotstarUrl string) string {
	var url = "http://localhost:3000/?url=" + hotstarUrl
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	req.Header.Set("authority", "www.hotstar.com")
	req.Header.Set("x-platform-code", "TABLET")
	req.Header.Set("x-country-code", "IN")
	req.Header.Set("hotstarauth", "st=1540230733~exp=1540236733~acl=/*~hmac=9ed52df2a7ff3dae36b16e6deb73f25a5ea0f85a1cdc8be7c53d2eca4b3b93df")
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	_metaDataStruct := new(get_metaData_struct)
	fmt.Println(resp.Body)
	json.NewDecoder(resp.Body).Decode(_metaDataStruct)
	return _metaDataStruct.Body.Results.Item.PlaybackURL
}
func ThreeGetQualityMetaData(url string, quality string) string {
	client := &http.Client{}
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

func FourGetVideoChunksMetaData(url string) ([]string, []byte) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	req.Header.Set("authority", "www.hotstar.com")
	resp, _ := client.Do(req)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	sl := strings.Split(string(bodyBytes), "\n")
	chunks := make([]string, 0)
	decryptUrl := make([]string, 0)
	for i := 0; i < len(sl); i++ {
		if i == 3 {
			decryptUrl = strings.Split(sl[i], "\"")
		}
		if strings.Index(sl[i], "#") == -1 && len(sl[i]) != 0 {
			chunks = append(chunks, sl[i])
		}
	}
	reqa, _ := http.NewRequest("GET", decryptUrl[1], nil)
	reqa.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	reqa.Header.Set("authority", "www.hotstar.com")
	respa, _ := client.Do(reqa)
	keybodyBytes, _ := ioutil.ReadAll(respa.Body)
	return chunks, keybodyBytes
}

func Initiate() (string, string, int, string) {
	url := os.Args[1]
	sl := strings.Split(os.Args[1], "/")
	workers := 2
	return sl[len(sl)-1], sl[len(sl)-4] + "_" + sl[len(sl)-3] + "_" + sl[len(sl)-2] + ".ts", workers, url
}
func AESDecrypt(crypt []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("key error1", err)
	}
	if len(crypt) == 0 {
		fmt.Println("plain content empty")
	}
	ecb := cipher.NewCBCDecrypter(block, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	decrypted := make([]byte, len(crypt))
	ecb.CryptBlocks(decrypted, crypt)

	return PKCS5Trimming(decrypted)
}
func PKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}

func mergeParts(name string, limit int, key []byte) {
	os.Remove(name)
	f, err := os.OpenFile(strings.TrimSpace(name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	isError(err)
	defer f.Close()
	for i := 0; i < limit; i++ {
		var part = strconv.Itoa(i)
		content, err := ioutil.ReadFile(part) // just pass the file name
		isError(err)
		f.WriteString(string(AESDecrypt(content, key)))
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
	id, fileName, workers, url := Initiate()
	// println(id, fileName)
	cdnToken := ""
	QualityMetaDataUrl := TwoGetMetaDataURL(cdnToken, id, url)
	Quality := getQuality()
	VideoChunksMetaDataUrl := ThreeGetQualityMetaData(QualityMetaDataUrl, Quality)
	videoChunksMetaDatas, key := FourGetVideoChunksMetaData(VideoChunksMetaDataUrl)
	fmt.Println(key)
	fmt.Println("Count of Video Chunks: ", len(videoChunksMetaDatas))
	jobs := make(chan chunks, len(videoChunksMetaDatas)+1)
	results := make(chan int, len(videoChunksMetaDatas)+1)
	start := time.Now()
	fmt.Println("Workers: ", workers)
	for i := 0; i < workers; i++ {
		go worker(i, jobs, results)
	}
	for i := 0; i < len(videoChunksMetaDatas); i++ {
		jobs <- chunks{id: i, url: videoChunksMetaDatas[i]}
	}
	close(jobs)
	// Finally we collect all the results of the work.
	for a := 0; a < len(videoChunksMetaDatas); a++ {
		<-results
	}
	mergeParts(fileName, len(videoChunksMetaDatas), key)
	fmt.Println("Downloaded ", fileName)
	fmt.Println("Time Taken: ", time.Since(start).String())
}
