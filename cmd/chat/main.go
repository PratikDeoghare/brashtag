package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
	bt "github.com/pratikdeoghare/brashtag"
)

func main() {

	filename := flag.String("f", "", "name of the frontend file")
	dumpDir := flag.String("dumpDir", "", "location of directory where chats will be dumped")
	flag.Parse()

	_ = dumpDir

	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	var chatHistory []api.Message

	mustChatHistoryBytes := func() []byte {
		data, err := json.MarshalIndent(chatHistory, "", " ")
		if err != nil {
			panic(err)
		}
		return data
	}

	buf := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\033[H\033[2J")
		fmt.Println("> ")
		command, err := buf.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			} else {
				slog.Error("failed to read [Enter]", err)
			}
		} else {

			switch strings.TrimSpace(string(command)) {

			case "new":
				// dump context to file
				err = os.WriteFile(
					path.Join(*dumpDir, fmt.Sprint(time.Now().UnixNano())),
					mustChatHistoryBytes(), 0644)
				if err != nil {
					slog.Error("failed to update file", err)
				}

				// clear it
				chatHistory = nil

			// case "save":
			// will implement this when needed.

			default:
				hs, err := processFile(*filename, client, chatHistory)
				if err != nil {
					slog.Error("error processing file", err)
				} else {
					chatHistory = hs
				}

				fmt.Println("len of history:", len(chatHistory))
			}
		}
	}
}

func getChatBag(tree bt.Node) *bt.Bag {
	switch x := tree.(type) {
	case bt.Bag:
		if x.Tag() == "+chat/llm" {
			return &x
		}
		for _, k := range x.Kids() {
			y := getChatBag(k)
			if y != nil {
				return y
			}
		}
	}

	return nil
}

func processFile(filename string, client *api.Client, chatHistory []api.Message) ([]api.Message, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	tree, err := bt.Parse(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, err
	}

	chatBag := getChatBag(tree)
	if chatBag == nil {
		return nil, fmt.Errorf("no chat prompt")
	}

	kids := chatBag.Kids()
	i := len(kids) - 1
loop:
	for i > 0 {
		switch x := kids[i].(type) {
		case bt.Bag:
			if x.Tag() == "llmout" {
				i++
				break loop
			}
		}
		i--
	}

	fmt.Println("i = ", i, "n = ", len(kids))
	prompt := ""
	for _, k := range kids[i:] {
		prompt += k.String()
	}

	msg := api.Message{
		Role:    "user",
		Content: prompt,
	}

	ii := getIndent(prompt)
	promptLines := strings.Split(prompt, "\n")
	lastLine := promptLines[len(promptLines)-1]

	chatHistory = append(chatHistory, msg)

	for _, entry := range chatHistory {
		s := entry.Content
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "\n", "")

		fmt.Println(entry.Role, len(s), s[:min(10, len(s))])
	}

	req := &api.ChatRequest{
		Model:    "codellama",
		Messages: chatHistory,
		Stream:   new(bool),
	}

	ctx := context.Background()
	llmoutBag := bt.NewBag("llmout")
	respFunc := func(resp api.ChatResponse) error {
		llmoutBag.AddKids(
			bt.NewCode(
				strings.Repeat("`", 10),
				//"\n"+resp.Message.Content+"\n",

				indent("\n"+resp.Message.Content+"\n", ii+5),
			),
		)

		chatBag.AddKids(
			bt.NewBlob("\n"),
			bt.NewBlob(indent("", ii+5)),
			llmoutBag,
			bt.NewBlob("\n"),
			bt.NewBlob(lastLine),
		)

		chatHistory = append(chatHistory, resp.Message)
		return nil
	}

	err = client.Chat(ctx, req, respFunc)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(filename, []byte(tree.String()), 0644)
	if err != nil {
		return nil, err
	}

	return chatHistory, nil
}

func getIndent(s string) int {
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l != "" {
			j := 0
			for line[j] != l[0] {
				j++
			}
			return j
		}
	}
	return -1
}

func indent(s string, i int) string {
	prefix := strings.Repeat(" ", i)
	lines := strings.Split(s, "\n")
	for j, line := range lines {
		lines[j] = prefix + line
	}
	return strings.Join(lines, "\n")
}
