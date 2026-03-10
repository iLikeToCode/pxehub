package httpserver

import (
	"html/template"
	"net/http"
	"os"
	"pxehub/internal/db"
	"pxehub/ui"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (h *HttpServer) UI(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "text/html")

	path := strings.Trim(r.URL.Path, "/")

	var name = ""

	if path != "login" {
		if username, ok := getSessionUsername(r); ok {
			name = username
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}

	switch path {
	case "":
		files := []string{
			"base.html",
			"index.html",
		}

		tmpl, err := template.ParseFS(ui.Content, files...)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		graphData1, graphData2, graphDates, err := db.GetRequestGraphData(time.Now(), h.Database)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		totalRequests, err := db.GetTotalRequestCount(h.Database)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		totalHosts, err := db.GetTotalHostCount(h.Database)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		activeTasks, err := db.GetActiveTaskCount(h.Database)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		caser := cases.Title(language.English)
		data := map[string]any{
			"Title":                 caser.String("Home"),
			"Name":                  name,
			"Path":                  r.URL.Path,
			"RegisteredGraphData":   graphData1,
			"UnregisteredGraphData": graphData2,
			"GraphDates":            graphDates,
			"Month":                 time.Now().Month(),
			"TotalRequests":         totalRequests,
			"TotalHosts":            totalHosts,
			"ActiveTasks":           activeTasks,
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case "hosts", "hosts/new":
		files := []string{
			"base.html",
			"hosts.html",
		}

		tmpl, err := template.ParseFS(ui.Content, files...)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hostsHtml, err := db.GetHostsAsHTML(h.Database)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		tasks, err := db.GetTasks(h.Database)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		caser := cases.Title(language.English)
		data := map[string]any{
			"Title": caser.String("hosts"),
			"Name":  name,
			"Path":  r.URL.Path,
			"Hosts": template.HTML(hostsHtml),
			"Tasks": tasks,
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case "tasks", "tasks/new":
		files := []string{
			"base.html",
			"tasks.html",
		}

		tmpl, err := template.ParseFS(ui.Content, files...)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tasksHtml, err := db.GetTasksAsHTML(h.Database)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		caser := cases.Title(language.English)
		data := map[string]any{
			"Title": caser.String("tasks"),
			"Name":  name,
			"Path":  r.URL.Path,
			"Tasks": tasksHtml,
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case "login":
		files := []string{
			"base.html",
			"login.html",
		}

		tmpl, err := template.ParseFS(ui.Content, files...)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]any{
			"Title": "Login",
			"Path":  r.URL.Path,
			"Error": r.URL.Query().Get("error"), // optional: pass error via query
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	default:
		if strings.HasPrefix(path, "tasks/edit/") {
			id := strings.TrimPrefix(path, "tasks/edit/")

			files := []string{
				"base.html",
				"tasks_edit.html",
			}

			tmpl, err := template.ParseFS(ui.Content, files...)
			if err != nil {
				if os.IsNotExist(err) {
					http.NotFound(w, r)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			task, err := db.GetTaskByID(id, h.Database)
			if err != nil {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}

			caser := cases.Title(language.English)
			data := map[string]any{
				"Title": caser.String("edit task"),
				"Name":  name,
				"Path":  r.URL.Path,
				"Task":  task,
			}

			if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		} else if strings.HasPrefix(path, "hosts/edit/") {
			id := strings.TrimPrefix(path, "hosts/edit/")

			files := []string{
				"base.html",
				"hosts_edit.html",
			}

			tmpl, err := template.ParseFS(ui.Content, files...)
			if err != nil {
				if os.IsNotExist(err) {
					http.NotFound(w, r)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			host, err := db.GetHostByID(id, h.Database)
			if err != nil {
				http.Error(w, "Host not found", http.StatusNotFound)
				return
			}

			tasks, err := db.GetTasks(h.Database)
			if err != nil {
				http.Error(w, "Tasks not found", http.StatusNotFound)
				return
			}

			caser := cases.Title(language.English)
			data := map[string]any{
				"Title": caser.String("edit task"),
				"Name":  name,
				"Path":  r.URL.Path,
				"Host":  host,
				"Tasks": tasks,
			}

			if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		http.NotFound(w, r)
	}
}
