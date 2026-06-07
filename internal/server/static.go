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
      *{box-sizing:border-box}body{margin:0;min-height:100vh;display:grid;place-items:center;background:#f5f5f5;color:#242424;font-family:"Segoe UI",SegoeUI,-apple-system,BlinkMacSystemFont,Roboto,"Helvetica Neue",sans-serif}.card{width:min(420px,calc(100vw - 36px));background:#fff;border:1px solid #e1dfdd;border-radius:2px;box-shadow:0 3.2px 7.2px rgba(0,0,0,.13),0 .6px 1.8px rgba(0,0,0,.11);padding:28px}.brand{display:grid;gap:12px;margin-bottom:22px}.logo{width:190px;height:62px;object-fit:contain;object-position:left center;border:1px solid #e1dfdd;border-radius:2px;background:#fff;padding:8px 10px}h1,p{margin:0}h1{font-size:24px;font-weight:600}p{color:#616161;line-height:1.5}.copy{margin-bottom:20px}.langs{display:flex;gap:6px;margin:16px 0}.langs button{border:1px solid #c8c6c4;border-radius:2px;background:#fff;padding:7px 10px;cursor:pointer}.langs button:hover{background:#f3f2f1}.langs button.active{background:#0078d4;color:#fff;border-color:#0078d4}label{display:grid;gap:7px;margin-bottom:14px;color:#616161;font-size:13px;font-weight:600}input{width:100%;border:1px solid #c8c6c4;border-radius:2px;padding:10px;font:inherit}input:hover{border-color:#605e5c}input:focus{outline:none;border-color:#0078d4;box-shadow:inset 0 0 0 1px #0078d4}button.submit{width:100%;border:1px solid #0078d4;border-radius:2px;background:#0078d4;color:#fff;padding:10px 14px;font-weight:600;cursor:pointer}button.submit:hover{background:#106ebe}.error{padding:10px 12px;border-radius:2px;background:#fde7e9;color:#d13438;margin-bottom:14px;font-size:14px}
    </style>
  </head>
  <body>
    <form class="card" method="post" action="/login">
      <div class="brand"><img class="logo" src="/logo.png" alt="Haiwo" /><div><p data-zh="交付控制台" data-en="Delivery Console" data-ja="デリバリーコンソール">交付控制台</p></div></div>
      <div class="copy"><p data-zh="请输入密码进入 CI/CD 管理页面。" data-en="Enter the password to open the CI/CD console." data-ja="CI/CD コンソールを開くにはパスワードを入力してください。">请输入密码进入 CI/CD 管理页面。</p></div>
      <div class="langs"><button type="button" class="active" data-lang="zh">中文</button><button type="button" data-lang="en">English</button><button type="button" data-lang="ja">日本語</button></div>
      <label><span data-zh="密码" data-en="Password" data-ja="パスワード">密码</span><input name="password" type="password" required autofocus /></label>
      <button class="submit" type="submit" data-zh="进入控制台" data-en="Enter Console" data-ja="コンソールを開く">进入控制台</button>
    </form>
    <script>const pick=()=>{const saved=localStorage.getItem("haiwo_lang");if(saved==="zh"||saved==="en"||saved==="ja")return saved;const lang=(navigator.language||navigator.userLanguage||"").toLowerCase();if(lang.startsWith("zh"))return"zh";if(lang.startsWith("ja"))return"ja";return"en"};const apply=lang=>{localStorage.setItem("haiwo_lang",lang);document.documentElement.lang=lang==="zh"?"zh-CN":lang==="ja"?"ja":"en";document.querySelectorAll("[data-lang]").forEach(x=>x.classList.toggle("active",x.dataset.lang===lang));document.querySelectorAll("[data-zh]").forEach(el=>{el.textContent=el.dataset[lang]})};document.querySelectorAll("[data-lang]").forEach(b=>b.onclick=()=>apply(b.dataset.lang));apply(pick());</script>
  </body>
</html>`

const loginHTMLWithError = `<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Haiwo Login</title>
    <style>
      *{box-sizing:border-box}body{margin:0;min-height:100vh;display:grid;place-items:center;background:#f5f5f5;color:#242424;font-family:"Segoe UI",SegoeUI,-apple-system,BlinkMacSystemFont,Roboto,"Helvetica Neue",sans-serif}.card{width:min(420px,calc(100vw - 36px));background:#fff;border:1px solid #e1dfdd;border-radius:2px;box-shadow:0 3.2px 7.2px rgba(0,0,0,.13),0 .6px 1.8px rgba(0,0,0,.11);padding:28px}.brand{display:grid;gap:12px;margin-bottom:22px}.logo{width:190px;height:62px;object-fit:contain;object-position:left center;border:1px solid #e1dfdd;border-radius:2px;background:#fff;padding:8px 10px}h1,p{margin:0}h1{font-size:24px;font-weight:600}p{color:#616161;line-height:1.5}.copy{margin-bottom:20px}.langs{display:flex;gap:6px;margin:16px 0}.langs button{border:1px solid #c8c6c4;border-radius:2px;background:#fff;padding:7px 10px;cursor:pointer}.langs button:hover{background:#f3f2f1}.langs button.active{background:#0078d4;color:#fff;border-color:#0078d4}label{display:grid;gap:7px;margin-bottom:14px;color:#616161;font-size:13px;font-weight:600}input{width:100%;border:1px solid #c8c6c4;border-radius:2px;padding:10px;font:inherit}input:hover{border-color:#605e5c}input:focus{outline:none;border-color:#0078d4;box-shadow:inset 0 0 0 1px #0078d4}button.submit{width:100%;border:1px solid #0078d4;border-radius:2px;background:#0078d4;color:#fff;padding:10px 14px;font-weight:600;cursor:pointer}button.submit:hover{background:#106ebe}.error{padding:10px 12px;border-radius:2px;background:#fde7e9;color:#d13438;margin-bottom:14px;font-size:14px}
    </style>
  </head>
  <body>
    <form class="card" method="post" action="/login">
      <div class="brand"><img class="logo" src="/logo.png" alt="Haiwo" /><div><p data-zh="交付控制台" data-en="Delivery Console" data-ja="デリバリーコンソール">交付控制台</p></div></div>
      <div class="copy"><p data-zh="请输入密码进入 CI/CD 管理页面。" data-en="Enter the password to open the CI/CD console." data-ja="CI/CD コンソールを開くにはパスワードを入力してください。">请输入密码进入 CI/CD 管理页面。</p></div>
      <div class="error" data-zh="密码错误，请重试。" data-en="Wrong password. Try again." data-ja="パスワードが違います。もう一度お試しください。">密码错误，请重试。</div>
      <div class="langs"><button type="button" class="active" data-lang="zh">中文</button><button type="button" data-lang="en">English</button><button type="button" data-lang="ja">日本語</button></div>
      <label><span data-zh="密码" data-en="Password" data-ja="パスワード">密码</span><input name="password" type="password" required autofocus /></label>
      <button class="submit" type="submit" data-zh="进入控制台" data-en="Enter Console" data-ja="コンソールを開く">进入控制台</button>
    </form>
    <script>const pick=()=>{const saved=localStorage.getItem("haiwo_lang");if(saved==="zh"||saved==="en"||saved==="ja")return saved;const lang=(navigator.language||navigator.userLanguage||"").toLowerCase();if(lang.startsWith("zh"))return"zh";if(lang.startsWith("ja"))return"ja";return"en"};const apply=lang=>{localStorage.setItem("haiwo_lang",lang);document.documentElement.lang=lang==="zh"?"zh-CN":lang==="ja"?"ja":"en";document.querySelectorAll("[data-lang]").forEach(x=>x.classList.toggle("active",x.dataset.lang===lang));document.querySelectorAll("[data-zh]").forEach(el=>{el.textContent=el.dataset[lang]})};document.querySelectorAll("[data-lang]").forEach(b=>b.onclick=()=>apply(b.dataset.lang));apply(pick());</script>
  </body>
</html>`
