package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) <= 1 {
		log.Println("Usage: go run distorted.go <dir_path>")
		return
	}
	items, _ := ioutil.ReadDir(os.Args[1])
	for _, item := range items {

		if !strings.HasSuffix(item.Name(), "wav") {
			continue
		}
		duration, _ := strconv.Atoi(os.Args[2])
		if item.ModTime().Before(time.Now().Add(time.Duration(-duration) * time.Second)) {
			continue
		}
		log.Println("Analysing File: " + item.Name())
		amp, err := getMeanAmplitude(os.Args[1] + "/" + item.Name())
		if err != nil {
			fmt.Println(err)
		}
		absAmp := math.Abs(amp)
		// log.Printf("Mean Amplitude: %f", x)
		if absAmp > 0.001 {
			log.Printf("Distorted File: %s | Mean Amplitude: %f", item.Name(), absAmp)
		}
	}
	return
}

func getMeanAmplitude(audioFile string) (float64, error) {
	cmd := "/usr/local/bin/sox " + audioFile + " -n stat 2>&1 | grep \"Mean    amplitude\" | awk '{print $3}'"
	// cmd := "/usr/local/bin/sox " + audioFile + " -n stat"
	// log.Println(cmd)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return 0, err
	}
	// log.Println("Command Output: ", string(out))
	mAmp, err := strconv.ParseFloat(strings.TrimSuffix(string(out), "\n"), 64)
	if err != nil {
		return 0, err
	}
	return mAmp, nil
}
