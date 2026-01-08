package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/system-go/config"
)

func main() {
	/*file, _ := os.Open(".env")
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Bytes()
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		eq := bytes.IndexByte(line, '=')
		if eq == -1 {
			continue
		}

		env_key := line[:eq]
		env_value := line[eq+1:]
		fmt.Println("Key: ", string(env_key), "Value: ", string(env_value))
		os.Setenv(string(env_key), string(env_value))
		env := os.Getenv(string(env_key))
		fmt.Println(env)*/

	root, _ := os.Getwd()
	file, err := os.Open("C__Windows_system32_cmd.exe  2025-02-26 00-45-14.mp4")
	if err != nil {
		fmt.Println("Error opening file")
	}
	defer file.Close()

	dst, err := os.Create(filepath.Join(root, "movie.mp4"))
	if err != nil {
		panic(err)
	}
	defer dst.Close()

	buffer := make([]byte, 10*1024*1024)
	pathNumber := 1
	var totalSize int
	uploadId := config.GetUploadId(file.Name())
	startTime := time.Now()
	var uploadEtag []*s3.CompletedPart

	for {
		r, err := file.Read(buffer)
		if r > 0 {
			totalSize += r
			/*_, writeErr := dst.Write(buffer[:r])
			if writeErr != nil {
				panic(writeErr)
			}*/
			fmt.Println("Part ID: ", pathNumber, "Part size: ", r/(1024*1024), "mb")

			// Upload to S3 bucket
			etag, err := config.UploadPartToS3(file.Name(), *uploadId, pathNumber, buffer[:r])
			if err != nil {
				fmt.Println("Error getting part url")
			}
			fmt.Println("Part Upload: ", etag)
			uploadEtag = append(uploadEtag, &s3.CompletedPart{
				ETag:       etag.ETag,
				PartNumber: aws.Int64(int64(pathNumber)),
			})

			pathNumber++
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err.Error())
		}

	}

	url, err := config.UploadFinish(file.Name(), uploadEtag, *uploadId)
	fmt.Println(url)
	fmt.Println("Total file size: ", totalSize/(1024*1024), "MB")
	fmt.Println("Upload took:", time.Since(startTime).Minutes(), "minutes")

}
