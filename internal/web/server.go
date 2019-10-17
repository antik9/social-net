package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antik9/social-net/internal/config"
	"github.com/antik9/social-net/internal/errors"
	"github.com/antik9/social-net/pkg/models"
	"github.com/eknkc/amber"
	"github.com/gorilla/mux"
)

func renderTemplate(path string, data interface{}, w http.ResponseWriter) error {
	compiler := amber.New()
	err := compiler.ParseFile(path)
	if err != nil {
		return errors.New(projecterrors.UnknownTemplateError)
	}
	tpl, err := compiler.Compile()
	if err != nil {
		return errors.New(projecterrors.UnknownTemplateError)
	}
	tpl.Execute(w, data)
	return nil
}

func authenticateUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		renderTemplate("internal/web/templates/login.amber", nil, w)
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		if sess := models.NewSession(r.FormValue("email"), r.FormValue("password")); sess != nil {
			expiration := time.Now().Add(365 * 24 * time.Hour)
			cookie := http.Cookie{Name: "sn-session", Value: sess.Value, Expires: expiration}
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/mypage", http.StatusFound)
		} else {
			renderTemplate("internal/web/templates/login.amber", nil, w)
		}
	default:
		fmt.Fprint(w, "Sorry, only GET and POST methods are supported.")
	}
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		renderTemplate("internal/web/templates/registration.amber", nil, w)
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		if r.FormValue("password") == r.FormValue("password2") && len(r.FormValue("password")) > 0 {
			if age, err := strconv.Atoi(r.FormValue("age")); err == nil {
				models.NewUser(
					r.FormValue("city"),
					r.FormValue("first_name"),
					r.FormValue("last_name"),
					r.FormValue("email"),
					r.FormValue("password"),
					age,
					strings.Split(r.FormValue("interests"), ","),
				)
			}
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	default:
		fmt.Fprint(w, "Sorry, only GET and POST methods are supported.")
	}
}

func selfUserPage(w http.ResponseWriter, r *http.Request) {
	if cookie, ok := r.Cookie("sn-session"); ok == nil {
		user := models.GetUserBySessionValue(cookie.Value)
		if user != nil {
			renderTemplate("internal/web/templates/userpage.amber", user, w)
		}
	}
}

func otherUserPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if id, err := strconv.Atoi(vars["id"]); err == nil {
		user := models.GetUserById(id)
		renderTemplate("internal/web/templates/otheruserpage.amber", user, w)
	}
}

func ServeForever() {
	router := mux.NewRouter()
	router.HandleFunc("/signup", registerUser)
	router.HandleFunc("/login", authenticateUser)
	router.HandleFunc("/mypage", selfUserPage)
	router.HandleFunc("/user/{id}", otherUserPage)

	log.Fatal(http.ListenAndServe(
		fmt.Sprintf(
			"%s:%s",
			config.Conf.Server.Host, config.Conf.Server.Port,
		), router,
	))
}
