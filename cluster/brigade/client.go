package brigade

import (
	"time"

	"github.com/sunshineOfficial/golib/goctx"
	"github.com/sunshineOfficial/golib/gohttp"
)

type Client struct {
	client  gohttp.Client
	baseURL string
}

func NewClient(client gohttp.Client, baseURL string) *Client {
	return &Client{
		client:  client,
		baseURL: baseURL,
	}
}

func (c *Client) GetBrigadeByID(ctx goctx.Context, id int) (Brigade, error) {
	// TODO: убрать мок
	return Brigade{
		ID:     id,
		Status: StatusOnTask,
		Inspectors: []Inspector{
			{
				ID:          1,
				Surname:     "Хрунин",
				Name:        "Дмитрий",
				Patronymic:  "Алексеевич",
				PhoneNumber: "+791234567352",
				Email:       "asdfsd@gmail.com",
				AssignedAt:  time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          2,
				Surname:     "Пресняков",
				Name:        "Артем",
				Patronymic:  "Дмитриевич",
				PhoneNumber: "+791234567353",
				Email:       "asdfыsd2@gmail.com",
				AssignedAt:  time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
