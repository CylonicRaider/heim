package mock

import (
	"euphoria.leet.nu/heim/proto"
	"euphoria.leet.nu/heim/proto/security"
	"euphoria.leet.nu/heim/proto/snowflake"
	"euphoria.io/scope"
)

type agentTracker struct {
	b *TestBackend
}

func (t *agentTracker) Register(ctx scope.Context, agent *proto.Agent) error {
	t.b.Lock()
	defer t.b.Unlock()

	if _, ok := t.b.agents[agent.IDString()]; ok {
		return proto.ErrAgentAlreadyExists
	}
	if t.b.agents == nil {
		t.b.agents = map[string]*proto.Agent{agent.IDString(): agent}
	} else {
		t.b.agents[agent.IDString()] = agent
	}
	return nil
}

func (t *agentTracker) Get(ctx scope.Context, agentID string) (*proto.Agent, error) {
	agent, ok := t.b.agents[agentID]
	if !ok {
		return nil, proto.ErrAgentNotFound
	}
	return agent, nil
}

func (t *agentTracker) SetClientKey(
	ctx scope.Context, agentID string, accessKey *security.ManagedKey,
	accountID snowflake.Snowflake, clientKey *security.ManagedKey) error {

	t.b.Lock()
	defer t.b.Unlock()

	agent, err := t.Get(ctx, agentID)
	if err != nil {
		return err
	}

	if err := agent.SetClientKey(accessKey, clientKey); err != nil {
		return err
	}

	agent.AccountID = accountID.String()
	return nil
}

func (t *agentTracker) ClearClientKey(ctx scope.Context, agentID string) error {
	t.b.Lock()
	defer t.b.Unlock()

	agent, err := t.Get(ctx, agentID)
	if err != nil {
		return err
	}

	agent.AccountID = ""
	return nil
}
