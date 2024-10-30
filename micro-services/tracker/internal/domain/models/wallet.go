package models

type WalletCommand struct {
    Action  string `json:"action"`  // "add" or "remove"
    Address string `json:"address"`
    ChatID  int64  `json:"chat_id"`
}

type Subscription struct {
    Address string
    ChatIDs map[int64]struct{}
}
