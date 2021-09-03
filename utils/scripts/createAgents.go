package main

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	// Open the file
	csvfile, err := os.Open("agents.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	// Parse the file
	r := csv.NewReader(csvfile)

	var allAgents []string
	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		agentName := record[0]
		if inSlice(allAgents, agentName) {
			fmt.Printf("#Agent already exists. [%s]\n", agentName)
			continue
		}
		allAgents = append(allAgents, agentName)
		md5Hash := createMD5Hash(agentName, "12321")
		query := "INSERT INTO sip_buddies (`name`, `defaultuser`, `md5secret`, `context`, `host`, `nat`, `qualify`, `type`) VALUES ('" +
			agentName +
			"','" +
			agentName +
			"','" +
			md5Hash +
			"', 'incoming', 'dynamic', 'yes', 'yes', 'friend');"
		fmt.Println(query)
	}
}

func inSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func createMD5Hash(user, password string) string {
	str := user + ":asterisk:" + password
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}
