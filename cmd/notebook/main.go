package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/acarl005/stripansi"
	bt "github.com/pratikdeoghare/brashtag"
	lit "github.com/pratikdeoghare/brashtag/cmd/lit/extractor"
	"golang.org/x/net/websocket"
)

func main() {

	filename := flag.String("f", "", "name of the frontend file")
	connStr := flag.String("c", "", "jupyter notebook connection string")
	flag.Parse()

	ws := setupConn(*filename, *connStr)

	respCh := make(chan []byte, 100)

	go func() {
		for {

			ws.SetReadDeadline(time.Now().Add(5 * time.Second))
			var data []byte
			err := websocket.Message.Receive(ws, &data)
			if err != nil {
				if errors.Is(err, io.EOF) {
					panic(1)
				}
				if errors.Is(err, os.ErrDeadlineExceeded) {
					slog.Debug("timed out", err)
					continue
				}
				slog.Error("error receiveing message", "error", err)
			} else {
				// maybe this is the root cause of out of order messages.
				// go func() {
				respCh <- data
				// }()
			}

		}
	}()

	buf := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\033[H\033[2J")
		fmt.Println("> ")
		_, err := buf.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			} else {
				slog.Error("failed to read [Enter]", err)
			}
		} else {
			err := processFile(*filename, ws, respCh)
			if err != nil {
				slog.Error("error processing file", err)
			}
		}
	}
}

func setupConn(filename, connStr string) *websocket.Conn {

	uu, err := url.Parse(connStr)
	if err != nil {
		panic(err)
	}

	hostport := fmt.Sprintf("%s:%s", uu.Hostname(), uu.Port())
	apiContents := fmt.Sprintf("http://%s/api/contents", hostport)
	apiSessions := fmt.Sprintf("http://%s/api/sessions", hostport)

	// apiContents := "http://localhost:8888/api/contents"
	// apiKernelSpecs := "http://localhost:8888/api/kernelspecs"
	// apiSessions := "http://localhost:8888/api/sessions"

	u, err := url.Parse(connStr)
	if err != nil {
		slog.Error("failed to parse jupyter connection string", "connStr", connStr, "error", err)
	}

	token := u.Query().Get("token")

	doReq := func(req *http.Request) (*http.Response, error) {
		slog.Info("sending request")
		req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
		return http.DefaultClient.Do(req)
	}

	interact := func(method, path string, body []byte) []byte {
		req, err := http.NewRequest(method, path, bytes.NewReader(body))

		if err != nil {
			slog.Error("failed to create request", "error", err, "path", path)
			return nil
		}

		resp, err := doReq(req)
		if err != nil {
			slog.Error("request returned error", "error", err)
			return nil
		}

		respData, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("failed to read response", "error", err)
			return nil
		}

		return respData
	}

	data := interact("POST", apiContents, MustJSONBytes(
		map[string]string{
			"type": "notebook",
		},
	))

	slog.Info(string(data))

	interact("PATCH", apiContents+"/Untitled.ipynb", MustJSONBytes(
		map[string]string{
			"path": fmt.Sprintf("%s.ipynb", filename),
		},
	))

	slog.Info(string(data))

	data = interact("POST", apiSessions, MustJSONBytes(
		map[string]any{
			"kernel": dict{
				"name": "python3",
			},
			"type": "notebook",
			"name": fmt.Sprintf("%s.ipynb", filename),
			"path": fmt.Sprintf("%s.ipynb", filename),
		},
	))

	type SessionResponse struct {
		ID     string `json:"id"`
		Kernel dict   `json:"kernel"`
	}

	sessionResp := new(SessionResponse)
	err = json.Unmarshal(data, &sessionResp)
	if err != nil {
		slog.Error("failed to get session", err)
	}

	sessionID, kernelID := sessionResp.ID, sessionResp.Kernel["id"]

	wsAPI := fmt.Sprintf("ws://%s/api/kernels/%s/channels?session_id=%s", hostport, kernelID, sessionID)

	fmt.Println(string(data))
	fmt.Println(sessionResp.Kernel["id"])

	wsCfg := &websocket.Config{
		Location: mustURLParse(wsAPI),
		Origin:   mustURLParse("http://localhost/"),
		Header:   make(map[string][]string),
		Version:  websocket.ProtocolVersionHybi,
	}

	wsCfg.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	ws, err := websocket.DialConfig(wsCfg)
	if err != nil {
		slog.Error("failed to ws connect", err)
	}

	fmt.Println(ws)

	return ws

}

