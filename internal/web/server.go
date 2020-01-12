package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eknkc/amber"
	"github.com/gorilla/mux"

	"github.com/antik9/social-net/internal/cache"
	"github.com/antik9/social-net/internal/config"
	"github.com/antik9/social-net/internal/errors"
	"github.com/antik9/social-net/internal/queue"
	"github.com/antik9/social-net/internal/ws"
	"github.com/antik9/social-net/pkg/models"
)

type ListOfUsers struct {
	Users []models.User
}

type ListOfMessages struct {
	Messages []models.Message `json:"messages"`
	User     *models.User     `json:"user"`
	Other    *models.User     `json:"other"`
}

type UserFeed struct {
	FeedMessages []models.FeedMessage
	User         *models.User
}

var (
	producer = queue.NewClient("producer")
)

func prepareTemplate(path string) (*template.Template, error) {
	compiler := amber.New()
	err := compiler.ParseFile(path)
	if err != nil {
		return nil, errors.New(projecterrors.UnknownTemplateError)
	}
	tpl, err := compiler.Compile()
	if err != nil {
		return nil, errors.New(projecterrors.UnknownTemplateError)
	}
	return tpl, nil
}

func renderTemplate(path string, data interface{}, w http.ResponseWriter) error {
	tpl, err := prepareTemplate(path)
	if err != nil {
		return err
	}
	tpl.Execute(w, data)
	return nil
}

func renderFeedWithCache(key, path string, data interface{}, w http.ResponseWriter) error {
	cachedPage := cache.RedisCache.GetFeedPage(key)
	if cachedPage != "" {
		w.Write([]byte(cachedPage))
		return nil
	}

	tpl, err := prepareTemplate(path)
	if err != nil {
		return err
	}
	buffer := bytes.NewBufferString("")
	tpl.Execute(buffer, data)
	cache.RedisCache.CacheFeedPage(key, buffer.String())

	tpl.Execute(w, data)
	return nil
}

func authenticateUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if user := getUserBySession(r); user != nil {
			http.Redirect(w, r, "mypage", http.StatusFound)
			return
		}
		renderTemplate("internal/web/templates/login.amber", nil, w)
	case http.MethodPost:
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

func indexPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate("internal/web/templates/base.amber", nil, w)
}

