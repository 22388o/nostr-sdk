package sdk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type ProfileMetadata struct {
	pubkey string
	event  *nostr.Event

	Name        string `json:"name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	About       string `json:"about,omitempty"`
	Website     string `json:"website,omitempty"`
	Picture     string `json:"picture,omitempty"`
	Banner      string `json:"banner,omitempty"`
	NIP05       string `json:"nip05,omitempty"`
	LUD16       string `json:"lud16,omitempty"`
}

func (p ProfileMetadata) Npub() string {
	v, _ := nip19.EncodePublicKey(p.pubkey)
	return v
}

func (p ProfileMetadata) Nprofile(ctx context.Context, sys *System, nrelays int) string {
	v, _ := nip19.EncodeProfile(p.pubkey, sys.FetchOutboxRelays(ctx, p.pubkey))
	return v
}

func FetchProfileMetadata(ctx context.Context, pool *nostr.SimplePool, pubkey string, relays ...string) ProfileMetadata {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := pool.SubManyEose(ctx, relays, nostr.Filters{
		{
			Kinds:   []int{nostr.KindProfileMetadata},
			Authors: []string{pubkey},
			Limit:   1,
		},
	})

	for ie := range ch {
		if m, err := ParseMetadata(ie.Event); err == nil {
			m.pubkey = pubkey
			m.event = ie.Event
			return *m
		}
	}

	return ProfileMetadata{pubkey: pubkey}
}

func ParseMetadata(event *nostr.Event) (*ProfileMetadata, error) {
	if event.Kind != 0 {
		return nil, fmt.Errorf("event %s is kind %d, not 0", event.ID, event.Kind)
	}

	var meta ProfileMetadata
	if err := json.Unmarshal([]byte(event.Content), &meta); err != nil {
		cont := event.Content
		if len(cont) > 100 {
			cont = cont[0:99]
		}
		return nil, fmt.Errorf("failed to parse metadata (%s) from event %s: %w", cont, event.ID, err)
	}

	return &meta, nil
}
