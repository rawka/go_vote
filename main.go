package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/codegangsta/martini"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/render"
)

type qst struct {
	Question string   `json:"question"`
	Answers  []string `json:"answers"`
}

type question struct {
	id       int
	question string
	active   int
}

type answer struct {
	id          int
	answer      string
	question_id int
	vote        int
}

type qst_ans struct {
	Question string
	Answer   []string
}

var upgrader = websocket.Upgrader{}

func indexHandler(rnd render.Render) {
	rnd.HTML(200, "index", nil)
}

func clientHandler(rnd render.Render) {
	rnd.HTML(200, "client", nil)
}

func monitorHandler(rnd render.Render) {
	rnd.HTML(200, "monitor", nil)
}

func echo(w http.ResponseWriter, r *http.Request) {
	var id int

	connStr := "user=postgres password=123456 dbname=go_vote sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)

		quest := qst{}
		json.Unmarshal([]byte(message), &quest)

		db.QueryRow("insert into questions (question) values ($1) returning id", quest.Question).Scan(&id)

		for _, value := range quest.Answers {
			db.QueryRow("insert into answers (answer, question_id) values ($1, $2)", value, id)
		}

		//ok := "OK"
		//log.Printf(ok)

		//err = c.WriteMessage(mt, []byte(ok))
		//if err != nil {
		//	panic(err.Error())
		//}

		ans, err := db.Query("select * from answers")
		if err != nil {
			panic(err.Error())
		}
		answers := []answer{}

		for ans.Next() {
			a := answer{}
			err := ans.Scan(&a.id, &a.answer, &a.question_id, &a.vote)
			if err != nil {
				panic(err.Error())
			}
			answers = append(answers, a)
		}

		qst, err := db.Query("select * from questions")
		if err != nil {
			panic(err.Error())
		}
		questions := []question{}

		for qst.Next() {
			q := question{}
			err := qst.Scan(&q.id, &q.question, &q.active)
			if err != nil {
				panic(err.Error())
			}
			questions = append(questions, q)
		}
		result := []qst_ans{}
		for _, q := range questions {
			res := qst_ans{}
			res.Question = q.question
			for _, a := range answers {
				if a.question_id == q.id {
					res.Answer = append(res.Answer, a.answer)
				}
			}
			result = append(result, res)
		}
		res_json, err := json.Marshal(result)
		log.Println(string(res_json))
		if err != nil {
			panic(err.Error())
		}
		err = c.WriteMessage(mt, res_json)
		if err != nil {
			panic(err.Error())
		}
	}
}

func main() {

	m := martini.Classic()

	m.Use(render.Renderer(render.Options{
		Directory:  "templates",                // Specify what path to load the templates from.
		Extensions: []string{".tmpl", ".html"}, // Specify extensions to load for templates.
		Charset:    "UTF-8",                    // Sets encoding for json and html content-types. Default is "UTF-8".
		IndentJSON: true,                       // Output human readable JSON
	}))

	staticOptions := martini.StaticOptions{Prefix: "assets"}
	m.Use(martini.Static("assets", staticOptions))

	m.Get("/", indexHandler)
	m.Get("/client", clientHandler)
	m.Get("/monitor", monitorHandler)
	m.Get("/echo", echo)

	m.Run()
}
