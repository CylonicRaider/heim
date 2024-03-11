package backend

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/euphoria-io/scope"
	"github.com/gorilla/securecookie"

	"euphoria.leet.nu/heim/proto"
	"euphoria.leet.nu/heim/proto/logging"
	"euphoria.leet.nu/heim/proto/security"
	"euphoria.leet.nu/heim/proto/snowflake"
)

const (
	agentCookieName     = "b"
	agentCookieDuration = 365 * 24 * time.Hour
)

func newAgentCredentials(agent *proto.Agent, agentKey *security.ManagedKey) (*agentCredentials, error) {
	if agentKey.Encrypted() {
		return nil, security.ErrKeyMustBeDecrypted
	}

	ac := &agentCredentials{
		ID:  agent.IDString(),
		Key: agentKey.Plaintext,
	}
	return ac, nil
}

type agentCredentials struct {
	ID  string `json:"i"`
	Key []byte `json:"k"`
}

func (ac *agentCredentials) Cookie(sc *securecookie.SecureCookie) (*http.Cookie, error) {
	encoded, err := json.Marshal(ac)
	if err != nil {
		return nil, err
	}

	secured, err := sc.Encode(agentCookieName, encoded)
	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{
		Name:     agentCookieName,
		Value:    secured,
		Path:     "/",
		Expires:  time.Now().Add(agentCookieDuration),
		HttpOnly: true,
	}
	if !Config.Settings.SetInsecureCookies {
		cookie.Secure = true
	}
	return cookie, nil
}

func assignAgent(ctx scope.Context, s *Server, bot bool) (*proto.Agent, *http.Cookie, *security.ManagedKey, error) {
	agentKey := &security.ManagedKey{
		KeyType:   proto.AgentKeyType,
		Plaintext: make([]byte, proto.AgentKeyType.KeySize()),
	}
	_, err := rand.Read(agentKey.Plaintext)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("agent generation error: %s", err)
	}

	var agentID []byte
	if s.agentIDGenerator != nil {
		agentID, err = s.agentIDGenerator()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("agent generation error: %s", err)
		}
	}

	agent, err := proto.NewAgent(agentID, agentKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("agent generation error: %s", err)
	}

	agent.Bot = bot
	if err := s.b.AgentTracker().Register(ctx, agent); err != nil {
		return nil, nil, nil, fmt.Errorf("agent generation error: %s", err)
	}

	ac, err := newAgentCredentials(agent, agentKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("agent generation error: %s", err)
	}

	cookie, err := ac.Cookie(s.sc)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("agent generation error: %s", err)
	}

	return agent, cookie, agentKey, nil
}

func getAgent(
	ctx scope.Context, s *Server, r *http.Request) (
	*proto.Agent, *http.Cookie, *security.ManagedKey, error) {

	if err := r.ParseForm(); err != nil {
		return nil, nil, nil, err
	}
	bot := r.Form.Get("h") != "1"

	cookie, err := r.Cookie(agentCookieName)
	if err != nil {
		return assignAgent(ctx, s, bot)
	}

	encoded := []byte{}
	if err := s.sc.Decode(agentCookieName, cookie.Value, &encoded); err != nil {
		return assignAgent(ctx, s, bot)
	}

	ac := agentCredentials{}
	if err := json.Unmarshal(encoded, &ac); err != nil {
		return assignAgent(ctx, s, bot)
	}

	agent, err := s.b.AgentTracker().Get(ctx, ac.ID)
	if err != nil {
		return assignAgent(ctx, s, bot)
	}

	cookie, err = ac.Cookie(s.sc)
	if err != nil {
		return nil, nil, nil, err
	}

	agentKey := &security.ManagedKey{
		KeyType:   proto.AgentKeyType,
		Plaintext: ac.Key,
	}

	return agent, cookie, agentKey, nil
}

func getClient(ctx scope.Context, s *Server, r *http.Request) (*proto.Client, *http.Cookie, *security.ManagedKey, error) {
	agent, cookie, agentKey, err := getAgent(ctx, s, r)
	if err != nil {
		return nil, nil, nil, err
	}

	client := &proto.Client{
		Agent: agent,
		Authorization: proto.Authorization{
			ClientKey: agentKey,
		},
	}
	client.FromRequest(ctx, r)

	// Look up account associated with agent.
	var accountID snowflake.Snowflake
	if err := accountID.FromString(agent.AccountID); agent.AccountID != "" && err == nil {
		if err := client.AuthenticateWithAgent(ctx, s.b, agent, agentKey); err != nil {
			logging.Logger(ctx).Printf("agent auth failed: %s\n", err)
			switch err {
			case proto.ErrAccessDenied:
				// allow session to proceed, but agent will not be logged into account
				agent.AccountID = ""
			default:
				return nil, nil, nil, err
			}
		}
	}

	return client, cookie, agentKey, nil
}
