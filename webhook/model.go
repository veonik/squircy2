package webhook

import (
	"encoding/json"
	"sort"

	"github.com/veonik/squircy2/data"
)

type Webhook struct {
	ID              int
	Title           string
	Key             string
	SignatureHeader string // header containing signature
	Enabled         bool
}

type WebhookRepository struct {
	database *data.DB
}

func NewWebhookRepository(database *data.DB) WebhookRepository {
	return WebhookRepository{database}
}

func hydrateWebhook(rawWebhook map[string]interface{}) *Webhook {
	webhook := &Webhook{}

	webhook.Title = rawWebhook["Title"].(string)
	webhook.Key = rawWebhook["Key"].(string)
	webhook.SignatureHeader = rawWebhook["SignatureHeader"].(string)
	webhook.Enabled = rawWebhook["Enabled"].(bool)

	return webhook
}

func flattenWebhook(webhook *Webhook) map[string]interface{} {
	rawWebhook := make(map[string]interface{})

	rawWebhook["Title"] = webhook.Title
	rawWebhook["Key"] = webhook.Key
	rawWebhook["SignatureHeader"] = webhook.SignatureHeader
	rawWebhook["Enabled"] = webhook.Enabled

	return rawWebhook
}

type webhookSlice []*Webhook

func (s webhookSlice) Len() int {
	return len(s)
}

func (s webhookSlice) Less(i, j int) bool {
	return s[i].Title < s[j].Title
}

func (s webhookSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (repo *WebhookRepository) FetchAll() []*Webhook {
	col := repo.database.Use("Webhooks")
	webhooks := make([]*Webhook, 0)
	col.ForEachDoc(func(id int, doc []byte) (moveOn bool) {
		moveOn = true

		val := make(map[string]interface{}, 0)

		json.Unmarshal(doc, &val)
		webhook := hydrateWebhook(val)
		webhook.ID = id

		webhooks = append(webhooks, webhook)

		return
	})

	sort.Sort(webhookSlice(webhooks))

	return webhooks
}

func (repo *WebhookRepository) Fetch(id int) *Webhook {
	col := repo.database.Use("Webhooks")

	rawWebhook, err := col.Read(id)
	if err != nil {
		panic(err)
	}
	webhook := hydrateWebhook(rawWebhook)
	webhook.ID = id

	return webhook
}

func (repo *WebhookRepository) Save(webhook *Webhook) {
	col := repo.database.Use("Webhooks")
	data := flattenWebhook(webhook)

	if webhook.ID <= 0 {
		id, _ := col.Insert(data)
		webhook.ID = id

	} else {
		col.Update(webhook.ID, data)
	}
}

func (repo *WebhookRepository) Delete(id int) {
	col := repo.database.Use("Webhooks")
	col.Delete(id)
}