func logoutUser(w http.ResponseWriter, r *http.Request) {
	expiration := time.Now().Add(1 * time.Hour)
	cookie := http.Cookie{Name: "sn-session", Value: "", Expires: expiration}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if user := getUserBySession(r); user != nil {
			http.Redirect(w, r, "mypage", http.StatusFound)
			return
		}
		renderTemplate("internal/web/templates/registration.amber", nil, w)
	case http.MethodPost:
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
	if user := getUserBySession(r); user != nil {
		userFeed := UserFeed{
			User:         user,
			FeedMessages: user.ListOwnFeedLimitBy(10),
		}
		renderTemplate("internal/web/templates/userpage.amber", userFeed, w)
		return
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func feedMessages(w http.ResponseWriter, r *http.Request) {
	if user := getUserBySession(r); user != nil {
		userFeed := UserFeed{
			User:         user,
			FeedMessages: user.ListFeedLimitBy(10),
		}
		renderFeedWithCache(strconv.Itoa(user.Id), "internal/web/templates/feed.amber", userFeed, w)
		return
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func feedUserMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if user := getUserById(vars); user != nil {
		userFeed := UserFeed{
			User:         user,
			FeedMessages: user.ListOwnFeedLimitBy(100),
		}
		renderTemplate("internal/web/templates/feed.amber", userFeed, w)
		return
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func getUserBySession(r *http.Request) *models.User {
	if cookie, ok := r.Cookie("sn-session"); ok == nil {
		return models.GetUserBySessionValue(cookie.Value)
	}
	return nil
}

func getUserById(vars map[string]string) *models.User {
	if id, err := strconv.Atoi(vars["id"]); err == nil {
		return models.GetUserById(id)
	}
	return nil
}

func chatWith(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	req, err := http.NewRequest(
		http.MethodGet, fmt.Sprintf(
			"http://%s:%s/chat/%s",
			config.Conf.ChatServer.Host, config.Conf.ChatServer.Port, vars["id"],
		), nil,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sessCookie, _ := r.Cookie("sn-session")
	req.AddCookie(sessCookie)

	client := http.Client{Timeout: time.Second}
	response, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch response.StatusCode {
	case http.StatusOK:
		var listOfMessages ListOfMessages
		data, _ := ioutil.ReadAll(response.Body)
		_ = json.Unmarshal(data, &listOfMessages)
		renderTemplate(
			"internal/web/templates/messages.amber",
			listOfMessages,
			w,
		)
	case http.StatusInternalServerError:
		http.Error(w, "cannot get messages from the server", http.StatusInternalServerError)
	case http.StatusForbidden:
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	default:
		http.NotFound(w, r)
	}
}

func chatMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if user := getUserBySession(r); user != nil {
		if other := getUserById(vars); other != nil {
			messages := models.GetMessagesForUsers(user, other)
			listOfMessages := ListOfMessages{Messages: messages, User: user, Other: other}
			marshalled, err := json.Marshal(listOfMessages)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.Write(marshalled)
			return
		}
		http.NotFound(w, r)
		return
	}
	http.Error(w, "session does not exist", http.StatusForbidden)
}

func saveChatMessage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	requestData := struct {
		Message  string `json:"message"`
		UserId   int    `json:"userId"`
		FriendId int    `json:"friendId"`
	}{}
	data, _ := ioutil.ReadAll(r.Body)

	err = json.Unmarshal(data, &requestData)
	if err != nil {
		http.Error(w, "inproper data", http.StatusBadRequest)
		return
	}

	if requestData.Message != "" && requestData.UserId != 0 && requestData.FriendId != 0 {
		user := models.GetUserById(requestData.UserId)
		friend := models.GetUserById(requestData.FriendId)
		if user == nil || friend == nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		models.SaveMessage(requestData.Message, user, friend)
	}
	http.Error(w, "empty data", http.StatusBadRequest)
}

func saveFeedMessage(w http.ResponseWriter, r *http.Request) {
	if user := getUserBySession(r); user != nil {
		switch r.Method {
		case http.MethodPost:
			if err := r.ParseForm(); err == nil {
				if message := r.FormValue("message"); message != "" {
					id, err := user.CreateFeedMessage(message)
					if err == nil {
						producer.SendMessage(strconv.FormatInt(id, 10))
					}
				}
			}
			http.Redirect(w, r, "/mypage", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
		return
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func subscribeTo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if user := getUserBySession(r); user != nil {
		if other := getUserById(vars); other != nil {
			if r.Method == http.MethodPost {
				user.SubscribeTo(other)
				http.Redirect(w, r, "/user/"+vars["id"], http.StatusFound)
			}
			return
		}
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func otherUserPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if user := getUserById(vars); user != nil {
		userFeed := UserFeed{
			User:         user,
			FeedMessages: user.ListOwnFeedLimitBy(10),
		}
		renderTemplate("internal/web/templates/otheruserpage.amber", userFeed, w)
		return
	}
	http.NotFound(w, r)
}

func searchUserPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil || len(r.FormValue("q")) < 1 {
		users := models.GetUsersLimitBy(100)
		renderTemplate("internal/web/templates/search.amber", ListOfUsers{Users: users}, w)
		return
	}

	users := models.GetUsersByNamePrefix(r.FormValue("q"), 100)
	renderTemplate("internal/web/templates/search.amber", ListOfUsers{Users: users}, w)
}

func ServeForever() {
	router := mux.NewRouter()
	hubs := ws.NewHubs()

	router.HandleFunc("/login", authenticateUser)
	router.HandleFunc("/logout", logoutUser)
	router.HandleFunc("/mypage", selfUserPage)
	router.HandleFunc("/search", searchUserPage)
	router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("internal/web/static"))),
	)
	router.HandleFunc("/signup", registerUser)
	router.HandleFunc("/chat/{id}", chatWith)
	router.HandleFunc("/new_feed/", saveFeedMessage)
	router.HandleFunc("/feed", feedMessages)
	router.HandleFunc("/feed/{id}", feedUserMessages)
	router.HandleFunc("/subscribe/{id}", subscribeTo)
	router.HandleFunc("/user/{id}", otherUserPage)
	router.HandleFunc("/ws/chat", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hubs, w, r)
	})
	router.HandleFunc("/", indexPage)

	log.Fatal(http.ListenAndServe(
		fmt.Sprintf(
			"%s:%s",
			config.Conf.Server.Host, config.Conf.Server.Port,
		), router,
	))
}

func ServeChatForever() {
	router := mux.NewRouter()
	router.HandleFunc("/chat/message", saveChatMessage)
	router.HandleFunc("/chat/{id}", chatMessages)

	log.Fatal(http.ListenAndServe(
		fmt.Sprintf(
			"%s:%s",
			config.Conf.ChatServer.Host, config.Conf.ChatServer.Port,
		), router,
	))
}
