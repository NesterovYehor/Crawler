package politeness

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/NesterovYehor/Crawler/internal/config"
	"github.com/NesterovYehor/Crawler/internal/storage"
	"github.com/temoto/robotstxt"
)

type PolitenessManager struct {
	st      storage.Interface
	scripts map[string]string
}
type RateLimitResult struct {
	Allowed bool
	Rules   string
}

func NewPM(st storage.Interface, cfg *config.Scripts) *PolitenessManager {
	s := map[string]string{
		"access": cfg.Access,
		"update": cfg.Update,
	}
	return &PolitenessManager{
		st:      st,
		scripts: s,
	}
}

func (p *PolitenessManager) GetRules(key string, ctx context.Context) (*RateLimitResult, error) {
	reply, err := p.st.RunScript(key, p.scripts["access"], ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %v", err)
	}
	res, err := parseScriptResult(reply)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *PolitenessManager) SaveRules(ctx context.Context, host string, rawRules string) error {
	rules, err := robotstxt.FromString(rawRules)
	if err != nil {
		return err
	}

	group := rules.FindGroup("MyCrawler")
	if group == nil {
		group = rules.FindGroup("*")
	}
	values := map[string]any{}

	blockDelaySeconds := 1
	if group.CrawlDelay != 0 {
		blockDelaySeconds = int(math.Ceil(group.CrawlDelay.Seconds() * 1000))
		if blockDelaySeconds == 0 {
			blockDelaySeconds = 1
		}
	}
	values["delay"] = blockDelaySeconds

	values["tokens_num"] = 0
	values["max_tokens_num"] = 0
	values["refill_time"] = time.Now().Unix() - 1
	values["rules"] = rawRules

	err = p.st.SaveToCache(ctx, host, values)
	if err != nil {
		return fmt.Errorf("failed to store new rules: %v", err)
	}
	return nil
}

func (p *PolitenessManager) UpdateHostLimit(key string, ctx context.Context) error {
	_, err := p.st.RunScript(key, p.scripts["update"], ctx)
	if err != nil {
		return fmt.Errorf("failed update host limit: %v", err)
	}
	return nil
}

func parseScriptResult(result any) (*RateLimitResult, error) {
	resSlice, ok := result.([]any)
	if !ok || len(resSlice) != 2 {
		return &RateLimitResult{Allowed: false}, nil
	}

	var rules string
	if resSlice[0] != nil {
		rules = resSlice[0].(string)
	}

	allowed := false
	if allowedVal, ok := resSlice[1].(int64); ok {
		allowed = allowedVal == 1
	}

	return &RateLimitResult{
		Allowed: allowed,
		Rules:   rules,
	}, nil
}
