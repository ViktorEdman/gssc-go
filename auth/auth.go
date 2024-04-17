package auth

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func NewAuthorizer() *Authorizer {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env present")
	}
	CLIENT_KEY := os.Getenv("CLIENT_KEY")
	CLIENT_SECRET := os.Getenv("CLIENT_SECRET")
	CALLBACK_URI := os.Getenv("CALLBACK_URI")
	if CLIENT_KEY == "" || CLIENT_SECRET == "" || CALLBACK_URI == "" {
		log.Fatal("Missing either CLIENT_KEY, CLIENT_SECRET or CALLBACK_URI environment variables")
	}
	goth.UseProviders(
		google.New(
			CLIENT_KEY,
			CLIENT_SECRET,
			CALLBACK_URI,
		),
	)
	gothicStore := sessions.NewCookieStore(securecookie.GenerateRandomKey(64))

	gothicStore.MaxAge(30 * 86400)
	gothicStore.Options.Path = "/"
	gothicStore.Options.HttpOnly = true // HttpOnly should always be enabled
	gothicStore.Options.Secure = true

	gothic.Store = gothicStore
	gothic.GetProviderName = func(r *http.Request) (string, error) {
		return r.PathValue("provider"), nil
	}

	sessionStore := sessions.NewCookieStore(securecookie.GenerateRandomKey(64))
	sessionStore.MaxAge(24 * 3600)
	sessionStore.Options.HttpOnly = true
	sessionStore.Options.Path = "/"
	sessionStore.Options.Secure = true

	authorizer := &Authorizer{
		SessionStore: sessionStore,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/{provider}", authorizer.signInHandler)
	mux.HandleFunc("/auth/{provider}/callback", authorizer.callbackHandler)
	mux.HandleFunc("/auth/{provider}/signout", authorizer.signOutHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world")
	})

	authorizer.Mux = mux
	return authorizer
}

type Authorizer struct {
	Mux          *http.ServeMux
	SessionStore *sessions.CookieStore
}

func getAuthorizedUsers() ([]string, error) {
	file, err := os.Open("authorized_users.txt")
	if err != nil {
		file, _ = os.Create("authorized_users.txt")
		fmt.Println("Couldn't find any authorized users, please add them to authorized_users.txt")
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func (a *Authorizer) signInHandler(w http.ResponseWriter, r *http.Request) {
	userSession, _ := a.SessionStore.Get(r, "user-session")
	if userSession.Values["Authorized"] != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		userSession.Values["Email"] = gothUser.Email
		userSession.Values["Authorized"] = true
		userSession.Values["Provider"] = gothUser.Provider
		userSession.Save(r, w)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func (a *Authorizer) callbackHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.SessionStore.Get(r, "user-session")
	if session.Values["Authorized"] == true {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	authorizedUsers, err := getAuthorizedUsers()
	if len(authorizedUsers) == 0 {
		fmt.Println("Couldn't find any authorized users, please add them to authorized_users.txt")
	}
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if !slices.Contains(authorizedUsers, user.Email) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	session.Values["Provider"] = user.Provider
	session.Values["Email"] = user.Email
	session.Values["Authorized"] = true
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (a *Authorizer) signOutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signing out")
	err := gothic.Logout(w, r)
	if err != nil {
		fmt.Println(err)
	}
	session, _ := a.SessionStore.Get(r, "user-session")
	session.Values["Authorized"] = false
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		fmt.Println(err)
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

type ContextKey string

var AuthorizedKey = ContextKey("Authorized")

func IsAuthorized(r *http.Request) bool {
	val, ok := r.Context().Value(AuthorizedKey).(bool)
	if ok {
		return val
	} else {
		return false
	}
}

func (a *Authorizer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := a.SessionStore.Get(r, "user-session")
		if session.Values["Authorized"] == true {
			log.Println(r.Method, r.URL, "Middleware: Authorized")
			ctx := context.WithValue(r.Context(), AuthorizedKey, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			log.Println(r.Method, r.URL, "Middleware: Not Authorized")
			ctx := context.WithValue(r.Context(), AuthorizedKey, false)
			next.ServeHTTP(w, r.WithContext(ctx))

		}
	})
}
