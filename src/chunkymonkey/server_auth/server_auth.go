package server_auth

import (
	"expvar"
	"http"
	"os"
	"time"
	"url"
)

var (
	expVarServerAuthSuccessCount *expvar.Int
	expVarServerAuthFailCount    *expvar.Int
	expVarServerAuthTimeNs       *expvar.Int
)

func init() {
	expVarServerAuthSuccessCount = expvar.NewInt("server-auth-success-count")
	expVarServerAuthFailCount = expvar.NewInt("server-auth-fail-count")
	expVarServerAuthTimeNs = expvar.NewInt("server-auth-time-ns")
}

// An IAuthenticator takes a sessionId and a user string and attempts to
// authenticate against a server. This interface allows for the use of a dummy
// authentication server for testing purposes.
type IAuthenticator interface {
	Authenticate(sessionId, user string) (ok bool, err os.Error)
}

// DummyAuth is a no-op authentication server, always returning the value of
// 'Result'.
type DummyAuth struct {
	Result bool
}

// Authenticate implements the IAuthenticator.Authenticate method
func (d *DummyAuth) Authenticate(sessionId, user string) (authenticated bool, err os.Error) {
	return d.Result, nil
}

// ServerAuth represents authentication against a server, particularly the
// main minecraft server at http://www.minecraft.net/game/checkserver.jsp.
type ServerAuth struct {
	serverId string
	baseUrl  url.URL
}

func NewServerAuth(serverId, baseUrlStr string) (s *ServerAuth, err os.Error) {
	baseUrl, err := url.Parse(baseUrlStr)
	if err != nil {
		return
	}
	s = &ServerAuth{
		serverId: serverId,
		baseUrl:  *baseUrl,
	}
	return
}

// Build a URL+query string based on a given server URL, serverId and user
// input
func (s *ServerAuth) buildQuery(sessionId, user string) (query string) {
	queryValues := url.Values{
		"serverId":  {s.serverId},
		"sessionId": {sessionId},
		"user":      {user},
	}

	queryUrl := s.baseUrl
	queryUrl.RawQuery = queryValues.Encode()

	return queryUrl.String()
}

// Authenticate implements the IAuthenticator.Authenticate method
func (s *ServerAuth) Authenticate(sessionId, user string) (authenticated bool, err os.Error) {
	before := time.Nanoseconds()
	defer func() {
		after := time.Nanoseconds()
		expVarServerAuthTimeNs.Add(after - before)
		if authenticated {
			expVarServerAuthSuccessCount.Add(1)
		} else {
			expVarServerAuthFailCount.Add(1)
		}
	}()

	authenticated = false

	url_ := s.buildQuery(sessionId, user)

	response, err := http.Get(url_)
	if err != nil {
		return
	}

	if response.StatusCode == 200 {
		// We only need to read up to 3 bytes for "YES" or "NO"
		buf := make([]byte, 3)
		bufferPos := 0
		var numBytesRead int

		for err == nil && bufferPos < 3 {
			numBytesRead, err = response.Body.Read(buf[bufferPos:])
			if err != nil && err != os.EOF {
				return
			}
			bufferPos += numBytesRead
		}

		result := string(buf[0:bufferPos])
		authenticated = (result == "YES")
	} else {
		return
	}

	return
}
