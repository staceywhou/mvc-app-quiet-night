package main

import (
	"html/template"
)

var tmpl = make(map[string]*template.Template)

func init() {
	m := template.Must
	p := template.ParseFiles
	tmpl["index"] = m(p("templates/index.gohtml", "templates/layout.gohtml"))
	tmpl["about"] = m(p("templates/about.gohtml", "templates/layout.gohtml"))
	tmpl["details"] = m(p("templates/details.gohtml", "templates/layout.gohtml"))
	tmpl["visitor"] = m(p("templates/visitor.gohtml", "templates/layout.gohtml"))
	tmpl["create"] = m(p("templates/create.gohtml", "templates/layout.gohtml"))
}
