package server

import (
	"embed"
	"io/fs"
	"net/http"
	"time"
)

//go:embed web/*
var webFiles embed.FS

func staticHandler() http.Handler {
	sub, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}

const webSessionCookie = "haiwo_web_session"
const webSessionValue = "authenticated"

func (a *App) requireWebAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.cfg.WebPassword == "" || a.isWebAuthenticated(r) {
			next.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	})
}

func (a *App) isWebAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie(webSessionCookie)
	return err == nil && cookie.Value == webSessionValue
}

func (a *App) loginPage(w http.ResponseWriter, r *http.Request) {
	if a.cfg.WebPassword == "" || a.isWebAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(loginHTML))
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if r.FormValue("password") != a.cfg.WebPassword {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(loginHTMLWithError))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     webSessionCookie,
		Value:    webSessionValue,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     webSessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

const loginHTML = `<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Haiwo Login</title>
    <style>
      *{box-sizing:border-box}body{margin:0;min-height:100vh;display:grid;place-items:center;background:#eef3f8;color:#142033;font-family:Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}.card{width:min(420px,calc(100vw - 36px));background:#fff;border:1px solid #d8e2ec;border-radius:16px;box-shadow:0 22px 58px rgba(26,42,62,.14);padding:28px}.brand{display:flex;align-items:center;gap:12px;margin-bottom:22px}.mark{display:grid;place-items:center;width:44px;height:44px;border-radius:10px;background:#0f1b2a;color:#fff;font-weight:900}h1,p{margin:0}h1{font-size:24px}p{color:#657387;line-height:1.5}.copy{margin-bottom:22px}.langs{display:flex;gap:8px;margin:16px 0}.langs button{border:1px solid #d8e2ec;border-radius:8px;background:#fff;padding:7px 10px;cursor:pointer}.langs button.active{background:#176b87;color:#fff;border-color:#176b87}label{display:grid;gap:8px;margin-bottom:14px;color:#657387;font-size:13px;font-weight:750}input{width:100%;border:1px solid #d8e2ec;border-radius:10px;padding:12px;font:inherit}input:focus{outline:none;border-color:#176b87;box-shadow:0 0 0 4px rgba(23,107,135,.12)}button.submit{width:100%;border:0;border-radius:10px;background:#176b87;color:#fff;padding:12px 14px;font-weight:800;cursor:pointer}.error{padding:11px 12px;border-radius:10px;background:#fee2e2;color:#b42318;margin-bottom:14px;font-size:14px}
    </style>
  </head>
  <body>
    <form class="card" method="post" action="/login">
      <div class="brand"><div class="mark">H</div><div><h1>Haiwo</h1><p data-zh="交付控制台" data-en="Delivery Console">交付控制台</p></div></div>
      <div class="copy"><p data-zh="请输入密码进入 CI/CD 管理页面。" data-en="Enter the password to open the CI/CD console.">请输入密码进入 CI/CD 管理页面。</p></div>
      <div class="langs"><button type="button" class="active" data-lang="zh">中文</button><button type="button" data-lang="en">English</button></div>
      <label><span data-zh="密码" data-en="Password">密码</span><input name="password" type="password" required autofocus /></label>
      <button class="submit" type="submit" data-zh="进入控制台" data-en="Enter Console">进入控制台</button>
    </form>
    <script>const pick=()=>{const saved=localStorage.getItem("haiwo_lang");if(saved==="zh"||saved==="en")return saved;const lang=(navigator.language||navigator.userLanguage||"").toLowerCase();return lang.startsWith("zh")?"zh":"en"};const apply=lang=>{localStorage.setItem("haiwo_lang",lang);document.documentElement.lang=lang==="zh"?"zh-CN":"en";document.querySelectorAll("[data-lang]").forEach(x=>x.classList.toggle("active",x.dataset.lang===lang));document.querySelectorAll("[data-zh]").forEach(el=>{el.textContent=el.dataset[lang]})};document.querySelectorAll("[data-lang]").forEach(b=>b.onclick=()=>apply(b.dataset.lang));apply(pick());</script>
  </body>
</html>`

const loginHTMLWithError = `<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Haiwo Login</title>
    <style>
      *{box-sizing:border-box}body{margin:0;min-height:100vh;display:grid;place-items:center;background:#eef3f8;color:#142033;font-family:Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}.card{width:min(420px,calc(100vw - 36px));background:#fff;border:1px solid #d8e2ec;border-radius:16px;box-shadow:0 22px 58px rgba(26,42,62,.14);padding:28px}.brand{display:flex;align-items:center;gap:12px;margin-bottom:22px}.mark{display:grid;place-items:center;width:44px;height:44px;border-radius:10px;background:#0f1b2a;color:#fff;font-weight:900}h1,p{margin:0}h1{font-size:24px}p{color:#657387;line-height:1.5}.copy{margin-bottom:22px}.langs{display:flex;gap:8px;margin:16px 0}.langs button{border:1px solid #d8e2ec;border-radius:8px;background:#fff;padding:7px 10px;cursor:pointer}.langs button.active{background:#176b87;color:#fff;border-color:#176b87}label{display:grid;gap:8px;margin-bottom:14px;color:#657387;font-size:13px;font-weight:750}input{width:100%;border:1px solid #d8e2ec;border-radius:10px;padding:12px;font:inherit}input:focus{outline:none;border-color:#176b87;box-shadow:0 0 0 4px rgba(23,107,135,.12)}button.submit{width:100%;border:0;border-radius:10px;background:#176b87;color:#fff;padding:12px 14px;font-weight:800;cursor:pointer}.error{padding:11px 12px;border-radius:10px;background:#fee2e2;color:#b42318;margin-bottom:14px;font-size:14px}
    </style>
  </head>
  <body>
    <form class="card" method="post" action="/login">
      <div class="brand"><div class="mark">H</div><div><h1>Haiwo</h1><p data-zh="交付控制台" data-en="Delivery Console">交付控制台</p></div></div>
      <div class="copy"><p data-zh="请输入密码进入 CI/CD 管理页面。" data-en="Enter the password to open the CI/CD console.">请输入密码进入 CI/CD 管理页面。</p></div>
      <div class="error" data-zh="密码错误，请重试。" data-en="Wrong password. Try again.">密码错误，请重试。</div>
      <div class="langs"><button type="button" class="active" data-lang="zh">中文</button><button type="button" data-lang="en">English</button></div>
      <label><span data-zh="密码" data-en="Password">密码</span><input name="password" type="password" required autofocus /></label>
      <button class="submit" type="submit" data-zh="进入控制台" data-en="Enter Console">进入控制台</button>
    </form>
    <script>const pick=()=>{const saved=localStorage.getItem("haiwo_lang");if(saved==="zh"||saved==="en")return saved;const lang=(navigator.language||navigator.userLanguage||"").toLowerCase();return lang.startsWith("zh")?"zh":"en"};const apply=lang=>{localStorage.setItem("haiwo_lang",lang);document.documentElement.lang=lang==="zh"?"zh-CN":"en";document.querySelectorAll("[data-lang]").forEach(x=>x.classList.toggle("active",x.dataset.lang===lang));document.querySelectorAll("[data-zh]").forEach(el=>{el.textContent=el.dataset[lang]})};document.querySelectorAll("[data-lang]").forEach(b=>b.onclick=()=>apply(b.dataset.lang));apply(pick());</script>
  </body>
</html>`
