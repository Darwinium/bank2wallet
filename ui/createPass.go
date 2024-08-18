package main

import (
	"net/http"
	"net/url"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type CreatePass struct {
	app.Compo

	companyID   string
	companyName string
	iban        string
	bic         string
	address     string
	cashback    string
}

func (c *CreatePass) Render() app.UI {
	return app.Div().Body(
		app.H1().Body(
			app.Text("Create a new Apple Wallet Pass"),
		),
		app.P().Body(
			app.Input().
				Type("text").
				Value(c.companyID).
				Placeholder("Company ID").
				OnChange(c.ValueTo(&c.companyID)),

			app.Input().
				Type("text").
				Value(c.companyName).
				Placeholder("Company Name").
				OnChange(c.ValueTo(&c.companyName)),

			app.Input().
				Type("text").
				Value(c.iban).
				Placeholder("IBAN").
				OnChange(c.ValueTo(&c.iban)),

			app.Input().
				Type("text").
				Value(c.bic).
				Placeholder("BIC").
				OnChange(c.ValueTo(&c.bic)),

			app.Textarea().
				Text(c.address).
				Placeholder("Address").
				OnChange(c.ValueTo(&c.address)),

			app.Input().
				Type("text").
				Value(c.cashback).
				Placeholder("Cashback").
				OnChange(c.ValueTo(&c.cashback)),

			app.Button().
				Text("Create Pass").
				OnClick(c.createNewPass),
		),
	)
}

func (c *CreatePass) createNewPass(ctx app.Context, e app.Event) {
	go func() {
		// Replace with your API endpoint and adjust as necessary.
		res, err := http.PostForm("https://your-api-endpoint.com",
			url.Values{
				"companyID":   {c.companyID},
				"companyName": {c.companyName},
				"iban":        {c.iban},
				"bic":         {c.bic},
				"address":     {c.address},
				"cashback":    {c.cashback},
			})
		if err != nil {
			app.Log(err.Error())
		}

		app.Log("Pass created successfully %v, %v\n", res.Status, res.Body)
		// Optionally, update your UI or state here.
	}()
}
