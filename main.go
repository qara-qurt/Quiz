package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	TOKEN   = "6040490988:AAFdeU7w_369vpt7iuwyfN4tqhK8wnUcHlM"
	CHAT_ID = "958729521"
)

func main() {
	data, timeLimit := getDataFromCSV()
	//Welcome text
	fmt.Print("Welcome to Quiz game !!! \n" +
		"Rules : \n " +
		"1.Given a time default 30 \n " +
		"2.Answer to question\n" +
		"What's your name : ")

	//Get name
	var name string
	fmt.Scan(&name)

	//Create a file to save user result
	fileUserState, err := os.Create(fmt.Sprintf("result/%s.csv", name))
	if err != nil {
		exit("Can't save result")
	}

	//Parse data
	questions := parseQuestionAnswer(data)

	//Make timer
	timer := time.NewTimer(time.Duration(*timeLimit) * time.Second)
	SendMessage(fmt.Sprintf("User %s start a QUIZ", name))

	var correct int
	for i, question := range questions {
		fmt.Printf("%d. Question : %s = ", i+1, question.q)
		answerCh := make(chan string)
		go func() {
			var answer string
			fmt.Scan(&answer)
			answerCh <- answer
		}()

		select {
		case answer := <-answerCh:
			isCorrect := answer == question.a
			if isCorrect {
				correct++
			}

			//Write result of question to file user result
			res := fmt.Sprintf("Question : %s , yourAnswer : %s , correctAnswer : %s , isCorrect %t\n",
				question.q, answer, question.a, isCorrect)
			fileUserState.WriteString(res)
			SendMessage(res)

		case <-timer.C:
			msg := fmt.Sprintf("\nTime's up your score - %d \n", correct)
			fmt.Printf(msg)
			SendMessage(msg)
			os.Exit(1)
		}
	}

	fmt.Printf("Your score - %d !", correct)
}

func getDataFromCSV() ([][]string, *int) {
	//flags file and timer
	filename := flag.String("csv", "problem1.csv", "a csv file in the format question&answer")
	timeLimit := flag.Int("timeLimit", 30, "the limit for the quiz second ")
	flag.Parse()

	//read data from .csv file
	file, err := os.Open(fmt.Sprintf("problems/%+v", *filename))
	if err != nil {
		exit(fmt.Sprintf("Failed to open the CSV file: %s\n", *filename))
	}
	defer file.Close()
	csvReader := csv.NewReader(file)
	data, err := csvReader.ReadAll()
	if err != nil {
		exit("Failed to parse the provided CSV file.")
	}

	return data, timeLimit
}

func parseQuestionAnswer(data [][]string) []QuestionAnswer {
	QnA := make([]QuestionAnswer, len(data))
	for i, item := range data {
		QnA[i] = QuestionAnswer{
			q: strings.TrimSpace(item[0]),
			a: strings.TrimSpace(item[1]),
		}
	}
	return QnA
}

type QuestionAnswer struct {
	q string
	a string
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func getUrl() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s", TOKEN)
}

func SendMessage(text string) error {
	var err error
	var response *http.Response

	url := fmt.Sprintf("%s/sendMessage", getUrl())

	body, _ := json.Marshal(map[string]string{
		"chat_id": CHAT_ID,
		"text":    text,
	})

	response, err = http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	fmt.Println(response)
	defer response.Body.Close()

	body, err = io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return nil
}
