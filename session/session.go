package session

import (
	"database/sql"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Session struct {
	CookieLifetime string
	CookiePersist  string
	CookieName     string
	CookieDomain   string
	CookieSecure   string
	SessionType    string
	DBPool         *sql.DB
	RedisPool      *redis.Pool
}

func (c *Session) InitSession() *scs.SessionManager {
	var persist, secure bool

	// how long should sessions last?
	minutes, err := strconv.Atoi(c.CookieLifetime)
	if err != nil {
		minutes = 60
	}

	// should cookies persist?
	persist, _ = strconv.ParseBool(c.CookiePersist)

	// must cookies be secure
	secure, _ = strconv.ParseBool(c.CookieSecure)

	// create session
	session := scs.New()
	session.Lifetime = time.Duration(minutes) * time.Minute
	session.Cookie.Domain = c.CookieDomain
	session.Cookie.Name = c.CookieName
	session.Cookie.Persist = persist
	session.Cookie.Secure = secure
	session.Cookie.SameSite = http.SameSiteLaxMode

	// which session store?
	switch strings.ToLower(c.SessionType) {
	case "redis":
		session.Store = redisstore.New(c.RedisPool)
	case "mysql", "mariadb":
		session.Store = mysqlstore.New(c.DBPool)
	case "postgres", "postgresql":
		session.Store = postgresstore.New(c.DBPool)
	default:
		// cookie
	}

	return session
}
