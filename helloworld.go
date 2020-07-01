// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Sample helloworld is a basic App Engine flexible app.
package main

import (
	"fmt"
	"log"
	"net/http"
	"context"
	"os"
	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"html/template"
	"google.golang.org/api/iterator"
)

type getHandler struct {
	client *firestore.Client
}

type Document struct {
	User string
	Note string 
}

var tmplStr = `
<html><head>
<style>
*{
  margin:0;
  padding:0;
}
body{
  font-family:arial,sans-serif;
  font-size:100%;
  margin:3em;
  background:#666;
  color:#fff;
}
h2,p{
  font-size:100%;
  font-weight:normal;
}
ul,li{
  list-style:none;
}
ul{
  overflow:hidden;
  padding:3em;
}
ul li a{
  text-decoration:none;
  color:#000;
  background:#ffc;
  display:block;
  height:10em;
  width:10em;
  padding:1em;
  -moz-box-shadow:5px 5px 7px rgba(33,33,33,1);
  -webkit-box-shadow: 5px 5px 7px rgba(33,33,33,.7);
  box-shadow: 5px 5px 7px rgba(33,33,33,.7);
  -moz-transition:-moz-transform .15s linear;
  -o-transition:-o-transform .15s linear;
  -webkit-transition:-webkit-transform .15s linear;
}
ul li{
  margin:1em;
  float:left;
}
ul li h2{
  font-size:140%;
  font-weight:bold;
  padding-bottom:10px;
}
ul li p{
  font-family:"Reenie Beanie",arial,sans-serif;
  font-size:180%;
}
ul li a{
  -webkit-transform: rotate(-6deg);
  -o-transform: rotate(-6deg);
  -moz-transform:rotate(-6deg);
}
ul li:nth-child(even) a{
  -o-transform:rotate(4deg);
  -webkit-transform:rotate(4deg);
  -moz-transform:rotate(4deg);
  position:relative;
  top:5px;
  background:#cfc;
}
ul li:nth-child(3n) a{
  -o-transform:rotate(-3deg);
  -webkit-transform:rotate(-3deg);
  -moz-transform:rotate(-3deg);
  position:relative;
  top:-5px;
  background:#ccf;
}
ul li:nth-child(5n) a{
  -o-transform:rotate(5deg);
  -webkit-transform:rotate(5deg);
  -moz-transform:rotate(5deg);
  position:relative;
  top:-10px;
}
ul li a:hover,ul li a:focus{
  box-shadow:10px 10px 7px rgba(0,0,0,.7);
  -moz-box-shadow:10px 10px 7px rgba(0,0,0,.7);
  -webkit-box-shadow: 10px 10px 7px rgba(0,0,0,.7);
  -webkit-transform: scale(1.25);
  -moz-transform: scale(1.25);
  -o-transform: scale(1.25);
  position:relative;
  z-index:5;
}

.user {
  text-decoration:none;
  text-align: center;
  color:#000;
  background:#ffc;
  display:block;
  height:2em;
  width:10em;
  padding:0.5em;
  -moz-box-shadow:5px 5px 7px rgba(33,33,33,1);
  -webkit-box-shadow: 5px 5px 7px rgba(33,33,33,.7);
  box-shadow: 5px 5px 7px rgba(33,33,33,.7);
  -moz-transition:-moz-transform .15s linear;
  -o-transition:-o-transform .15s linear;
  -webkit-transition:-webkit-transform .15s linear;
}

.note {
  padding:0.5em;
  text-align: center;
  text-decoration:none;
  color:#000;
  background:#ffc;
  display:block;
  height:10em;
  width:10em;
  padding:1em;
  -moz-box-shadow:5px 5px 7px rgba(33,33,33,1);
  -webkit-box-shadow: 5px 5px 7px rgba(33,33,33,.7);
  box-shadow: 5px 5px 7px rgba(33,33,33,.7);
  -moz-transition:-moz-transform .15s linear;
  -o-transition:-o-transform .15s linear;
  -webkit-transition:-webkit-transform .15s linear;
}

ul li{
  margin:1em;
  float:left;
}
</style>
</head>

<body>
<ul>
{{range .}}
    <li>
    <a href="#">
        <h2>{{.User}}</h2>
        <p>{{.Note}}</p>
    </a>
    </li>
{{end}}
</ul>
  <form action="/" method="POST" novalidate>
    <textarea placeholder='user' name='user' class="user"></textarea>
    <textarea placeholder='note' name='note' class="note"></textarea>
    <input type="submit" value="Submit new note">
 </form>
</body>
</html>
`

var tmpl = template.Must(template.New("t").Parse(tmplStr))


func (h *getHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	ctx := context.Background()
	iter := h.client.Collection("Wall").Query.OrderBy("User", firestore.Asc).Documents(ctx)
	defer iter.Stop()
	var docs []*Document
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		d := &Document{}
		doc.DataTo(d)
		docs = append(docs, d)
	}
	tmpl.Execute(w, docs)
}

type putHandler struct {
	client *firestore.Client
}

func (h *putHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	note := r.FormValue("note")
	user := r.FormValue("user")
	fmt.Printf("Got form: %v, %v", note, user)
	ref := h.client.Collection("Wall").NewDoc()
	d := &Document{}
	d.User = user
	d.Note = note
	ctx := context.Background()
	if _, err := ref.Create(ctx, d); err != nil {
		fmt.Printf("Create: %v", err)
	}
}

func registerHandlers(h *getHandler) {
	r := mux.NewRouter()
	r.Methods("GET").Path("/").Handler(h)
	r.Methods("POST").Path("/").Handler(h)
	http.Handle("/", r)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT must be set")
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)
	}
	//bucketName := projectID + "_bucket"
	h := &getHandler{
		client,
	}
	registerHandlers(h)
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