func processFile(filename string, ws *websocket.Conn, respCh chan []byte) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	fmt.Println(strings.Repeat("-", 20))
	fmt.Println(string(data))
	fmt.Println(strings.Repeat("-", 20))

	tree, err := bt.Parse(strings.TrimSpace(string(data)))
	if err != nil {
		return err
	}

	fmt.Println(strings.Repeat("-", 20))
	fmt.Println(tree.String())
	fmt.Println(strings.Repeat("-", 20))

	extr := lit.NewExtractor(tree)
	fmt.Println(extr)
	cells := f(tree, extr)

	fmt.Println(cells)

	for _, cell := range cells {

		fmt.Printf("SRC %+v\n\n", cell)

		reqID := fmt.Sprintf("exec-req-%d", rand.Int())
		codeMsg := makeCodeMsg(cell.SourceCode, reqID)

		ws.SetReadDeadline(time.Now().Add(time.Second * 10))

		fmt.Println(codeMsg)

		err = websocket.JSON.Send(ws, codeMsg)
		if err != nil {
			slog.Error("Code send problem", err)
		} else {
			log.Print("Code sent!")
		}

		fmt.Println("START ", reqID)
		_, i := fixIndent(cell.SourceCode)

		s := ""
		lastSave := time.Now()

		for {
			data := <-respCh

			x := make(dict)
			err = json.Unmarshal(data, &x)
			if err != nil {
				slog.Error("failed to unmarshall", err)
				continue
			}

			parentID, ok := x["parent_header"].(map[string]any)["msg_id"].(string)

			// fmt.Println(string(data1))

			if !ok {
				data, _ := json.MarshalIndent(x, "", " ")
				slog.Error("no parentid", "msg", string(data))
				continue
			}

			if parentID == reqID {
				v, done := handleResponse(x)
				s += v

				nKids := len(cell.Node.Kids())
				for nKids > 0 {
					cell.Node.RemoveKid(0)
					nKids--
				}
				cell.Node.AddKids(bt.NewBlob("\n" + strings.Repeat(" ", i)))
				cell.Node.AddKids(bt.NewCode(
					strings.Repeat("`", min(80, 8+longestLine(stripansi.Strip(s)))),
					indent("\n"+stripansi.Strip(s)+"\n", i)))

				if done || time.Since(lastSave) > 5*time.Second {

					err = os.WriteFile(filename, []byte(tree.String()), 0644)
					if err != nil {
						slog.Error("failed to update file", err)
					}

				}

				if done {
					break
				}
			} else {
				fmt.Println(x)
			}
		}

		nKids := len(cell.Node.Kids())
		for nKids > 0 {
			cell.Node.RemoveKid(0)
			nKids--
		}
		cell.Node.AddKids(bt.NewBlob("\n" + strings.Repeat(" ", i)))
		cell.Node.AddKids(bt.NewCode(
			strings.Repeat("`", min(80, 8+longestLine(stripansi.Strip(s)))),
			indent("\n"+stripansi.Strip(s)+"\n", i)))

		err = os.WriteFile(filename, []byte(tree.String()), 0644)
		if err != nil {
			slog.Error("failed to update file", err)
		}

	}

	return nil
}

func fixIndent(s string) (string, int) {
	lines := strings.Split(s, "\n")
	if len(lines) < 2 {
		return s, 0
	}

	i := 0
	for i < len(lines[1]) && lines[1][i] == ' ' {
		i++
	}
	for j, line := range lines {
		if len(line) >= i {
			if strings.TrimSpace(line[:i]) == "" {
				lines[j] = line[i:]
			}
		}
	}
	return strings.Join(lines, "\n"), i
}

func indent(s string, i int) string {
	prefix := strings.Repeat(" ", i)
	lines := strings.Split(s, "\n")
	for j, line := range lines {
		lines[j] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func longestLine(s string) int {
	l := math.MinInt
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		l = max(l, len(line))
	}
	return l
}

func MustJSONBytes[T any](body T) []byte {
	data, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return data
}

type dict map[string]any

func mustURLParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

type Header struct {
	MessageID   string `json:"msg_id"`
	MessageType string `json:"msg_type"`
}

type Content2 struct {
	Silent bool   `json:"silent"`
	Code   string `json:"code"`
}

type CodeMsg struct {
	Channel      string         `json:"channel"`
	Content2     Content2       `json:"content"`
	Header       Header         `json:"header"`
	Metadata     map[string]any `json:"metadata"`
	ParentHeader map[string]any `json:"parent_header"`
}

func makeCodeMsg(code string, uuid string) CodeMsg {
	return CodeMsg{
		Channel: "shell",
		Content2: Content2{
			Silent: false,
			Code:   code,
		},
		Header: Header{
			MessageID:   uuid,
			MessageType: "execute_request",
		},
		Metadata:     make(map[string]any),
		ParentHeader: make(map[string]any),
	}
}

type Cell struct {
	Node       *bt.Bag
	SourceCode string
}

func f(tree bt.Node, extr lit.Extractor) []Cell {

	fmt.Println(tree.String())

	var cells []Cell
	switch x := tree.(type) {
	case bt.Blob:
	case bt.Code:
	case bt.Bag:
		var prevCode bt.Code
		for _, k := range x.Kids() {
			switch y := k.(type) {
			case bt.Bag:
				if y.Tag() == "out" {
					cells = append(cells, Cell{
						Node:       &y,
						SourceCode: extr.ResolveDeps(prevCode.Text()),
					})
				}
				cells = append(cells, f(y, extr)...)
			case bt.Code:
				prevCode = y

				fmt.Println(prevCode, "<---")
			}
		}
	}
	return cells
}

func handleResponse(x dict) (string, bool) {
	get := func(s string) (string, bool) {
		xs := strings.Split(s, ".")

		if len(xs) == 1 {
			val, ok := x[s]
			return fmt.Sprint(val), ok
		}

		var v map[string]any = x[xs[0]].(map[string]any)
		for _, k := range xs[1 : len(xs)-1] {
			v = v[k].(map[string]any)
		}
		val, ok := v[xs[len(xs)-1]]
		return fmt.Sprint(val), ok
	}

	switch x["msg_type"] {
	case "execute_request":
	case "status":
		v, _ := get("content.execution_state")
		if v == "idle" {
			return "", true
		}
	case "execute_input":
		fmt.Println(x)
	case "stream":
		v, _ := get("content.text")
		return v, false
	case "error":
		// We handle this in execute_reply case.
	case "execute_result":
		v, _ := get("content.data.text/plain")
		return v, false
	case "execute_reply":
		v, _ := get("content.status")
		if v == "error" {
			v, _ := get("content.traceback")
			return v, true
		}
		return "", true
	default:
		return "", false
	}
	return "", false
}
